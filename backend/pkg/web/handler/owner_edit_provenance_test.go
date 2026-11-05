package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestOwnerHTTPUpdatePreservesFamilyEditProvenance(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-owner-edits.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-owner-edit-key")
	db, err := database.NewRepository(cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	owner := &models.User{Username: "synthetic-edit-owner", Password: "synthetic-password"}
	delegate := &models.User{Username: "synthetic-edit-family", Password: "synthetic-password"}
	require.NoError(t, db.CreateUser(context.Background(), owner))
	require.NoError(t, db.CreateUser(context.Background(), delegate))
	ctx := func(u *models.User) context.Context {
		c := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username)
		return context.WithValue(c, pkg.ContextKeyTypeAuthUserID, u.ID)
	}
	sources, err := db.GetSources(ctx(owner))
	require.NoError(t, err)
	source := sources[0].ID
	access := db.(database.SourceAccessRepository)
	local, err := access.CreateLocalResource(ctx(owner), source, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic family observation"},"valueQuantity":{"value":1.00}}`))
	require.NoError(t, err)
	_, err = access.SetSourceGrant(ctx(owner), source, delegate.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	before, err := access.GetSharedResource(ctx(delegate), source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.True(t, before.Editable, "positive delegated-edit precondition")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(pkg.ContextKeyTypeDatabase, db)
		c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logrus.New()))
		c.Set(pkg.ContextKeyTypeAuthUsername, owner.Username)
		c.Set(pkg.ContextKeyTypeAuthUserID, owner.ID)
	})
	router.PATCH("/resource/fhir/:resourceType/:resourceId", UpdateResourceFhir)
	raw := fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic family observation"},"valueQuantity":{"value":2.00}}`, local.SourceResourceID)
	body := `{"resource_raw":` + raw + `}`
	patch := func(body string, status int) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/resource/fhir/Observation/"+local.SourceResourceID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, status, w.Code, w.Body.String())
	}
	patch(body, http.StatusOK)
	after, err := access.GetSharedResource(ctx(delegate), source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.True(t, after.Editable, "owner edits must not revoke existing local-edit eligibility")
	require.Equal(t, before.ID, after.ID)
	require.Equal(t, before.CreatedAt, after.CreatedAt)
	require.Equal(t, before.SourceID, after.SourceID)
	require.Contains(t, string(after.ResourceRaw), `"value":2.00`)
	require.NoError(t, access.UpdateSharedResource(ctx(delegate), source, "Observation", local.SourceResourceID, json.RawMessage(raw)))

	patch(`{"resource_raw":`+raw+`,"sort_title":"Synthetic dated record","sort_date":"2020-01-02"}`, 200)
	patch(body, 200)
	dated, err := access.GetSharedResource(ctx(owner), source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.Equal(t, "Synthetic dated record", *dated.SortTitle)
	require.True(t, dated.SortDate.Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)))
	for _, invalid := range []string{`"sort_date":25`, `"sort_title":false`} {
		patch(`{"resource_raw":`+raw+`,`+invalid+`}`, 400)
	}
	patch(`{"resource_raw":`+raw+`,"sort_title":null,"sort_date":null}`, 200)
	cleared, err := access.GetSharedResource(ctx(owner), source, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.Empty(t, *cleared.SortTitle)
	require.Nil(t, cleared.SortDate)
}
