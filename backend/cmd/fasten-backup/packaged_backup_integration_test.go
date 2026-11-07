//go:build packaged_backup_integration

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPackagedBackupBinaryRestoresSyntheticWAL(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("packaged backup qualification requires native macOS arm64")
	}
	binary := os.Getenv("CHARTROVE_PACKAGED_BACKUP")
	require.True(t, filepath.IsAbs(binary), "select the explicitly extracted package binary")
	info, err := os.Lstat(binary)
	require.NoError(t, err)
	require.True(t, info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0)
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0700))
	key := "synthetic-" + uuid.NewString() + " ?&'\n"
	wrongKey := "synthetic-wrong-" + uuid.NewString()
	dsn := func(path, key string, readOnly bool) string {
		options := url.Values{"_cipher": {"sqlcipher"}, "_legacy": {"3"}, "_hmac_use": {"off"}, "_kdf_iter": {"4000"}, "_legacy_page_size": {"1024"}, "_key": {key}}
		if readOnly {
			options.Set("mode", "ro")
		} else {
			options.Set("_journal_mode", "WAL")
		}
		return (&url.URL{Scheme: "file", Path: path, RawQuery: options.Encode()}).String()
	}
	source := filepath.Join(dir, "synthetic?#.db")
	db, err := sql.Open("sqlite3", dsn(source, key, false))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	_, err = db.Exec("CREATE TABLE synthetic_canary (value TEXT)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO synthetic_canary VALUES ('synthetic-base')")
	require.NoError(t, err)
	_, err = db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO synthetic_canary VALUES ('synthetic-wal')")
	require.NoError(t, err)
	wal, err := os.Stat(source + "-wal")
	require.NoError(t, err)
	require.Positive(t, wal.Size(), "the live source must have committed WAL state")
	keyFile := filepath.Join(dir, "private-key")
	wrongKeyFile := filepath.Join(dir, "private-wrong-key")
	require.NoError(t, os.WriteFile(keyFile, []byte(key), 0600))
	require.NoError(t, os.WriteFile(wrongKeyFile, []byte(wrongKey), 0600))
	tripwire := filepath.Join(dir, "never-open", "unrelated.db")
	t.Setenv("FASTEN_DATABASE_LOCATION", tripwire)
	t.Setenv("FASTEN_DATABASE_TYPE", "postgres")
	t.Setenv("FASTEN_AUTH_OIDC_ENABLED", "true")
	invoke := func(destination, selectedKey string) ([]byte, error) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, binary, "--source", source, "--output", destination, "--key-file", selectedKey, "--timeout", "10s").CombinedOutput()
		if strings.Contains(string(out), key) || strings.Contains(string(out), wrongKey) {
			t.Fatal("packaged backup diagnostics exposed a synthetic key")
		}
		return out, err
	}
	output := filepath.Join(dir, "new-backup")
	out, err := invoke(output, keyFile)
	require.NoError(t, err, "actual extracted executable must produce the synthetic backup")
	var receipt database.SQLiteBackupReceipt
	require.NoError(t, json.Unmarshal(out, &receipt))
	ciphertext, err := os.ReadFile(filepath.Join(output, "database.sqlite"))
	require.NoError(t, err)
	checksum := sha256.Sum256(ciphertext)
	require.Equal(t, hex.EncodeToString(checksum[:]), receipt.SHA256)
	require.EqualValues(t, len(ciphertext), receipt.Bytes)
	require.NotContains(t, string(ciphertext), "synthetic-wal")
	restored := filepath.Join(dir, "restored.sqlite")
	require.NoError(t, os.WriteFile(restored, ciphertext, 0600))
	for i := 0; i < 2; i++ {
		recovery, err := sql.Open("sqlite3", dsn(restored, key, true))
		require.NoError(t, err)
		var count int
		err = recovery.QueryRow("SELECT COUNT(*) FROM synthetic_canary WHERE value IN ('synthetic-base', 'synthetic-wal')").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count, "standalone restore must include the committed WAL row")
		var integrity string
		require.NoError(t, recovery.QueryRow("PRAGMA integrity_check").Scan(&integrity))
		require.Equal(t, "ok", integrity)
		require.NoError(t, recovery.Close())
	}
	wrong, err := sql.Open("sqlite3", dsn(restored, wrongKey, true))
	require.NoError(t, err)
	var count int
	require.Error(t, wrong.QueryRow("SELECT COUNT(*) FROM synthetic_canary").Scan(&count))
	require.NoError(t, wrong.Close())
	wrongOutput := filepath.Join(dir, "wrong-key-backup")
	_, err = invoke(wrongOutput, wrongKeyFile)
	var exit *exec.ExitError
	require.ErrorAs(t, err, &exit)
	require.Equal(t, 1, exit.ExitCode())
	require.NoDirExists(t, wrongOutput)
	require.NoFileExists(t, tripwire)
}
