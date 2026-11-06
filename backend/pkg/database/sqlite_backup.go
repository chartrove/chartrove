package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type SQLiteBackupReceipt struct {
	Format    string    `json:"format"`
	Completed time.Time `json:"completed_at_utc"`
	SHA256    string    `json:"database_sha256"`
	Bytes     int64     `json:"database_bytes"`
}

// BackupEncryptedSQLite is a full-instance operator operation, never a patient
// or administrator HTTP API. It uses no application configuration, migrations,
// account recovery, provider connection or plaintext intermediate. The caller
// must separately preserve the key and installation configuration.
//
// output must be a NEW directory inside a private existing parent. Only the
// artifact and receipt are created there. Existing output is never overwritten.
func BackupEncryptedSQLite(ctx context.Context, source, output, key string) (_ *SQLiteBackupReceipt, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(source) || !filepath.IsAbs(output) || key == "" || len(key) > 4096 || !utf8.ValidString(key) || strings.ContainsRune(key, 0) {
		return nil, errors.New("backup requires absolute paths and a nonempty valid encryption key")
	}
	source, output = filepath.Clean(source), filepath.Clean(output)
	info, err := os.Lstat(source)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 16 {
		return nil, errors.New("backup source must be an existing regular encrypted SQLite file")
	}
	// Refuse plaintext/foreign inputs rather than silently converting formats.
	f, err := os.Open(source)
	if err != nil {
		return nil, errors.New("could not read backup source")
	}
	var header [16]byte
	_, readErr := io.ReadFull(f, header[:])
	closeErr := f.Close()
	if readErr != nil || closeErr != nil || string(header[:]) == "SQLite format 3\x00" {
		return nil, errors.New("backup source is not a supported encrypted SQLite file")
	}
	parent, err := os.Lstat(filepath.Dir(output))
	if err != nil || !parent.IsDir() || parent.Mode().Perm()&0077 != 0 {
		return nil, errors.New("backup output parent must be an existing private directory")
	}
	if os.Mkdir(output, 0700) != nil {
		return nil, errors.New("backup output must not already exist")
	}
	complete := false
	defer func() {
		if !complete {
			// Only known files in the directory exclusively created by this call.
			// Never recurse into or replace an existing installation/backup.
			for _, name := range []string{"database.sqlite", "database.sqlite-journal", "database.sqlite-wal", "database.sqlite-shm", "receipt.json"} {
				_ = os.Remove(filepath.Join(output, name))
			}
			_ = os.Remove(output)
		}
	}()
	backup := filepath.Join(output, "database.sqlite")
	f, err = os.OpenFile(backup, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.New("could not create private backup file")
	}
	if f.Close() != nil {
		return nil, errors.New("could not close private backup file")
	}
	src, err := openSQLiteBackupConnection(source, key, true)
	if err != nil {
		return nil, errors.New("could not open encrypted backup source")
	}
	defer src.Close()
	var schema int
	if src.QueryRowContext(ctx, "PRAGMA schema_version").Scan(&schema) != nil {
		return nil, backupFailure(ctx, "could not unlock encrypted backup source")
	}
	dst, err := openSQLiteBackupConnection(backup, key, false)
	if err != nil {
		return nil, errors.New("could not open encrypted backup destination")
	}
	defer dst.Close()
	if copySQLiteBackup(ctx, src, dst) != nil {
		return nil, backupFailure(ctx, "encrypted SQLite backup failed")
	}
	var journal string
	if dst.QueryRowContext(ctx, "PRAGMA journal_mode=DELETE").Scan(&journal) != nil || journal != "delete" {
		return nil, backupFailure(ctx, "could not finalize a standalone backup file")
	}
	if dst.Close() != nil {
		return nil, errors.New("could not close encrypted backup")
	}
	// Reopen the completed single file without its source/WAL connections. Do
	// not use NewRepository: recovery must not run startup migrations/jobs.
	check, err := openSQLiteBackupConnection(backup, key, true)
	if err != nil {
		return nil, errors.New("could not reopen encrypted backup")
	}
	var integrity string
	checkErr := check.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity)
	closeErr = check.Close()
	if checkErr != nil || closeErr != nil || integrity != "ok" {
		return nil, backupFailure(ctx, "encrypted backup integrity verification failed")
	}
	f, err = os.OpenFile(backup, os.O_RDWR, 0)
	if err != nil {
		return nil, errors.New("could not verify backup file")
	}
	defer f.Close()
	hash := sha256.New()
	n, err := io.Copy(hash, backupContextReader{ctx, f})
	if err != nil || f.Sync() != nil {
		return nil, backupFailure(ctx, "could not hash or sync encrypted backup")
	}
	if f.Close() != nil {
		return nil, errors.New("could not close verified backup")
	}
	receipt := &SQLiteBackupReceipt{Format: "chartrove-encrypted-sqlite-v1", Completed: time.Now().UTC(), SHA256: hex.EncodeToString(hash.Sum(nil)), Bytes: n}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, errors.New("could not encode backup receipt")
	}
	f, err = os.OpenFile(filepath.Join(output, "receipt.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.New("could not create backup receipt")
	}
	defer f.Close()
	_, err = f.Write(append(raw, '\n'))
	if err != nil || f.Sync() != nil {
		return nil, errors.New("could not sync backup receipt")
	}
	if f.Close() != nil {
		return nil, errors.New("could not close backup receipt")
	}
	for _, path := range []string{output, filepath.Dir(output)} {
		dir, err := os.Open(path)
		if err != nil {
			return nil, errors.New("could not open backup directory")
		}
		syncErr := dir.Sync()
		closeErr = dir.Close()
		if syncErr != nil || closeErr != nil {
			return nil, errors.New("could not sync backup directory")
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	complete = true
	return receipt, nil
}

// No DSN, key, SQL value or driver error is included in public error text.
func backupFailure(ctx context.Context, message string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New(message)
}

func openSQLiteBackupConnection(path, key string, readOnly bool) (*sql.DB, error) {
	q := url.Values{}
	for k, v := range sqliteCipherPragmas(key) {
		q.Set(k, v)
	}
	q.Set("_busy_timeout", "100")
	q.Set("cache", "private")
	if readOnly {
		q.Set("mode", "ro")
		q.Set("_query_only", "on")
	} else {
		q.Set("mode", "rw")
		q.Set("_journal_mode", "DELETE")
		q.Set("_synchronous", "FULL")
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: q.Encode()}
	db, err := sql.Open("sqlite3", u.String())
	if err == nil {
		db.SetMaxOpenConns(1)
	}
	return db, err
}

func copySQLiteBackup(ctx context.Context, src, dst *sql.DB) error {
	s, err := src.Conn(ctx)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := dst.Conn(ctx)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Raw(func(dest interface{}) error {
		return s.Raw(func(source interface{}) (err error) {
			b, err := dest.(*sqlite3.SQLiteConn).Backup("main", source.(*sqlite3.SQLiteConn), "main")
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, b.Finish()) }()
			for {
				if err := ctx.Err(); err != nil {
					return err
				}
				done, err := b.Step(128)
				if err != nil || done {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Millisecond):
				}
			}
		})
	})
}

type backupContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r backupContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
