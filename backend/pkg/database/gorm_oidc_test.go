package database

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestOIDCNativeBindingRevocationAndSourceIsolation(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-oidc.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-oidc-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	a := &models.User{Username: "synthetic-owner", Password: "synthetic-password", Email: "duplicate@example.invalid", Role: pkg.UserRoleAdmin}
	b := &models.User{Username: "synthetic-delegate", Password: "synthetic-password", Email: a.Email, Role: pkg.UserRoleUser}
	for _, u := range []*models.User{a, b} {
		require.NoError(t, r.CreateUser(context.Background(), u))
		require.NoError(t, u.CheckPassword("synthetic-password"))
	}
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)
	_, _, err = r.CompleteOIDC(ctx, "https://issuer.example", "unknown", uuid.Nil, "", time.Now().UTC(), expires)
	require.ErrorIs(t, err, ErrOIDCDenied)
	user, s, err := r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", b.ID, b.Password, time.Now().UTC(), expires)
	require.NoError(t, err)
	require.Equal(t, b.ID, user.ID)
	require.Equal(t, pkg.UserRoleUser, user.Role)
	require.NoError(t, r.ValidateOIDCSession(ctx, s.ID, b.ID, "https://issuer.example"))
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, s.ID, a.ID, "https://issuer.example"), ErrOIDCDenied)
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, s.ID, b.ID, "https://other.example"), ErrOIDCDenied)
	_, _, err = r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", a.ID, a.Password, time.Now().UTC(), expires)
	require.ErrorIs(t, err, ErrOIDCDenied)
	_, _, err = r.CompleteOIDC(ctx, "https://issuer.example", "new-subject", a.ID, "stale-password-hash", time.Now().UTC(), expires)
	require.ErrorIs(t, err, ErrOIDCDenied)
	_, login, err := r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", uuid.Nil, "", time.Now().UTC(), expires)
	require.NoError(t, err)
	require.Equal(t, b.ID, login.UserID)
	owner := context.WithValue(ctx, pkg.ContextKeyTypeAuthUsername, a.Username)
	delegate := context.WithValue(ctx, pkg.ContextKeyTypeAuthUsername, b.Username)
	sources, err := r.GetSources(owner)
	require.NoError(t, err)
	source := sources[0].ID
	local, err := r.CreateLocalResource(owner, source, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic OIDC record"}}`))
	require.NoError(t, err)
	_, err = r.GetSharedResource(delegate, source, "Observation", local.SourceResourceID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	grant, err := r.SetSourceGrant(owner, source, b.Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	_, err = r.GetSharedResource(delegate, source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.ErrorIs(t, r.UpdateSharedResource(delegate, source, "Observation", local.SourceResourceID, json.RawMessage(local.ResourceRaw)), ErrSourceAccessDenied)
	require.NoError(t, r.RevokeSourceGrant(owner, source, grant.ID))
	_, err = r.GetSharedResource(delegate, source, "Observation", local.SourceResourceID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.ErrorIs(t, r.RevokeOIDCIdentity(ctx, a.ID, s.IdentityID), ErrOIDCDenied)
	require.NoError(t, r.RevokeOIDCSession(ctx, b.ID, s.ID))
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, s.ID, b.ID, "https://issuer.example"), ErrOIDCDenied)
	require.NoError(t, r.ValidateOIDCSession(ctx, login.ID, b.ID, "https://issuer.example"))
	require.NoError(t, r.RevokeOIDCIdentity(ctx, b.ID, s.IdentityID))
	// Model a READ COMMITTED login whose old-generation INSERT commits after
	// unlink's session sweep. Re-link must never make that session valid.
	stale := *login
	stale.ID = uuid.New()
	stale.RevokedAt = nil
	require.NoError(t, r.GormClient.Create(&stale).Error)
	_, _, err = r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", b.ID, b.Password, time.Now().UTC().Add(-time.Minute), expires)
	require.ErrorIs(t, err, ErrOIDCDenied) // A pending link cannot undo later revocation.
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, login.ID, b.ID, "https://issuer.example"), ErrOIDCDenied)
	_, _, err = r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", uuid.Nil, "", time.Now().UTC(), expires)
	require.ErrorIs(t, err, ErrOIDCDenied)
	_, relinked, err := r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", b.ID, b.Password, time.Now().UTC(), expires)
	require.NoError(t, err)
	require.NoError(t, r.ValidateOIDCSession(ctx, relinked.ID, b.ID, "https://issuer.example"))
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, stale.ID, b.ID, "https://issuer.example"), ErrOIDCDenied)
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, login.ID, b.ID, "https://issuer.example"), ErrOIDCDenied)
	require.NoError(t, r.GormClient.Model(&models.OIDCSession{}).Where("id = ?", relinked.ID).Update("expires_at", time.Now().UTC().Add(-time.Minute)).Error)
	require.ErrorIs(t, r.ValidateOIDCSession(ctx, relinked.ID, b.ID, "https://issuer.example"), ErrOIDCDenied)
	identities, err := r.ListOIDCIdentities(ctx, a.ID)
	require.NoError(t, err)
	require.Empty(t, identities)
	audit, err := r.ListOIDCAudit(ctx, b.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(audit), 6)
	// Deleted/recreated usernames cannot inherit the original UUID's binding.
	require.NoError(t, r.GormClient.Model(&models.User{}).Where("id = ?", b.ID).Update("deleted_at", time.Now()).Error)
	_, _, err = r.CompleteOIDC(ctx, "https://issuer.example", "subject-b", uuid.Nil, "", time.Now().UTC(), expires)
	require.ErrorIs(t, err, ErrOIDCDenied)
}
