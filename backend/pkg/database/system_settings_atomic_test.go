package database

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSystemSettingsSavePropagatesFailureAndPreservesCredentialPair(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-settings.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-settings-storage-fixture-only")
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	repository, err := newTestRepository(t, cfg, logger, event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	systemSettingsAtomicControls(t, repository)
}

func systemSettingsAtomicControls(t *testing.T, repository DatabaseRepository) {
	t.Helper()
	concrete := repository.(*GormRepository)
	original := models.SystemSettings{InstallationID: "synthetic-original-id", InstallationSecret: "synthetic-original-secret"}
	candidate := models.SystemSettings{InstallationID: "synthetic-replacement-id", InstallationSecret: "synthetic-replacement-secret"}
	require.NoError(t, repository.SaveSystemSettings(t.Context(), &original))
	before, err := repository.LoadSystemSettings(t.Context())
	require.NoError(t, err)
	require.Equal(t, original, *before)
	retained := models.SystemSettingEntry{SettingKeyName: "synthetic_future_setting", SettingDataType: "string", SettingValueString: "synthetic-retained-value"}
	require.NoError(t, concrete.GormClient.Create(&retained).Error)

	fault := errors.New("synthetic system-settings write failure")
	const callback = "synthetic_second_system_setting_write_failure"
	writes := 0
	require.NoError(t, concrete.GormClient.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "system_settings" {
			writes++
			if writes == 2 {
				tx.AddError(fault)
			}
		}
	}))
	removed := false
	t.Cleanup(func() {
		if !removed {
			_ = concrete.GormClient.Callback().Update().Remove(callback)
		}
	})
	err = repository.SaveSystemSettings(t.Context(), &candidate)
	assert.ErrorIs(t, err, fault, "a failed write must not be reported as successful persistence")
	require.Equal(t, 2, writes, "fail after one actual credential-row update")
	require.NoError(t, concrete.GormClient.Callback().Update().Remove(callback))
	removed = true
	stored, err := repository.LoadSystemSettings(t.Context())
	require.NoError(t, err)
	assert.Equal(t, original, *stored, "ID and secret must not become a partial/mixed credential pair")

	require.NoError(t, repository.SaveSystemSettings(t.Context(), &candidate))
	stored, err = repository.LoadSystemSettings(t.Context())
	require.NoError(t, err)
	require.Equal(t, candidate, *stored)
	var stillRetained models.SystemSettingEntry
	require.NoError(t, concrete.GormClient.First(&stillRetained, retained.ID).Error)
	require.Equal(t, retained.SettingValueString, stillRetained.SettingValueString, "do not rewrite unknown/future settings as installation IDs")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.Error(t, repository.SaveSystemSettings(ctx, &original))
	stored, err = repository.LoadSystemSettings(t.Context())
	require.NoError(t, err)
	require.Equal(t, candidate, *stored)

	// Missing credential rows are storage damage, not successful persistence.
	deleted := concrete.GormClient.Where("setting_key_name = ?", "installation_secret").Delete(&models.SystemSettingEntry{})
	require.NoError(t, deleted.Error)
	require.EqualValues(t, 1, deleted.RowsAffected)
	err = repository.SaveSystemSettings(t.Context(), &original)
	assert.Error(t, err, "missing credential rows must fail closed before any surviving row is updated")
	stored, err = repository.LoadSystemSettings(t.Context())
	require.NoError(t, err)
	assert.Equal(t, candidate.InstallationID, stored.InstallationID)
	assert.Empty(t, stored.InstallationSecret)
}
