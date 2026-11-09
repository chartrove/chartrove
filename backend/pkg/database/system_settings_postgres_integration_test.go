//go:build postgres_integration

package database

import (
	"io"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSystemSettingsPostgresAtomicSave(t *testing.T) {
	connection := disposablePostgres(t)
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "postgres")
	cfg.Set("database.location", connection.String())
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	repository, err := newTestRepository(t, cfg, logger, event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	systemSettingsAtomicControls(t, repository)
}
