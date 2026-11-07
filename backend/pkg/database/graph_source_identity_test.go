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
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestGraphSourceIdentitySQLite(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-graph-identities.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-graph-identities-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	runGraphSourceIdentity(t, db.(*GormRepository))
}

func runGraphSourceIdentity(t *testing.T, r *GormRepository) {
	t.Helper()
	owner := &models.User{Username: "synthetic-graph-owner", Password: "synthetic-password"}
	admin := &models.User{Username: "synthetic-graph-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}
	delegate := &models.User{Username: "synthetic-graph-delegate", Password: "synthetic-password"}
	actor := func(user *models.User) context.Context {
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
		return context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, user.ID)
	}
	for _, user := range []*models.User{owner, admin, delegate} {
		require.NoError(t, r.CreateUser(context.Background(), user))
	}
	ctx := actor(owner)
	sources := []*models.SourceCredential{{Patient: "synthetic-A"}, {Patient: "synthetic-B"}}
	id := "synthetic-same-id"
	for _, source := range sources {
		require.NoError(t, r.CreateSource(ctx, source))
		_, err := r.UpsertRawResource(ctx, source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: id, ResourceRaw: json.RawMessage(`{"resourceType":"Observation","id":"synthetic-same-id","status":"final","code":{"text":"Synthetic value"},"valueInteger":0}`)})
		require.NoError(t, err)
		vertex := &VertexResourcePlaceholder{UserID: owner.ID.String(), SourceID: source.ID.String(), ResourceType: "Observation", ResourceID: id}
		row, err := r.graphResource(ctx, owner, vertex)
		require.NoError(t, err)
		require.Equal(t, source.ID, row.SourceID)
	}
	selected := func(source *models.SourceCredential) models.ResourceGraphOptions {
		return models.ResourceGraphOptions{ResourcesIds: []models.OriginBase{{SourceID: source.ID, SourceResourceType: "Observation", SourceResourceID: id}}}
	}
	assertHidden := func(user *models.User, source *models.SourceCredential) {
		t.Helper()
		vertex := &VertexResourcePlaceholder{UserID: owner.ID.String(), SourceID: source.ID.String(), ResourceType: "Observation", ResourceID: id}
		_, err := r.graphResource(actor(user), user, vertex)
		require.Error(t, err)
		roots, err := r.InflateSelectedResourcesInResourceGraph(user, nil, selected(source))
		require.NoError(t, err)
		require.Empty(t, roots["Observation"])
	}
	for _, user := range []*models.User{admin, delegate} {
		assertHidden(user, sources[0])
	}
	initialGrant, err := r.SetSourceGrant(ctx, sources[0].ID, delegate.Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	// This legacy graph is owner-only. Delegation continues through the explicit
	// source-authorized shared endpoints; a grant never broadens this graph.
	assertHidden(delegate, sources[0])
	for _, state := range []string{"revoked", "expired"} {
		if state == "revoked" {
			require.NoError(t, r.RevokeSourceGrant(ctx, sources[0].ID, initialGrant.ID))
		} else {
			expires := time.Now().Add(time.Hour)
			grant, err := r.SetSourceGrant(ctx, sources[0].ID, delegate.Username, models.SourceAccessView, &expires)
			require.NoError(t, err)
			require.NoError(t, r.GormClient.Model(grant).Update("expires_at", time.Now().Add(-time.Hour)).Error)
		}
		assertHidden(delegate, sources[0])
	}
	_, err = r.GetFlattenedResourceGraph(context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, uuid.New()), pkg.ResourceGraphTypeMedicalHistory, selected(sources[0]))
	require.Error(t, err)
	forged := *owner
	forged.Username = admin.Username
	assertHidden(&forged, sources[0])
	// Simulate stale graph placeholders after deletion. Content predicates must
	// recheck live resource, source and native identity before returning rows.
	require.NoError(t, r.GormClient.Table("fhir_observation").Where("user_id = ? AND source_id = ?", owner.ID, sources[0].ID).Update("deleted_at", time.Now()).Error)
	assertHidden(owner, sources[0])
	require.NoError(t, r.GormClient.Model(sources[1]).Update("deleted_at", time.Now()).Error)
	assertHidden(owner, sources[1])
	require.NoError(t, r.GormClient.Model(sources[1]).Update("deleted_at", nil).Error)
	require.NoError(t, r.GormClient.Model(owner).Update("deleted_at", time.Now()).Error)
	assertHidden(owner, sources[1])
}
