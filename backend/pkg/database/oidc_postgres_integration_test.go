//go:build postgres_integration

package database

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOIDCPostgresRevocationDuringLoginCommit(t *testing.T) {
	// Reuse the existing pinned, loopback-only, tmpfs-backed disposable fixture.
	// This never accepts an existing database or a host data volume.
	dsn := disposablePostgres(t)
	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.OIDCIdentity{}, &models.OIDCSession{}, &models.OIDCAudit{}))
	r := &GormRepository{GormClient: db}
	user := models.User{Username: "synthetic-oidc-postgres"}
	require.NoError(t, user.HashPassword("synthetic-password"))
	require.NoError(t, db.Create(&user).Error)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	expires := time.Now().Add(time.Hour)
	_, initial, err := r.CompleteOIDC(ctx, "https://issuer.example", "synthetic-subject", user.ID, user.Password, time.Now().UTC(), expires)
	require.NoError(t, err)
	paused, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	var held atomic.Bool
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("synthetic_oidc_pause_before_commit", func(tx *gorm.DB) {
		if event, ok := tx.Statement.Dest.(*models.OIDCAudit); ok && event.Action == "login" && held.CompareAndSwap(false, true) {
			close(paused)
			select {
			case <-release:
			case <-ctx.Done():
			}
		}
	}))
	type result struct {
		session *models.OIDCSession
		err     error
	}
	done := make(chan result, 1)
	go func() {
		_, session, err := r.CompleteOIDC(ctx, "https://issuer.example", "synthetic-subject", uuid.Nil, "", time.Now().UTC(), expires)
		done <- result{session, err}
	}()
	select {
	case <-paused:
	case <-ctx.Done():
		t.Fatal("login did not reach pre-commit barrier")
	}
	// The login has performed all reads and writes but has not committed. Under
	// READ COMMITTED unlink's sweep cannot see its new session row.
	require.NoError(t, r.RevokeOIDCIdentity(ctx, user.ID, initial.IdentityID))
	once.Do(func() { close(release) })
	var raced result
	select {
	case raced = <-done:
	case <-ctx.Done():
		t.Fatal("login did not commit")
	}
	require.NoError(t, raced.err)
	var stored models.OIDCSession
	require.NoError(t, db.Where("id = ?", raced.session.ID).First(&stored).Error)
	require.Nil(t, stored.RevokedAt) // Prove unlink's sweep missed this commit.
	_, fresh, err := r.CompleteOIDC(ctx, "https://issuer.example", "synthetic-subject", user.ID, user.Password, time.Now().UTC(), expires)
	require.NoError(t, err)
	require.NotEqual(t, stored.IdentityGeneration, fresh.IdentityGeneration)
	require.NoError(t, r.ValidateOIDCSession(ctx, fresh.ID, user.ID, "https://issuer.example"))
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, raced.session.ID, user.ID, "https://issuer.example"), ErrOIDCDenied)
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, initial.ID, user.ID, "https://issuer.example"), ErrOIDCDenied)
}
