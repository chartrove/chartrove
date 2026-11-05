package database

import (
	"context"
	"encoding/json"
	"errors"
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
	"gorm.io/gorm"
)

func TestOwnedMutationTargetsSQLite(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-mutation-targets.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-mutation-targets-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	runOwnedMutationTargets(t, db.(*GormRepository))
}

func runOwnedMutationTargets(t *testing.T, r *GormRepository) {
	t.Helper()
	owner := &models.User{Username: "synthetic-mutation-owner", Password: "synthetic-password"}
	other := &models.User{Username: "synthetic-mutation-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}
	for _, user := range []*models.User{owner, other} {
		require.NoError(t, r.CreateUser(context.Background(), user))
	}
	actor := func(user *models.User) context.Context {
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
		return context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, user.ID)
	}
	ctx, foreign := actor(owner), actor(other)
	sources := []*models.SourceCredential{{Patient: "synthetic-target-A", AccessToken: "synthetic-secret-A"}, {Patient: "synthetic-target-B"}, {Patient: "synthetic-target-foreign"}}
	id := "synthetic-repeated-id"
	raw := json.RawMessage(`{"resourceType":"Observation","id":"synthetic-repeated-id","status":"final","code":{"text":"Synthetic private target"}}`)
	for i, source := range sources {
		current := ctx
		if i == 2 {
			current = foreign
		}
		require.NoError(t, r.CreateSource(current, source))
		_, err := r.UpsertRawResource(current, source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: id, ResourceRaw: raw})
		require.NoError(t, err)
	}
	_, err := r.ResolveOwnedResource(ctx, "Observation", id, nil)
	require.ErrorIs(t, err, ErrOwnedResourceAmbiguous)
	require.ErrorIs(t, r.DeleteResourceByTypeAndId(ctx, "Observation", id), ErrOwnedResourceAmbiguous)
	resolved, err := r.ResolveOwnedResource(ctx, "Observation", id, &sources[1].ID)
	require.NoError(t, err)
	require.Equal(t, sources[1].ID, resolved.SourceID)
	require.Empty(t, resolved.ResourceRaw, "target resolution must not fetch clinical content")
	for _, source := range []*uuid.UUID{&sources[2].ID, func() *uuid.UUID { id := uuid.New(); return &id }()} {
		_, err := r.ResolveOwnedResource(ctx, "Observation", id, source)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
	}
	_, err = r.ResolveOwnedResource(context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, uuid.New()), "Observation", id, nil)
	require.Error(t, err)
	_, err = r.ResolveOwnedResource(ctx, "Observation", id, &uuid.Nil)
	require.ErrorIs(t, err, ErrSourceAccessInput)
	_, err = r.SetSourceGrant(ctx, sources[0].ID, other.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	_, err = r.ResolveOwnedResource(foreign, "Observation", id, &sources[0].ID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.ErrorIs(t, r.DeleteOwnedResource(foreign, sources[0].ID, "Observation", id), ErrSourceAccessDenied)

	// Precisely one selected source is deleted. A foreign owner's equal ID never
	// makes the remaining legacy target ambiguous and is never removed by it.
	require.NoError(t, r.DeleteOwnedResource(ctx, sources[1].ID, "Observation", id))
	_, err = r.GetSharedResource(ctx, sources[1].ID, "Observation", id)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = r.GetSharedResource(ctx, sources[0].ID, "Observation", id)
	require.NoError(t, err)
	resolved, err = r.ResolveOwnedResource(ctx, "Observation", id, nil)
	require.NoError(t, err)
	require.Equal(t, sources[0].ID, resolved.SourceID)
	require.NoError(t, r.DeleteResourceByTypeAndId(ctx, "Observation", id))
	_, err = r.GetSharedResource(foreign, sources[2].ID, "Observation", id)
	require.NoError(t, err)
	require.ErrorIs(t, r.DeleteOwnedResource(ctx, sources[0].ID, "Observation", id), ErrSourceAccessDenied)

	ownedSources, err := r.GetSources(ctx)
	require.NoError(t, err)
	var localSource uuid.UUID
	for _, source := range ownedSources {
		if source.Patient == "" {
			localSource = source.ID
		}
	}
	require.NotEqual(t, uuid.Nil, localSource)
	newLocal := func() *models.ResourceBase {
		row, err := r.CreateLocalResource(ctx, localSource, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic local deletion"}}`))
		require.NoError(t, err)
		return row
	}
	local := newLocal()
	events := &ownedEditEvents{Interface: r.EventBus}
	r.EventBus = events
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register("synthetic_delete_audit_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "source_access_audits" {
			_ = tx.AddError(errors.New("synthetic audit unavailable"))
		}
	}))
	require.Error(t, r.DeleteOwnedResource(ctx, localSource, "Observation", local.SourceResourceID))
	require.NoError(t, r.GormClient.Callback().Create().Remove("synthetic_delete_audit_failure"))
	retained, err := r.GetSharedResource(ctx, localSource, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.True(t, retained.Editable, "failed deletion must retain its local provenance too")
	require.Zero(t, events.published)
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	require.Error(t, r.DeleteOwnedResource(cancelled, localSource, "Observation", local.SourceResourceID))
	require.NoError(t, r.DeleteOwnedResource(ctx, localSource, "Observation", local.SourceResourceID))
	require.Equal(t, 1, events.published)
	var markers int64
	require.NoError(t, r.GormClient.Model(&models.LocalResource{}).Where("owner_user_id = ? AND source_id = ? AND resource_id = ?", owner.ID, localSource, local.SourceResourceID).Count(&markers).Error)
	require.Zero(t, markers)
	audit, err := r.ListSourceAccessAudit(ctx, localSource)
	require.NoError(t, err)
	deleted := 0
	for _, event := range audit {
		if event.Action == "delete_owned" {
			deleted++
			require.Equal(t, local.SourceResourceID, event.ResourceID)
			require.Equal(t, owner.ID, event.ActorUserID)
		}
	}
	require.Equal(t, 1, deleted)
	for _, boundary := range []string{"source deletion", "source ownership", "actor deletion"} {
		t.Run(boundary, func(t *testing.T) {
			row := newLocal()
			type deleteKey struct{}
			deleteCtx := context.WithValue(ctx, deleteKey{}, true)
			injected := false
			var competingErr error
			require.NoError(t, r.GormClient.Callback().Delete().Before("gorm:delete").Register("synthetic_delete_boundary", func(tx *gorm.DB) {
				if tx.Statement.Table != "fhir_observation" || tx.Statement.Context.Value(deleteKey{}) != true {
					return
				}
				injected = true
				switch boundary {
				case "source deletion":
					competingErr = r.GormClient.Table("source_credentials").Where("id = ?", localSource).Update("deleted_at", time.Now().UTC()).Error
				case "source ownership":
					competingErr = r.GormClient.Table("source_credentials").Where("id = ?", localSource).Update("user_id", other.ID).Error
				case "actor deletion":
					competingErr = r.GormClient.Table("users").Where("id = ?", owner.ID).Update("deleted_at", time.Now().UTC()).Error
				}
			}))
			err := r.DeleteOwnedResource(deleteCtx, localSource, "Observation", row.SourceResourceID)
			require.NoError(t, r.GormClient.Callback().Delete().Remove("synthetic_delete_boundary"))
			require.True(t, injected)
			require.NoError(t, competingErr)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
			require.NoError(t, r.GormClient.Table("source_credentials").Where("id = ?", localSource).Updates(map[string]interface{}{"deleted_at": nil, "user_id": owner.ID}).Error)
			require.NoError(t, r.GormClient.Table("users").Where("id = ?", owner.ID).Update("deleted_at", nil).Error)
			_, err = r.GetSharedResource(ctx, localSource, "Observation", row.SourceResourceID)
			require.NoError(t, err)
		})
	}
}
