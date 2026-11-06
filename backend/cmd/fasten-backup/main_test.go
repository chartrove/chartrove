package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestBackupCommandUsesOnlyExplicitSyntheticInputs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0700))
	key := "synthetic-key-exact-bytes ?&'"
	source := filepath.Join(dir, "synthetic?#.db")
	u := url.URL{Scheme: "file", Path: source, RawQuery: url.Values{"_cipher": {"sqlcipher"}, "_legacy": {"3"}, "_hmac_use": {"off"}, "_kdf_iter": {"4000"}, "_legacy_page_size": {"1024"}, "_key": {key}, "_journal_mode": {"WAL"}}.Encode()}
	db, err := sql.Open("sqlite3", u.String())
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("CREATE TABLE synthetic_canary (value TEXT)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO synthetic_canary VALUES ('synthetic-native-data')")
	require.NoError(t, err)
	keyPath := filepath.Join(dir, "private-key")
	require.NoError(t, os.WriteFile(keyPath, []byte(key), 0600))
	tripwire := filepath.Join(dir, "never-open", "unrelated.db")
	t.Setenv("FASTEN_DATABASE_LOCATION", tripwire)
	t.Setenv("FASTEN_DATABASE_TYPE", "postgres")
	t.Setenv("FASTEN_AUTH_OIDC_ENABLED", "true")
	var out, diagnostic bytes.Buffer
	output := filepath.Join(dir, "new-backup")
	err = run(context.Background(), []string{"--source", source, "--output", output, "--key-file", keyPath}, &out, &diagnostic)
	require.NoError(t, err)
	var receipt database.SQLiteBackupReceipt
	require.NoError(t, json.Unmarshal(out.Bytes(), &receipt))
	require.Positive(t, receipt.Bytes)
	require.Len(t, receipt.SHA256, 64)
	require.Empty(t, diagnostic.String())
	require.NotContains(t, out.String(), key)
	require.NoFileExists(t, tripwire)
	// CLI failure never exposes the supplied key, creates a source, replaces an
	// existing output, or emits a success receipt.
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing flags", nil},
		{"existing output", []string{"--source", source, "--output", output, "--key-file", keyPath}},
		{"missing source", []string{"--source", filepath.Join(dir, "missing.db"), "--output", filepath.Join(dir, "missing-output"), "--key-file", keyPath}},
		{"unbounded timeout", []string{"--source", source, "--output", filepath.Join(dir, "timeout-output"), "--key-file", keyPath, "--timeout", "0s"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var o, d bytes.Buffer
			err := run(context.Background(), tc.args, &o, &d)
			require.Error(t, err)
			require.Empty(t, o.String())
			require.NotContains(t, err.Error()+d.String(), key)
		})
	}
	for _, name := range []string{"missing.db", "missing-output", "timeout-output"} {
		_, err := os.Lstat(filepath.Join(dir, name))
		require.True(t, os.IsNotExist(err))
	}
	before, err := os.ReadFile(filepath.Join(output, "database.sqlite"))
	require.NoError(t, err)
	_, err = runResult(context.Background(), []string{"--source", source, "--output", output, "--key-file", keyPath})
	require.Error(t, err)
	after, err := os.ReadFile(filepath.Join(output, "database.sqlite"))
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func runResult(ctx context.Context, args []string) (string, error) {
	var out, diagnostic bytes.Buffer
	err := run(ctx, args, &out, &diagnostic)
	return out.String(), err
}

func TestBackupCommandRejectsExposedOrAmbiguousKeyFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0700))
	private := filepath.Join(dir, "key")
	require.NoError(t, os.WriteFile(private, []byte("synthetic-key"), 0600))
	for _, tc := range []struct {
		name    string
		prepare func() string
	}{
		{"public", func() string {
			p := filepath.Join(dir, "public-key")
			require.NoError(t, os.WriteFile(p, []byte("synthetic-key"), 0644))
			return p
		}},
		{"symlink", func() string {
			p := filepath.Join(dir, "key-link")
			require.NoError(t, os.Symlink(private, p))
			return p
		}},
		{"empty", func() string {
			p := filepath.Join(dir, "empty-key")
			require.NoError(t, os.WriteFile(p, nil, 0600))
			return p
		}},
		{"oversize", func() string {
			p := filepath.Join(dir, "long-key")
			require.NoError(t, os.WriteFile(p, bytes.Repeat([]byte("x"), 4097), 0600))
			return p
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := readPrivateKey(tc.prepare())
			require.Error(t, err)
			require.Empty(t, value)
		})
	}
	key, err := readPrivateKey(private)
	require.NoError(t, err)
	require.Equal(t, "synthetic-key", key)
}
