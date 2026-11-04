package database

import (
	"context"
	"encoding/json"
	"errors"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"gorm.io/gorm"
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

func TestSourceAccessIsolationAndLifecycle(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-grants.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-source-access-key")
	cfg.Set("database.validation_mode", false)
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	users := []*models.User{
		{Username: "synthetic-owner", Password: "synthetic-password"},
		{Username: "synthetic-delegate", Password: "synthetic-password"},
		{Username: "synthetic-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin},
	}
	contexts := []context.Context{}
	for _, u := range users {
		require.NoError(t, r.CreateUser(context.Background(), u))
		contexts = append(contexts, context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username))
	}
	owner, delegate, admin := contexts[0], contexts[1], contexts[2]
	sources, err := r.GetSources(owner)
	require.NoError(t, err)
	source := sources[0]
	otherSource := &models.SourceCredential{Patient: "synthetic-unshared-patient", Display: "Unshared synthetic source", AccessToken: "synthetic-secret"}
	require.NoError(t, r.CreateSource(owner, otherSource))
	local, err := r.CreateLocalResource(owner, source.ID, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic original"}}`))
	require.NoError(t, err)
	_, err = r.UpsertRawResource(owner, otherSource, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: local.SourceResourceID, ResourceRaw: json.RawMessage(`{"resourceType":"Observation","id":"other","status":"final"}`)})
	require.NoError(t, err)
	for _, ctx := range []context.Context{delegate, admin} {
		_, err = r.GetSharedSourceSummary(ctx, source.ID)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		_, err = r.GetSharedResource(ctx, source.ID, "Observation", local.SourceResourceID)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		_, err = r.SetSourceGrant(ctx, source.ID, users[1].Username, models.SourceAccessEdit, nil)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
	}
	grant, err := r.SetSourceGrant(owner, source.ID, users[1].Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	shared, err := r.ListSharedSources(delegate)
	require.NoError(t, err)
	require.Len(t, shared, 1)
	require.Equal(t, source.ID, shared[0].ID)
	safe, err := json.Marshal(shared)
	require.NoError(t, err)
	require.NotContains(t, string(safe), "access_token")
	require.NotContains(t, string(safe), "password")
	summary, err := r.GetSharedSourceSummary(delegate, source.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, summary.Counts["Observation"])
	rows, err := r.ListSharedResources(delegate, source.ID, "Observation", 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, users[0].ID, rows[0].UserID)
	_, err = r.ListSharedResources(delegate, otherSource.ID, "Observation", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	// Even a historical malformed row bearing the shared source must not leak
	// another owner's data through a source-only predicate.
	malformed := &dm.FhirObservation{}
	malformed.SetOriginBase(models.OriginBase{UserID: users[2].ID, SourceID: source.ID, SourceResourceType: "Observation", SourceResourceID: "synthetic-other-owner"})
	require.NoError(t, malformed.PopulateAndExtractSearchParameters(json.RawMessage(`{"resourceType":"Observation","id":"synthetic-other-owner","status":"final"}`)))
	require.NoError(t, r.GormClient.Create(malformed).Error)
	rows, err = r.ListSharedResources(delegate, source.ID, "Observation", 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	_, err = r.GetSharedResource(delegate, source.ID, "Observation", "synthetic-other-owner")
	require.Error(t, err)

	// Existing APIs remain own-only, including administration and credential access.
	for _, ctx := range []context.Context{delegate, admin} {
		_, err = r.GetSource(ctx, source.ID.String())
		require.Error(t, err)
		_, err = r.GetResourceBySourceId(ctx, source.ID.String(), local.SourceResourceID)
		require.Error(t, err)
	}
	updated := json.RawMessage(`{"resourceType":"Observation","id":"` + local.SourceResourceID + `","status":"amended","code":{"text":"Synthetic changed"}}`)
	require.ErrorIs(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated), ErrSourceAccessDenied)
	_, err = r.SetSourceGrant(delegate, source.ID, users[2].Username, models.SourceAccessEdit, nil)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	_, err = r.SetSourceGrant(owner, source.ID, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	require.NoError(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated))
	stored, err := r.GetSharedResource(owner, source.ID, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.JSONEq(t, string(updated), string(stored.ResourceRaw))
	var indexed dm.FhirObservation
	require.NoError(t, r.GormClient.Where("source_id = ? AND source_resource_id = ?", source.ID, local.SourceResourceID).First(&indexed).Error)
	require.Contains(t, string(indexed.Code), "Synthetic changed")
	require.NotContains(t, string(indexed.Code), "Synthetic original")
	// An audit failure rolls the clinical edit back in the same transaction.
	const callback = "synthetic:audit-failure"
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "source_access_audits" {
			tx.AddError(errors.New("synthetic audit failure"))
		}
	}))
	require.Error(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, json.RawMessage(`{"resourceType":"Observation","id":"`+local.SourceResourceID+`","status":"final"}`)))
	require.NoError(t, r.GormClient.Callback().Create().Remove(callback))
	stored, err = r.GetSharedResource(owner, source.ID, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.JSONEq(t, string(updated), string(stored.ResourceRaw))

	require.ErrorIs(t, r.UpdateSharedResource(delegate, otherSource.ID, "Observation", local.SourceResourceID, updated), ErrSourceAccessDenied)
	require.Error(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, json.RawMessage(`{"resourceType":"Patient","id":"changed"}`)))
	require.ErrorIs(t, r.RevokeSourceGrant(delegate, source.ID, grant.ID), ErrSourceAccessDenied)
	require.NoError(t, r.RevokeSourceGrant(owner, source.ID, grant.ID))
	_, err = r.GetSharedResource(delegate, source.ID, "Observation", local.SourceResourceID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.ErrorIs(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated), ErrSourceAccessDenied)
	expires := time.Now().Add(time.Hour)
	_, err = r.SetSourceGrant(owner, source.ID, users[1].Username, models.SourceAccessEdit, &expires)
	require.NoError(t, err)
	// Expire a once-valid grant to exercise enforcement, not only input validation.
	require.NoError(t, r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", grant.ID).Update("expires_at", time.Now().Add(-time.Hour)).Error)
	_, err = r.ListSharedResources(delegate, source.ID, "Observation", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	_, err = r.GetSharedSourceSummary(delegate, source.ID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.ErrorIs(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated), ErrSourceAccessDenied)
	_, err = r.SetSourceGrant(owner, source.ID, users[1].Username, "ADMIN", nil)
	require.Error(t, err)
	_, err = r.SetSourceGrant(owner, source.ID, users[0].Username, models.SourceAccessView, nil)
	require.Error(t, err)
	_, err = r.SetSourceGrant(owner, source.ID, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	// Removed clinical values must also be removed from derived indexes.
	withoutCode := json.RawMessage(`{"resourceType":"Observation","id":"` + local.SourceResourceID + `","status":"final"}`)
	require.NoError(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, withoutCode))
	indexed = dm.FhirObservation{}
	require.NoError(t, r.GormClient.Where("source_id = ? AND source_resource_id = ?", source.ID, local.SourceResourceID).First(&indexed).Error)
	require.NotContains(t, string(indexed.Code), "Synthetic changed")
	require.NoError(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated))

	// An SDK import replacing a formerly local resource must remove edit eligibility.
	// A failed replacement must preserve both clinical contents and provenance.
	_, err = r.UpsertRawResource(owner, &source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: local.SourceResourceID, ResourceRaw: updated,
		ReferencedResources: []string{"urn:fastenhealth-fhir:" + uuid.NewString() + ":Patient/synthetic-unavailable"}})
	require.Error(t, err)
	require.NoError(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated))
	_, err = r.UpsertRawResource(owner, &source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: local.SourceResourceID, ResourceRaw: updated})
	require.NoError(t, err)
	require.ErrorIs(t, r.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated), ErrSourceAccessDenied)
	audit, err := r.ListSourceAccessAudit(owner, source.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(audit), 6)
	_, err = r.ListSourceAccessAudit(delegate, source.ID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	_, err = r.CreateLocalResource(delegate, source.ID, updated)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	_, err = r.GetSharedResource(delegate, uuid.New(), "Observation", local.SourceResourceID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.NoError(t, r.Close())
	reopened, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	rr := reopened.(*GormRepository)
	grants, err := rr.ListSourceGrants(owner, source.ID)
	require.NoError(t, err)
	require.Len(t, grants, 1)
	auditAfter, err := rr.ListSourceAccessAudit(owner, source.ID)
	require.NoError(t, err)
	require.Equal(t, len(audit), len(auditAfter))
	_, err = rr.GetSharedResource(delegate, source.ID, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.ErrorIs(t, rr.UpdateSharedResource(delegate, source.ID, "Observation", local.SourceResourceID, updated), ErrSourceAccessDenied)

}
