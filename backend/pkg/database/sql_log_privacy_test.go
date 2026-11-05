package database

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRepositorySQLLoggingPrivacy(t *testing.T) {
	for _, level := range []string{"INFO", "WARN", "DEBUG"} {
		t.Run(level, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("database.type", "sqlite")
			cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-log-privacy.db"))
			cfg.Set("database.encryption.enabled", true)
			cfg.Set("database.encryption.key", "synthetic-log-privacy-key")
			cfg.Set("log.level", level)
			runRepositorySQLLoggingPrivacy(t, cfg)
		})
	}
}

// Exercise the actual constructor, migrations and legacy APIs. A deliberately
// short slow-query threshold covers WARN traces without sleeps. No parallel
// tests: logger.Default is restored before the next test may run.
func runRepositorySQLLoggingPrivacy(t *testing.T, cfg config.Interface) {
	t.Helper()
	var trace, application bytes.Buffer
	previous := logger.Default
	logger.Default = logger.New(log.New(&trace, "", 0), logger.Config{
		LogLevel: logger.Warn, SlowThreshold: time.Nanosecond,
	})
	t.Cleanup(func() { logger.Default = previous })
	appLog := logrus.New()
	appLog.SetOutput(&application)
	repo, err := newTestRepository(t, cfg, appLog, event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := repo.(*GormRepository)
	user := &models.User{Username: "synthetic-private-identity", Password: "synthetic-password", Email: "synthetic-private-email@example.invalid"}
	require.NoError(t, r.CreateUser(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
	ctx = context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, user.ID)
	source := &models.SourceCredential{Patient: "synthetic-private-patient", Display: "Synthetic private source", AccessToken: "synthetic-private-access-token", RefreshToken: "synthetic-private-refresh-token"}
	require.NoError(t, r.CreateSource(ctx, source))
	raw := json.RawMessage(`{"resourceType":"Observation","id":"synthetic-private-observation","status":"final","code":{"text":"synthetic-private-clinical-text"}}`)
	_, err = r.UpsertRawResource(ctx, source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: "synthetic-private-observation", ResourceRaw: raw})
	require.NoError(t, err)
	stored, err := r.GetResourceBySourceId(ctx, source.ID.String(), "synthetic-private-observation")
	require.NoError(t, err)
	require.JSONEq(t, string(raw), string(stored.ResourceRaw))
	credential, err := r.GetSource(ctx, source.ID.String())
	require.NoError(t, err)
	require.Equal(t, source.AccessToken, credential.AccessToken)

	// Error and missing-record traces are unsafe even without DEBUG. Errors
	// still reach the caller; suppressing automatic SQL output must not turn a
	// failed write into success or change rollback behavior.
	duplicate := models.User{Username: user.Username, Password: user.Password, Email: user.Email}
	err = r.GormClient.Create(&duplicate).Error
	require.Error(t, err)
	_, err = r.GetUserByUsername(ctx, "synthetic-private-missing-user")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	rollback := errors.New("synthetic-private-transaction-error")
	err = r.GormClient.Transaction(func(tx *gorm.DB) error {
		require.NoError(t, tx.Model(&models.User{}).Where("id = ?", user.ID).Update("email", "synthetic-private-rollback-email@example.invalid").Error)
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	current, err := r.GetUserByUsername(ctx, user.Username)
	require.NoError(t, err)
	require.Equal(t, user.Email, current.Email)

	// GORM's Scan recorder and an explicit Debug clone must also stay quiet.
	var scanned string
	require.NoError(t, r.GormClient.Debug().Raw("SELECT ?", "synthetic-private-scan-value").Scan(&scanned).Error)
	require.Equal(t, "synthetic-private-scan-value", scanned)
	for category, private := range map[string]string{
		"identity": user.Username, "email": user.Email, "password hash": user.Password,
		"patient": source.Patient, "access token": source.AccessToken, "refresh token": source.RefreshToken,
		"clinical text": "synthetic-private-clinical-text", "missing identity": "synthetic-private-missing-user",
		"scan value": scanned, "rollback value": "synthetic-private-rollback-email@example.invalid",
	} {
		t.Logf("synthetic %s present in automatic SQL output: %t", category, strings.Contains(trace.String(), private))
		require.False(t, strings.Contains(trace.String(), private), "automatic SQL output exposes %s", category)
		require.NotContains(t, application.String(), private, "application output exposes %s", category)
	}
	require.Zero(t, trace.Len(), "repository must not inherit the process-wide SQL logger")
	require.Contains(t, application.String(), "Successfully connected", "ordinary application lifecycle diagnostics remain available")
}
