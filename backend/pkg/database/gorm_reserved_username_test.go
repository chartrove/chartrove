package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestReservedNamesCannotCreateAccounts(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic.db"))
	cfg.Set("database.encryption.enabled", false)
	repo, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	for _, name := range []string{"api", "contact", "fasten", "chartrove", "help", "info", "login", "mail", "noreply", "postmaster", "root", "security", "support", "system", "webmaster", "www", " SyStEm "} {
		t.Run(name, func(t *testing.T) {
			user := &models.User{Username: name, Password: "synthetic-password"}
			err := repo.CreateUser(context.Background(), user)
			require.EqualError(t, err, "Username is reserved; choose another name")
			require.Equal(t, "synthetic-password", user.Password, "rejection precedes password hashing")
			require.Equal(t, uuid.Nil, user.ID)
		})
	}
	for _, model := range []interface{}{&models.User{}, &models.UserSettingEntry{}, &models.SourceCredential{}} {
		var count int64
		require.NoError(t, repo.(*GormRepository).GormClient.Model(model).Count(&count).Error)
		require.Zero(t, count, "rejection must not create users, settings or source credentials")
	}
}

func TestRegistrationAllowsMaintainerAndLiteralNames(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic.db"))
	cfg.Set("database.encryption.enabled", false)
	repo, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	for _, name := range []string{"admin", "administrator", "supporting-user", " Synthetic User Ω "} {
		t.Run(name, func(t *testing.T) {
			user := &models.User{Username: name, Password: "synthetic-password", Role: pkg.UserRoleUser}
			require.NoError(t, repo.CreateUser(context.Background(), user))
			stored, err := repo.GetUserByUsername(context.Background(), name)
			require.NoError(t, err)
			require.Equal(t, name, stored.Username)
			require.Equal(t, pkg.UserRoleUser, stored.Role)
			require.NoError(t, stored.CheckPassword("synthetic-password"))
		})
	}
}

func TestLegacyReservedAccountSurvivesEncryptedReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic-legacy.db")
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.location", path)
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-reserved-name-key")
	open := func() DatabaseRepository {
		repo, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		return repo
	}
	repo := open()
	// Direct insertion represents an account stored before the creation policy.
	legacy := &models.User{Username: "system", Role: pkg.UserRoleAdmin}
	require.NoError(t, legacy.HashPassword("synthetic-legacy-password"))
	require.NoError(t, repo.(*GormRepository).GormClient.Create(legacy).Error)
	require.NoError(t, repo.PopulateDefaultUserSettings(context.Background(), legacy.ID))
	source := &models.SourceCredential{UserID: legacy.ID}
	require.NoError(t, repo.(*GormRepository).GormClient.Create(source).Error)
	require.NoError(t, repo.Close())
	header, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Greater(t, len(header), 16)
	require.NotEqual(t, "SQLite format 3\x00", string(header[:16]))
	reopened := open()
	stored, err := reopened.GetUserByUsername(context.Background(), "system")
	require.NoError(t, err)
	require.Equal(t, legacy.ID, stored.ID)
	require.Equal(t, legacy.Password, stored.Password)
	require.Equal(t, pkg.UserRoleAdmin, stored.Role)
	require.NoError(t, stored.CheckPassword("synthetic-legacy-password"))
	var storedSource models.SourceCredential
	require.NoError(t, reopened.(*GormRepository).GormClient.First(&storedSource, "id = ?", source.ID).Error)
	require.Equal(t, legacy.ID, storedSource.UserID)
}
