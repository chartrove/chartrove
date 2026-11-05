package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRegistrationSQLiteSerializesBootstrapAcrossRepositories(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("database.type", "sqlite")
			cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-registration.db"))
			cfg.Set("database.encryption.enabled", true)
			cfg.Set("database.encryption.key", "synthetic-registration-key")
			repos := []*GormRepository{}
			for i := 0; i < 2; i++ {
				db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
				require.NoError(t, err)
				repos = append(repos, db.(*GormRepository))
			}
			var trace bytes.Buffer
			repos[0].GormClient.Logger = logger.New(log.New(&trace, "", 0), logger.Config{LogLevel: logger.Info, Colorful: false})
			start := make(chan struct{})
			results := make(chan error, 2)
			users := []*models.User{{Username: "synthetic-bootstrap-one", Password: "synthetic-private-password", Role: pkg.UserRoleAdmin}, {Username: "synthetic-bootstrap-two", Password: "synthetic-private-password", Role: pkg.UserRoleAdmin}}
			for i := range users {
				go func(i int) { <-start; results <- repos[i].RegisterUser(context.Background(), users[i], enabled) }(i)
			}
			close(start)
			success, closed := 0, 0
			for range users {
				err := <-results
				if err == nil {
					success++
				} else {
					require.ErrorIs(t, err, ErrRegistrationClosed)
					closed++
				}
			}
			if enabled {
				require.Equal(t, 2, success)
				require.Zero(t, closed)
			} else {
				require.Equal(t, 1, success)
				require.Equal(t, 1, closed)
			}
			var stored []models.User
			require.NoError(t, repos[0].GormClient.Find(&stored).Error)
			require.Len(t, stored, success)
			admins := 0
			for _, u := range stored {
				if u.Role == pkg.UserRoleAdmin {
					admins++
				}
				require.NoError(t, u.CheckPassword("synthetic-private-password"))
			}
			require.Equal(t, 1, admins)
			var sources int64
			require.NoError(t, repos[0].GormClient.Model(&models.SourceCredential{}).Count(&sources).Error)
			require.Equal(t, int64(success), sources)
			require.NotContains(t, trace.String(), "synthetic-private-password")
			require.NotContains(t, trace.String(), "$2a$")
		})
	}
}

func TestRegistrationInitializationRollbackAndCancellation(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-registration-rollback.db"))
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register("synthetic_registration_source_failure", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "source_credentials" {
			_ = tx.AddError(errors.New("synthetic initialization failure"))
		}
	}))
	require.Error(t, r.RegisterUser(context.Background(), &models.User{Username: "synthetic-failed-initialization", Password: "synthetic-password"}, false))
	count, err := r.GetUserCount(context.Background())
	require.NoError(t, err)
	require.Zero(t, count)
	require.NoError(t, r.GormClient.Callback().Create().Remove("synthetic_registration_source_failure"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Error(t, r.RegisterUser(ctx, &models.User{Username: "synthetic-cancelled", Password: "synthetic-password"}, false))
	user := &models.User{Username: "synthetic-recovered-bootstrap", Password: "synthetic-password", Role: pkg.UserRoleUser}
	require.NoError(t, r.RegisterUser(context.Background(), user, false))
	require.Equal(t, pkg.UserRoleAdmin, user.Role)
}
