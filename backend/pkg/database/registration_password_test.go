package database

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func registrationPasswordRepository(t *testing.T) (config.Interface, *GormRepository) {
	t.Helper()
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-registration-password.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-registration-password-key")
	r, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	return cfg, r.(*GormRepository)
}

func TestRegistrationHashesBeforeSQLiteWriterReservation(t *testing.T) {
	cfg, actor := registrationPasswordRepository(t)
	reader, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	user := &models.User{Username: "synthetic-bootstrap-password", Password: "synthetic-password", Role: pkg.UserRoleUser}
	reservations, hashCostAtReservation := 0, 0
	require.NoError(t, actor.GormClient.Callback().Raw().After("gorm:raw").Register("synthetic_observe_registration_password", func(tx *gorm.DB) {
		if strings.HasPrefix(tx.Statement.SQL.String(), "UPDATE users ") {
			reservations++
			// Observe the real reservation boundary in the registration goroutine;
			// do not replace hashing, alter locking, or depend on scheduler timing.
			hashCostAtReservation, _ = bcrypt.Cost([]byte(user.Password))
		}
	}))
	require.NoError(t, actor.RegisterUser(context.Background(), user, false))
	require.Equal(t, 1, reservations)
	require.Equal(t, 14, hashCostAtReservation, "password work must finish before the writer reservation")
	stored, err := reader.GetUserByUsername(context.Background(), user.Username)
	require.NoError(t, err)
	require.Equal(t, user.ID, stored.ID)
	require.Equal(t, pkg.UserRoleAdmin, stored.Role)
	require.NoError(t, stored.CheckPassword("synthetic-password"), "registration must hash exactly once")
	var settings, sources int64
	require.NoError(t, reader.(*GormRepository).GormClient.Model(&models.UserSettingEntry{}).Where("user_id = ?", user.ID).Count(&settings).Error)
	require.NoError(t, reader.(*GormRepository).GormClient.Model(&models.SourceCredential{}).Where("user_id = ?", user.ID).Count(&sources).Error)
	require.EqualValues(t, 1, settings)
	require.EqualValues(t, 1, sources)
}

func TestRegistrationRefusesBeforeWriterReservation(t *testing.T) {
	for _, name := range []string{"closed", "unhashable", "cancelled", "reserved"} {
		t.Run(name, func(t *testing.T) {
			_, r := registrationPasswordRepository(t)
			if name == "closed" {
				// The independent administrator path still creates and hashes a user
				// without depending on public registration availability.
				admin := &models.User{Username: "synthetic-existing-admin", Password: "synthetic-admin-password", Role: pkg.UserRoleAdmin}
				require.NoError(t, r.CreateUser(context.Background(), admin))
				cost, err := bcrypt.Cost([]byte(admin.Password))
				require.NoError(t, err)
				require.Equal(t, 14, cost)
				require.NoError(t, admin.CheckPassword("synthetic-admin-password"))
			}
			reservations := 0
			require.NoError(t, r.GormClient.Callback().Raw().After("gorm:raw").Register("synthetic_observe_early_registration_refusal", func(tx *gorm.DB) {
				if strings.HasPrefix(tx.Statement.SQL.String(), "UPDATE users ") {
					reservations++
				}
			}))
			ctx := context.Background()
			user := &models.User{Username: "synthetic-refused-registration", Password: strings.Repeat("x", 73), Role: pkg.UserRoleAdmin}
			want := error(bcrypt.ErrPasswordTooLong)
			enabled := true
			switch name {
			case "closed":
				enabled = false
				want = ErrRegistrationClosed
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = context.Canceled
			case "reserved":
				user.Username = "chartrove"
				want = models.ErrReservedUsername
			}
			require.ErrorIs(t, r.RegisterUser(ctx, user, enabled), want)
			require.Zero(t, reservations, "rejected registration must not reserve the writer")
			require.Equal(t, strings.Repeat("x", 73), user.Password, "rejection must not replace the submitted password")
			count, err := r.GetUserCount(context.Background())
			require.NoError(t, err)
			wantCount := 0
			if name == "closed" {
				wantCount = 1
			}
			require.Equal(t, wantCount, count)
		})
	}
}
