package database

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLocalResourceExactIdentityKeys(t *testing.T) {
	for name, raw := range map[string]string{
		"ID alias":                   `{"resourceType":"Observation","id":"synthetic-wrong","ID":"synthetic-id"}`,
		"resourceType alias":         `{"resourceType":"Patient","ResourceType":"Observation","id":"synthetic-id"}`,
		"missing canonical ID":       `{"resourceType":"Observation","ID":"synthetic-id"}`,
		"ambiguous matching aliases": `{"resourceType":"Observation","id":"synthetic-id","ID":"synthetic-id"}`,
		"duplicate ID":               `{"resourceType":"Observation","id":"synthetic-wrong","id":"synthetic-id"}`,
		"duplicate resourceType":     `{"resourceType":"Patient","resourceType":"Observation","id":"synthetic-id"}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseLocalResource(json.RawMessage(raw), "Observation", "synthetic-id")
			require.ErrorIs(t, err, ErrSourceAccessInput)
		})
	}
	_, err := parseLocalResource(json.RawMessage(`{"resourceType":"Observation","id":"synthetic-id","status":"final"}`), "Observation", "synthetic-id")
	require.NoError(t, err)
}

func TestLocalResourceNumbersAndEditContention(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-precision.db"))
	cfg.Set("database.encryption.enabled", false)
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	user := &models.User{Username: "synthetic-precision-owner", Password: "synthetic-password"}
	require.NoError(t, r.CreateUser(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
	sources, err := r.GetSources(ctx)
	require.NoError(t, err)
	source := sources[0].ID
	t.Run("preserve numeric literals", func(t *testing.T) {
		for _, value := range []string{"0.1234567890123456789", "2.00", "9007199254740993"} {
			raw := json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic"},"valueQuantity":{"value":` + value + `}}`)
			local, err := r.CreateLocalResource(ctx, source, raw)
			require.NoError(t, err)
			require.Contains(t, string(local.ResourceRaw), `"value":`+value)
			detail, err := r.GetSharedResource(ctx, source, "Observation", local.SourceResourceID)
			require.NoError(t, err)
			require.Equal(t, string(local.ResourceRaw), detail.ResourceRawText)
		}
	})
	t.Run("competing writer at edit boundary", func(t *testing.T) {
		local, err := r.CreateLocalResource(ctx, source, json.RawMessage(`{"resourceType":"Observation","status":"final"}`))
		require.NoError(t, err)
		require.NoError(t, r.GormClient.Table("fhir_observation").Where("user_id = ? AND source_id = ? AND source_resource_id = ?", user.ID, source, local.SourceResourceID).Update("sort_title", "Synthetic local title").Error)
		type editContentionKey struct{}
		editCtx := context.WithValue(ctx, editContentionKey{}, true)
		const callback = "synthetic:edit-contention"
		var writerErr error
		var injected bool
		require.NoError(t, r.GormClient.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
			if tx.Statement.Context.Value(editContentionKey{}) == true && tx.Statement.Table == "fhir_observation" {
				injected = true
				writerErr = r.GormClient.WithContext(context.Background()).Model(&models.User{}).Where("id = ?", user.ID).Update("role", user.Role).Error
			}
		}))
		raw := json.RawMessage(`{"resourceType":"Observation","id":"` + local.SourceResourceID + `","status":"amended"}`)
		err = r.UpdateSharedResource(editCtx, source, "Observation", local.SourceResourceID, raw)
		require.NoError(t, r.GormClient.Callback().Update().Remove(callback))
		require.True(t, injected)
		require.NoError(t, writerErr)
		require.NoError(t, err, "a competing writer must not invalidate an edit read snapshot")
		stored, err := r.GetSharedResource(ctx, source, "Observation", local.SourceResourceID)
		require.NoError(t, err)
		require.Equal(t, local.ID, stored.ID)
		require.Equal(t, local.UserID, stored.UserID)
		require.Equal(t, local.SourceID, stored.SourceID)
		require.Equal(t, local.CreatedAt, stored.CreatedAt)
		require.Equal(t, "Synthetic local title", *stored.SortTitle)
	})
}
