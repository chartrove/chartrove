package database

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOwnedResourceUpdatesSQLite(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-owned-updates.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-owned-updates-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	runOwnedResourceUpdates(t, db.(*GormRepository))
}

func runOwnedResourceUpdates(t *testing.T, r *GormRepository) {
	t.Helper()
	owner := &models.User{Username: "synthetic-owned-editor", Password: "synthetic-password"}
	delegate := &models.User{Username: "synthetic-owned-delegate", Password: "synthetic-password"}
	admin := &models.User{Username: "synthetic-owned-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}
	for _, u := range []*models.User{owner, delegate, admin} {
		require.NoError(t, r.CreateUser(context.Background(), u))
	}
	actor := func(u *models.User) context.Context {
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username)
		return context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, u.ID)
	}
	ctx := actor(owner)
	sources, err := r.GetSources(ctx)
	require.NoError(t, err)
	source := sources[0].ID
	newLocal := func() *models.ResourceBase {
		local, err := r.CreateLocalResource(ctx, source, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic original"},"valueQuantity":{"value":1.00}}`))
		require.NoError(t, err)
		return local
	}
	raw := func(id string) json.RawMessage {
		return json.RawMessage(fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic updated"},"valueQuantity":{"value":0.1234567890123456789}}`, id))
	}
	local := newLocal()
	title, date, uri := "Synthetic title", time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), "urn:synthetic:origin"
	require.NoError(t, r.GormClient.Table("fhir_observation").Where("id = ?", local.ID).Updates(map[string]interface{}{"sort_title": title, "sort_date": date, "source_uri": uri}).Error)
	grant, err := r.SetSourceGrant(ctx, source, delegate.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	for _, u := range []*models.User{delegate, admin} {
		require.ErrorIs(t, r.UpdateOwnedResource(actor(u), source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(local.SourceResourceID)}), ErrSourceAccessDenied)
	}
	wrongUUID := context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, uuid.New())
	require.Error(t, r.UpdateOwnedResource(wrongUUID, source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(local.SourceResourceID)}))
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	require.Error(t, r.UpdateOwnedResource(cancelled, source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(local.SourceResourceID)}))
	require.NoError(t, r.UpdateOwnedResource(ctx, source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(local.SourceResourceID)}))
	stored, err := r.GetSharedResource(actor(delegate), source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.True(t, stored.Editable)
	require.Equal(t, local.ID, stored.ID)
	require.Equal(t, local.CreatedAt, stored.CreatedAt)
	require.Equal(t, owner.ID, stored.UserID)
	require.Equal(t, title, *stored.SortTitle)
	require.True(t, stored.SortDate.Equal(date))
	require.Equal(t, uri, *stored.SourceUri)
	var compact bytes.Buffer
	require.NoError(t, json.Compact(&compact, stored.ResourceRaw))
	require.Contains(t, compact.String(), `"value":0.1234567890123456789`)
	var indexed dm.FhirObservation
	require.NoError(t, r.GormClient.Where("id = ?", local.ID).First(&indexed).Error)
	require.Contains(t, string(indexed.Code), "Synthetic updated")
	require.NotContains(t, string(indexed.Code), "Synthetic original")
	audit, err := r.ListSourceAccessAudit(ctx, source)
	require.NoError(t, err)
	ownedEdits := 0
	for _, e := range audit {
		if e.Action == "edit_owned" {
			ownedEdits++
			require.Equal(t, owner.ID, e.ActorUserID)
			require.Equal(t, owner.ID, e.OwnerUserID)
		}
	}
	require.Equal(t, 1, ownedEdits)
	empty := ""
	require.NoError(t, r.UpdateOwnedResource(ctx, source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(local.SourceResourceID), SortTitle: &empty, SetSortDate: true}))
	stored, err = r.GetSharedResource(ctx, source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.Empty(t, *stored.SortTitle)
	require.Nil(t, stored.SortDate)

	// Audit failure rolls back raw data and presentation changes together.
	events := &ownedEditEvents{Interface: r.EventBus}
	r.EventBus = events
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register("synthetic_owned_audit_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "source_access_audits" {
			_ = tx.AddError(errors.New("synthetic audit unavailable"))
		}
	}))
	failed := json.RawMessage(fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"final","code":{"text":"Synthetic must roll back"}}`, local.SourceResourceID))
	require.Error(t, r.UpdateOwnedResource(ctx, source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: failed, SortTitle: &title}))
	require.Zero(t, events.published, "rolled-back edits must not emit success notifications")
	require.NoError(t, r.GormClient.Callback().Create().Remove("synthetic_owned_audit_failure"))
	rolledBack, err := r.GetSharedResource(ctx, source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.JSONEq(t, string(stored.ResourceRaw), string(rolledBack.ResourceRaw))
	require.Empty(t, *rolledBack.SortTitle)
	require.True(t, rolledBack.Editable)

	// Existing local provenance remains subject to current revocation/expiration.
	require.NoError(t, r.RevokeSourceGrant(ctx, source, grant.ID))
	require.ErrorIs(t, r.UpdateSharedResource(actor(delegate), source, "Observation", local.SourceResourceID, raw(local.SourceResourceID)), ErrSourceAccessDenied)
	withoutValue := json.RawMessage(fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic updated"},"dataAbsentReason":{"text":"Synthetic not measured"}}`, local.SourceResourceID))
	require.NoError(t, r.UpdateOwnedResource(ctx, source, "Observation", local.SourceResourceID, OwnedResourceUpdate{ResourceRaw: withoutValue}))
	require.Equal(t, 1, events.published)
	indexed = dm.FhirObservation{}
	require.NoError(t, r.GormClient.Where("id = ?", local.ID).First(&indexed).Error)
	require.JSONEq(t, `[]`, string(indexed.ValueQuantity), "removed indexed values must clear along with the raw value")
	grant, err = r.SetSourceGrant(ctx, source, delegate.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	require.NoError(t, r.UpdateSharedResource(actor(delegate), source, "Observation", local.SourceResourceID, raw(local.SourceResourceID)))
	require.NoError(t, r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", grant.ID).Update("expires_at", time.Now().Add(-time.Minute).UTC()).Error)
	require.ErrorIs(t, r.UpdateSharedResource(actor(delegate), source, "Observation", local.SourceResourceID, raw(local.SourceResourceID)), ErrSourceAccessDenied)

	// Owner edits do not promote SDK/imported records. A later SDK replacement
	// must still remove local eligibility, even inside the owner's local source.
	external := &models.SourceCredential{Patient: "synthetic-external-patient", PlatformType: sp.PlatformType("synthetic-external")}
	require.NoError(t, r.CreateSource(ctx, external))
	_, err = r.UpsertRawResource(ctx, external, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: "synthetic-imported", ResourceRaw: raw("synthetic-imported")})
	require.NoError(t, err)
	require.NoError(t, r.UpdateOwnedResource(ctx, external.ID, "Observation", "synthetic-imported", OwnedResourceUpdate{ResourceRaw: raw("synthetic-imported")}))
	imported, err := r.GetSharedResource(ctx, external.ID, "Observation", "synthetic-imported")
	require.NoError(t, err)
	require.False(t, imported.Editable)
	_, err = r.SetSourceGrant(ctx, external.ID, delegate.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	require.ErrorIs(t, r.UpdateSharedResource(actor(delegate), external.ID, "Observation", "synthetic-imported", raw("synthetic-imported")), ErrSourceAccessDenied)
	replaced := newLocal()
	_, err = r.UpsertRawResource(ctx, &sources[0], sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: replaced.SourceResourceID, ResourceRaw: raw(replaced.SourceResourceID)})
	require.NoError(t, err)
	require.NoError(t, r.UpdateOwnedResource(ctx, source, "Observation", replaced.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(replaced.SourceResourceID)}))
	replacedDetail, err := r.GetSharedResource(ctx, source, "Observation", replaced.SourceResourceID)
	require.NoError(t, err)
	require.False(t, replacedDetail.Editable)

	// A missing/deleted row is never recreated by an edit.
	deleted := newLocal()
	require.NoError(t, r.GormClient.Table("fhir_observation").Where("id = ?", deleted.ID).Update("deleted_at", time.Now().UTC()).Error)
	require.ErrorIs(t, r.UpdateOwnedResource(ctx, source, "Observation", deleted.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(deleted.SourceResourceID)}), ErrSourceAccessDenied)
	var live int64
	require.NoError(t, r.GormClient.Table("fhir_observation").Where("source_id = ? AND source_resource_id = ? AND deleted_at IS NULL", source, deleted.SourceResourceID).Count(&live).Error)
	require.Zero(t, live)

	// Change ownership/liveness after actor lookup but before the first write.
	// The final database predicates must deny rather than relying on that lookup.
	for _, boundary := range []string{"source deletion", "source ownership", "actor deletion"} {
		t.Run(boundary, func(t *testing.T) {
			raced := newLocal()
			type boundaryKey struct{}
			editCtx := context.WithValue(ctx, boundaryKey{}, true)
			injected := false
			var competingErr error
			require.NoError(t, r.GormClient.Callback().Update().Before("gorm:update").Register("synthetic_owned_boundary", func(tx *gorm.DB) {
				if tx.Statement.Table != "fhir_observation" || tx.Statement.Context.Value(boundaryKey{}) != true {
					return
				}
				injected = true
				switch boundary {
				case "source deletion":
					competingErr = r.GormClient.Table("source_credentials").Where("id = ?", source).Update("deleted_at", time.Now().UTC()).Error
				case "source ownership":
					competingErr = r.GormClient.Table("source_credentials").Where("id = ?", source).Update("user_id", admin.ID).Error
				case "actor deletion":
					competingErr = r.GormClient.Table("users").Where("id = ?", owner.ID).Update("deleted_at", time.Now().UTC()).Error
				}
			}))
			err := r.UpdateOwnedResource(editCtx, source, "Observation", raced.SourceResourceID, OwnedResourceUpdate{ResourceRaw: raw(raced.SourceResourceID)})
			require.NoError(t, r.GormClient.Callback().Update().Remove("synthetic_owned_boundary"))
			require.True(t, injected)
			require.NoError(t, competingErr)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
			require.NoError(t, r.GormClient.Table("source_credentials").Where("id = ?", source).Updates(map[string]interface{}{"deleted_at": nil, "user_id": owner.ID}).Error)
			require.NoError(t, r.GormClient.Table("users").Where("id = ?", owner.ID).Update("deleted_at", nil).Error)
			unchanged, err := r.GetSharedResource(ctx, source, "Observation", raced.SourceResourceID)
			require.NoError(t, err)
			require.JSONEq(t, string(raced.ResourceRaw), string(unchanged.ResourceRaw))
			require.True(t, unchanged.Editable)
		})
	}
}

type ownedEditEvents struct {
	event_bus.Interface
	published int
}

func (e *ownedEditEvents) PublishMessage(message models.EventInterface) error {
	e.published++
	return e.Interface.PublishMessage(message)
}
