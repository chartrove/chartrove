//go:build !cgo

package database

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedSQLiteBackupWithoutCGO(t *testing.T) {
	output := filepath.Join(t.TempDir(), "not-created")
	receipt, err := BackupEncryptedSQLite(context.Background(), filepath.Join(t.TempDir(), "missing.db"), output, "synthetic-key")
	require.ErrorContains(t, err, "CGO")
	require.Nil(t, receipt)
	_, err = os.Stat(output)
	require.True(t, os.IsNotExist(err))
}
