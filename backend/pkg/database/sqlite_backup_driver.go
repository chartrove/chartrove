//go:build cgo

package database

import (
	"context"
	"database/sql"
	"errors"
	sqlite3 "github.com/mattn/go-sqlite3"
	"time"
)

const sqliteBackupEnabled = true

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
