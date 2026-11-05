package database

import (
	"context"
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

func TestNativeAuthenticatedLookupRetainsUUID(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-native-identity.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-native-identity-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	user := &models.User{Username: "synthetic-reused-name", Password: "synthetic-first-password"}
	require.NoError(t, r.CreateUser(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
	ctx = context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, user.ID)
	stored, err := r.GetCurrentUser(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, stored.ID)
	for _, bad := range []interface{}{uuid.New(), uuid.Nil, "invalid"} {
		_, err = r.GetCurrentUser(context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, bad))
		require.Error(t, err)
	}
	require.NoError(t, r.DeleteCurrentUser(ctx))
	replacement := &models.User{Username: user.Username, Password: "synthetic-second-password"}
	require.NoError(t, r.CreateUser(context.Background(), replacement))
	require.NotEqual(t, user.ID, replacement.ID)
	_, err = r.GetCurrentUser(ctx)
	require.Error(t, err, "an already authenticated request cannot switch to a replacement UUID")
	_, err = r.GetSources(ctx)
	require.Error(t, err)
	fresh := context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, replacement.ID)
	stored, err = r.GetCurrentUser(fresh)
	require.NoError(t, err)
	require.Equal(t, replacement.ID, stored.ID)
}
