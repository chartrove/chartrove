//go:build postgres_integration

package database

import (
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGraphSourceIdentityPostgres(t *testing.T) {
	dsn := disposablePostgres(t)
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "postgres")
	cfg.Set("database.location", dsn.String())
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	runGraphSourceIdentity(t, db.(*GormRepository))
}
