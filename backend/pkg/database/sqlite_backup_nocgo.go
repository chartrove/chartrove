//go:build !cgo

package database

import (
	"context"
	"database/sql"
	"errors"
)

const sqliteBackupEnabled = false

func copySQLiteBackup(context.Context, *sql.DB, *sql.DB) error {
	return errors.New("encrypted SQLite backup requires a CGO-enabled build")
}
