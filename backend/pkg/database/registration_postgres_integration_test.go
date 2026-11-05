//go:build postgres_integration

package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRegistrationPostgresSerializesBootstrap(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			dsn := disposablePostgres(t)
			settings := dsn.Query()
			settings.Set("default_transaction_isolation", "repeatable read")
			dsn.RawQuery = settings.Encode()
			db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{Logger: logger.Discard, IgnoreRelationshipsWhenMigrating: true})
			require.NoError(t, err)
			pool, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, pool.Close()) })
			require.NoError(t, db.AutoMigrate(&models.User{}, &models.UserSettingEntry{}, &models.SourceCredential{}))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			start := make(chan struct{})
			results := make(chan error, 2)
			for i := 0; i < 2; i++ {
				go func(i int) {
					<-start
					r := &GormRepository{GormClient: db}
					results <- r.RegisterUser(ctx, &models.User{Username: fmt.Sprintf("synthetic-pg-bootstrap-%d", i), Password: "synthetic-password", Role: pkg.UserRoleAdmin}, enabled)
				}(i)
			}
			close(start)
			success := 0
			for i := 0; i < 2; i++ {
				err := <-results
				if err == nil {
					success++
				} else {
					require.ErrorIs(t, err, ErrRegistrationClosed)
				}
			}
			if enabled {
				require.Equal(t, 2, success)
			} else {
				require.Equal(t, 1, success)
			}
			var users []models.User
			require.NoError(t, db.Find(&users).Error)
			require.Len(t, users, success)
			admins := 0
			for _, u := range users {
				if u.Role == pkg.UserRoleAdmin {
					admins++
				}
			}
			require.Equal(t, 1, admins)
		})
	}
}
