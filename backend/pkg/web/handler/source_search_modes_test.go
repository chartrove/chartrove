package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSourceSearchModeHTTP(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-mode-http.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-mode-http-key")
	r, err := database.NewRepository(cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, r.Close()) })
	users := []*models.User{{Username: "mode-http-owner", Password: "synthetic-password"}, {Username: "mode-http-delegate", Password: "synthetic-password"}, {Username: "mode-http-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}}
	for _, u := range users {
		require.NoError(t, r.CreateUser(context.Background(), u))
	}
	ownerCtx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, users[0].Username)
	ownerCtx = context.WithValue(ownerCtx, pkg.ContextKeyTypeAuthUserID, users[0].ID)
	sources, err := r.GetSources(ownerCtx)
	require.NoError(t, err)
	source := sources[0].ID
	table, err := dm.GetTableNameByResourceType("Observation")
	require.NoError(t, err)
	title := "Alpha report"
	for _, id := range []string{"Alpha", "Alpha-copy", "title-only"} {
		row := models.ResourceBase{OriginBase: models.OriginBase{UserID: users[0].ID, SourceID: source, SourceResourceType: "Observation", SourceResourceID: id}, SortTitle: &title}
		require.NoError(t, r.(*database.GormRepository).GormClient.Table(table).Create(&row).Error)
	}
	actor := users[0]
	router := gin.New()
	secure := router.Group("/secure", func(c *gin.Context) {
		c.Set(pkg.ContextKeyTypeAuthUsername, actor.Username)
		c.Set(pkg.ContextKeyTypeAuthUserID, actor.ID)
		c.Set(pkg.ContextKeyTypeDatabase, r)
	})
	RegisterSourceAccessRoutes(secure)
	request := func(body string, status int) []interface{} {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/secure/source-access/"+source.String()+"/search/Observation", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, status, w.Code, w.Body.String())
		require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		if status != 200 {
			require.Equal(t, false, response["success"])
			return nil
		}
		require.Equal(t, true, response["success"])
		rows := response["data"].([]interface{})
		for _, row := range rows {
			value := row.(map[string]interface{})
			require.NotContains(t, value, "user_id")
			require.NotContains(t, value, "resource_raw")
		}
		return rows
	}
	legacy := request(`{"text":"Alpha","page":0}`, 200)
	require.Len(t, legacy, 3)
	require.Equal(t, legacy, request(`{"text":"Alpha","page":0,"mode":"literal"}`, 200))
	exact := `{"text":"Alpha","page":0,"mode":"exact_id"}`
	rows := request(exact, 200)
	require.Len(t, rows, 1)
	require.Equal(t, "Alpha", rows[0].(map[string]interface{})["source_resource_id"])
	for _, text := range []string{"alpha", "Alpha ", " Alpha", "%", "Alpha report"} {
		body, _ := json.Marshal(map[string]interface{}{"mode": "exact_id", "text": text, "page": 0})
		require.Empty(t, request(string(body), 200))
	}
	require.Empty(t, request(`{"text":"Alpha","page":1,"mode":"exact_id"}`, 200))
	for _, body := range []string{`{"text":"Alpha","mode":"token_phrase"}`, `{"text":"Alpha","mode":"all_tokens"}`, `{"text":"Alpha","mode":"unknown"}`, `{"text":"Alpha","mode":true}`, `{"text":"Alpha","mode":null}`, `{"text":"Alpha","mode":""}`, `{"text":"Alpha","mode":"exact_id","owner_user_id":"untrusted"}`, `{"text":"Alpha","mode":"exact_id","page":-1}`, `{"text":"Alpha","mode":"exact_id","page":100001}`, `{"text":"","mode":"exact_id"}`, `{"text":"\u0000","mode":"exact_id"}`} {
		request(body, 400)
	}
	body, _ := json.Marshal(map[string]interface{}{"text": strings.Repeat("x", 129), "mode": "exact_id"})
	request(string(body), 400)
	request(`{"text":"Alpha","mode":"exact_id","padding":"`+strings.Repeat("x", 4096)+`"}`, 400)
	for _, u := range users[1:] {
		actor = u
		request(exact, 404)
	}
	grant, err := r.(database.SourceAccessRepository).SetSourceGrant(ownerCtx, source, users[1].Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	actor = users[1]
	require.Len(t, request(exact, 200), 1)
	require.NoError(t, r.(database.SourceAccessRepository).RevokeSourceGrant(ownerCtx, source, grant.ID))
	request(exact, 404)
	grant, err = r.(database.SourceAccessRepository).SetSourceGrant(ownerCtx, source, users[1].Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	require.Len(t, request(exact, 200), 1)
	require.NoError(t, r.(*database.GormRepository).GormClient.Model(&models.SourceGrant{}).Where("id = ?", grant.ID).Update("expires_at", time.Now().UTC().Add(-time.Minute)).Error)
	request(exact, 404)
}
