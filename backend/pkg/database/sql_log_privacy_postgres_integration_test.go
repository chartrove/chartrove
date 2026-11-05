//go:build postgres_integration

package database

import (
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepositorySQLLoggingPrivacy(t *testing.T) {
	dsn := disposablePostgres(t)
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "postgres")
	cfg.Set("database.location", dsn.String())
	cfg.Set("log.level", "DEBUG")
	runRepositorySQLLoggingPrivacy(t, cfg)
}
