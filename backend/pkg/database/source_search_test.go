package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm/logger"
)

func TestSourceSearchIsolationMetadataAndRevocation(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-source-search.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-search-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	users := []*models.User{{Username: "synthetic-search-owner", Password: "synthetic-password"}, {Username: "synthetic-search-delegate", Password: "synthetic-password"}, {Username: "synthetic-search-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}}
	contexts := []context.Context{}
	for _, u := range users {
		require.NoError(t, r.CreateUser(context.Background(), u))
		contexts = append(contexts, context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username))
	}
	sources, err := r.GetSources(contexts[0])
	require.NoError(t, err)
	source := sources[0].ID
	other, err := r.GetSources(contexts[1])
	require.NoError(t, err)
	table, err := dm.GetTableNameByResourceType("Observation")
	require.NoError(t, err)
	rows := []models.ResourceBase{}
	for i := 0; i < 52; i++ {
		title := "Synthetic clinic visit"
		if i == 0 {
			title = "Synthetic clinic 100% follow_up!"
		}
		if i == 1 {
			title = "Synthetic clinic 100units followXup!"
		}
		if i == 2 {
			title = "Synthetic clinic Écho"
		}
		rows = append(rows, models.ResourceBase{OriginBase: models.OriginBase{UserID: users[0].ID, SourceID: source, SourceResourceType: "Observation", SourceResourceID: fmt.Sprintf("synthetic-search-%03d", i)}, SortTitle: &title, ResourceRaw: datatypes.JSON(`{"resourceType":"Observation","status":"final","note":[{"text":"Synthetic private payload not in a search hit"}]}`)})
	}
	require.NoError(t, r.GormClient.Table(table).CreateInBatches(&rows, 20).Error)
	foreign := rows[0]
	foreign.UserID = users[1].ID
	foreign.SourceID = other[0].ID
	require.NoError(t, r.GormClient.Table(table).Create(&foreign).Error)
	for _, ctx := range contexts[1:] {
		_, err := r.SearchSourceResources(ctx, source, "Observation", "clinic", 0)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
	}
	grant, err := r.SetSourceGrant(contexts[0], source, users[1].Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	first, err := r.SearchSourceResources(contexts[1], source, "Observation", "CLINIC", 0)
	require.NoError(t, err)
	require.Len(t, first, 50)
	second, err := r.SearchSourceResources(contexts[1], source, "Observation", "clinic", 1)
	require.NoError(t, err)
	require.Len(t, second, 2)
	require.NotEqual(t, first[49].SourceResourceID, second[0].SourceResourceID)
	for _, hit := range append(first, second...) {
		require.Equal(t, source, hit.SourceID)
	}
	raw, err := json.Marshal(first)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "resource_raw")
	require.NotContains(t, string(raw), "private payload")
	require.NotContains(t, string(raw), "user_id")
	for _, literal := range []string{"100%", "follow_up!", "_", "Écho", "synthetic-search-000"} {
		found, err := r.SearchSourceResources(contexts[1], source, "Observation", literal, 0)
		require.NoError(t, err)
		require.Len(t, found, 1)
	}
	privateOnly, err := r.SearchSourceResources(contexts[0], source, "Observation", "private payload", 0)
	require.NoError(t, err)
	require.Empty(t, privateOnly)
	for _, text := range []string{"", "   ", strings.Repeat("x", 129), string([]byte{0xff})} {
		_, err = r.SearchSourceResources(contexts[0], source, "Observation", text, 0)
		require.ErrorIs(t, err, ErrSourceAccessInput)
	}
	for _, page := range []int{-1, 100001} {
		_, err = r.SearchSourceResources(contexts[0], source, "Observation", "clinic", page)
		require.ErrorIs(t, err, ErrSourceAccessInput)
	}
	_, err = r.SearchSourceResources(contexts[0], source, "UnknownResource", "clinic", 0)
	require.ErrorIs(t, err, ErrSourceAccessInput)
	// Permissions are not embedded in hits. Revoke/expire then deny both the
	// next search and opening a previously returned hit through the existing API.
	require.NoError(t, r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", grant.ID).Update("expires_at", time.Now().UTC().Add(-time.Minute)).Error)
	_, err = r.SearchSourceResources(contexts[1], source, "Observation", "clinic", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	grant, err = r.SetSourceGrant(contexts[0], source, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	_, err = r.SearchSourceResources(contexts[1], source, "Observation", "clinic", 0)
	require.NoError(t, err)
	require.NoError(t, r.RevokeSourceGrant(contexts[0], source, grant.ID))
	_, err = r.SearchSourceResources(contexts[1], source, "Observation", "clinic", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	_, err = r.GetSharedResource(contexts[1], source, "Observation", first[0].SourceResourceID)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.NoError(t, r.GormClient.Table(table).Where("user_id = ? AND source_id = ? AND source_resource_id = ?", users[0].ID, source, rows[0].SourceResourceID).Update("deleted_at", time.Now().UTC()).Error)
	found, err := r.SearchSourceResources(contexts[0], source, "Observation", "100%", 0)
	require.NoError(t, err)
	require.Empty(t, found)
	foreignHits, err := r.SearchSourceResources(contexts[1], other[0].ID, "Observation", "100%", 0)
	require.NoError(t, err)
	require.Len(t, foreignHits, 1)
	var trace bytes.Buffer
	r.GormClient.Logger = logger.New(log.New(&trace, "", 0), logger.Config{LogLevel: logger.Info, Colorful: false})
	_, err = r.SearchSourceResources(contexts[0], source, "Observation", "synthetic-private-search-term", 0)
	require.NoError(t, err)
	require.Contains(t, trace.String(), "LIKE LOWER")
	require.NotContains(t, trace.String(), "synthetic-private-search-term")
}
