package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A small non-Chartrove schema also proves the backup never runs migrations.
func syntheticBackupSource(t *testing.T, journal string) (string, string, *sql.DB) {
	t.Helper()
	path, key := filepath.Join(privateBackupParent(t), "synthetic.db"), "synthetic-boundary-key"
	require.NoError(t, os.WriteFile(path, nil, 0600))
	db, err := openSQLiteBackupConnection(path, key, false)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("CREATE TABLE synthetic_canary (id INTEGER PRIMARY KEY, generation INTEGER, content BLOB)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO synthetic_canary VALUES (1,1,zeroblob(262144)), (2,1,zeroblob(262144))")
	require.NoError(t, err)
	var mode string
	if journal == "wal" {
		require.NoError(t, db.QueryRow("PRAGMA journal_mode=WAL").Scan(&mode))
	}
	return path, key, db
}

func TestEncryptedSQLiteBackupRefusesUnsafeInputs(t *testing.T) {
	source, key, _ := syntheticBackupSource(t, "delete")
	root := privateBackupParent(t)
	existing := filepath.Join(root, "existing")
	require.NoError(t, os.Mkdir(existing, 0700))
	sentinel := filepath.Join(existing, "database.sqlite")
	require.NoError(t, os.WriteFile(sentinel, []byte("retained-existing-backup"), 0600))
	linkedSource, linkedOutput := filepath.Join(root, "source-link"), filepath.Join(root, "output-link")
	require.NoError(t, os.Symlink(source, linkedSource))
	require.NoError(t, os.Symlink(existing, linkedOutput))
	public := filepath.Join(root, "public")
	require.NoError(t, os.Mkdir(public, 0755))
	plain := filepath.Join(root, "plaintext.db")
	db, err := sql.Open("sqlite3", plain)
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE synthetic_plaintext (value TEXT)")
	require.NoError(t, err)
	require.NoError(t, db.Close())
	for _, tc := range []struct{ name, source, output, key string }{
		{"existing output", source, existing, key},
		{"source symlink", linkedSource, filepath.Join(root, "source-link-output"), key},
		{"output symlink", source, linkedOutput, key},
		{"public parent", source, filepath.Join(public, "new-output"), key},
		{"missing source", filepath.Join(root, "missing.db"), filepath.Join(root, "missing-output"), key},
		{"source alias", source, source, key},
		{"plaintext source", plain, filepath.Join(root, "plain-output"), key},
		{"empty key", source, filepath.Join(root, "empty-output"), ""},
		{"wrong key", source, filepath.Join(root, "wrong-output"), "synthetic-wrong-key"},
		{"relative source", "synthetic.db", filepath.Join(root, "relative-output"), key},
		{"PostgreSQL is not SQLite", "postgres://synthetic.invalid/never-contact", filepath.Join(root, "pg-output"), key},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := BackupEncryptedSQLite(context.Background(), tc.source, tc.output, tc.key)
			require.Error(t, err)
			require.Nil(t, r)
			require.NotContains(t, err.Error(), key)
			if tc.output != existing && tc.output != linkedOutput && tc.output != source {
				_, statErr := os.Lstat(tc.output)
				require.True(t, os.IsNotExist(statErr), "failed backup must not leave a completed or partial artifact")
			}
		})
	}
	retained, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	require.Equal(t, "retained-existing-backup", string(retained))
	_, err = os.Lstat(filepath.Join(root, "missing.db"))
	require.True(t, os.IsNotExist(err))
}

func TestEncryptedSQLiteBackupCancellationAndReadOnlySource(t *testing.T) {
	source, key, db := syntheticBackupSource(t, "wal")
	// The backup must exclude uncommitted changes and must not commit/roll back
	// another connection's transaction or alter a non-application schema.
	tx, err := db.Begin()
	require.NoError(t, err)
	_, err = tx.Exec("UPDATE synthetic_canary SET generation=999")
	require.NoError(t, err)
	output := filepath.Join(privateBackupParent(t), "backup")
	_, err = BackupEncryptedSQLite(context.Background(), source, output, key)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	copy, err := openSQLiteBackupConnection(filepath.Join(output, "database.sqlite"), key, true)
	require.NoError(t, err)
	defer copy.Close()
	var min, max int
	require.NoError(t, copy.QueryRow("SELECT min(generation), max(generation) FROM synthetic_canary").Scan(&min, &max))
	require.Equal(t, 1, min)
	require.Equal(t, 1, max)
	var count int
	require.NoError(t, copy.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&count))
	require.Equal(t, 1, count, "no startup migrations")
	require.NoError(t, db.QueryRow("SELECT min(generation) FROM synthetic_canary").Scan(&min))
	require.Equal(t, 999, min)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	cancelOutput := filepath.Join(privateBackupParent(t), "cancelled")
	_, err = BackupEncryptedSQLite(cancelled, source, cancelOutput, key)
	require.ErrorIs(t, err, context.Canceled)
	_, err = os.Stat(cancelOutput)
	require.True(t, os.IsNotExist(err))
	// A real SQLite exclusive writer makes unlock/read wait. The deadline must
	// fail the operation and remove its new output, rather than publish success.
	blocked, blockedKey, writer := syntheticBackupSource(t, "delete")
	conn, err := writer.Conn(context.Background())
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.ExecContext(context.Background(), "BEGIN EXCLUSIVE")
	require.NoError(t, err)
	defer conn.ExecContext(context.Background(), "ROLLBACK")
	deadline, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	blockedOutput := filepath.Join(privateBackupParent(t), "timed-out")
	start := time.Now()
	_, err = BackupEncryptedSQLite(deadline, blocked, blockedOutput, blockedKey)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(start), 2*time.Second)
	_, err = os.Stat(blockedOutput)
	require.True(t, os.IsNotExist(err))
}

func privateBackupParent(t *testing.T) string {
	t.Helper()
	path := t.TempDir()
	require.NoError(t, os.Chmod(path, 0700))
	return path
}

func TestEncryptedSQLiteBackupConcurrentTransactionsStayAtomic(t *testing.T) {
	source, key, writer := syntheticBackupSource(t, "wal")
	first, done := make(chan struct{}), make(chan error, 1)
	go func() {
		for generation := 2; generation <= 33; generation++ {
			err := func() error {
				tx, err := writer.Begin()
				if err != nil {
					return err
				}
				defer tx.Rollback()
				if _, err = tx.Exec("UPDATE synthetic_canary SET generation=? WHERE id=1", generation); err != nil {
					return err
				}
				// Expose a real uncommitted mixed state while the reader runs.
				time.Sleep(time.Millisecond)
				if _, err = tx.Exec("UPDATE synthetic_canary SET generation=? WHERE id=2", generation); err != nil {
					return err
				}
				return tx.Commit()
			}()
			if generation == 2 {
				close(first)
			}
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	<-first
	output := filepath.Join(privateBackupParent(t), "backup")
	_, err := BackupEncryptedSQLite(context.Background(), source, output, key)
	require.NoError(t, <-done)
	require.NoError(t, err)
	copy, err := openSQLiteBackupConnection(filepath.Join(output, "database.sqlite"), key, true)
	require.NoError(t, err)
	defer copy.Close()
	var min, max, count, size int
	require.NoError(t, copy.QueryRow("SELECT min(generation), max(generation), count(*), sum(length(content)) FROM synthetic_canary").Scan(&min, &max, &count, &size))
	require.Equal(t, min, max, "a backup must never expose half of a committed pair")
	require.GreaterOrEqual(t, min, 2)
	require.LessOrEqual(t, max, 33)
	require.Equal(t, 2, count)
	require.Equal(t, 524288, size)
}
