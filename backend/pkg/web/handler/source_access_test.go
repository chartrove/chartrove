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

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSourceAccessHTTPAuthorization(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-http-grants.db"))
	cfg.Set("database.encryption.enabled", false)
	db, err := database.NewRepository(cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	owner := &models.User{Username: "synthetic-http-owner", Password: "synthetic-password"}
	delegate := &models.User{Username: "synthetic-http-delegate", Password: "synthetic-password"}
	admin := &models.User{Username: "synthetic-http-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}
	for _, u := range []*models.User{owner, delegate, admin} {
		require.NoError(t, db.CreateUser(context.Background(), u))
	}
	ownerCtx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	sources, err := db.GetSources(ownerCtx)
	require.NoError(t, err)
	source := sources[0]
	actor := owner.Username
	router := gin.New()
	secure := router.Group("/api/secure", func(c *gin.Context) {
		// Production obtains this value from its existing authentication middleware.
		c.Set(pkg.ContextKeyTypeAuthUsername, actor)
		c.Set(pkg.ContextKeyTypeDatabase, db)
	})
	RegisterSourceAccessRoutes(secure)
	request := func(method, path, body string, status int) map[string]interface{} {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/api/secure/source-access"+path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, status, w.Code, w.Body.String())
		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		return response
	}
	base := "/" + source.ID.String()
	created := request(http.MethodPost, base+"/resources", `{"resource_raw":{"resourceType":"Observation","status":"final","code":{"text":"Synthetic"},"valueQuantity":{"value":0.1234567890123456789}}}`, 200)
	id := created["data"].(map[string]interface{})["source_resource_id"].(string)
	path := base + "/resources/Observation/" + id
	payload := `{"resource_raw":{"resourceType":"Observation","id":"` + id + `","status":"amended"}}`
	for _, name := range []string{delegate.Username, admin.Username} {
		actor = name
		request(http.MethodGet, base+"/summary", "", 404)
		request(http.MethodGet, base+"/resources/Observation", "", 404)
		request(http.MethodGet, path, "", 404)
		request(http.MethodPatch, path, payload, 404)
		request(http.MethodGet, base+"/grants", "", 404)
		request(http.MethodGet, base+"/audit", "", 404)
		request(http.MethodPost, base+"/grants", `{"delegate_username":"`+delegate.Username+`","access_level":"EDIT"}`, 404)
	}
	actor = owner.Username
	request(http.MethodPost, base+"/grants", `{"delegate_username":"`+delegate.Username+`","access_level":"ADMIN"}`, 400)
	request(http.MethodPost, base+"/grants", `{"delegate_username":"`+delegate.Username+`","access_level":"VIEW","owner_user_id":"`+admin.ID.String()+`"}`, 400)
	request(http.MethodPost, base+"/grants", `{"delegate_username":"`+delegate.Username+`","access_level":"VIEW"} {}`, 400)
	grant := request(http.MethodPost, base+"/grants", `{"delegate_username":"`+delegate.Username+`","access_level":"VIEW"}`, 200)
	grantID := grant["data"].(map[string]interface{})["id"].(string)
	actor = delegate.Username
	shared := request(http.MethodGet, "/shared", "", 200)
	encoded, err := json.Marshal(shared)
	require.NoError(t, err)
	for _, secret := range []string{"access_token", "refresh_token", "password", "dynamic_client_jwks"} {
		require.NotContains(t, string(encoded), secret)
	}
	request(http.MethodGet, base+"/summary", "", 200)
	request(http.MethodGet, base+"/resources/Observation", "", 200)
	detail := request(http.MethodGet, path, "", 200)
	require.Equal(t, false, detail["data"].(map[string]interface{})["editable"])
	require.Contains(t, detail["data"].(map[string]interface{})["resource_raw_text"], `"value":0.1234567890123456789`)
	request(http.MethodPatch, path, payload, 404)
	request(http.MethodPost, base+"/resources", `{"resource_raw":{"resourceType":"Observation"}}`, 404)
	request(http.MethodDelete, base+"/grants/"+grantID, "", 404)
	request(http.MethodGet, "/not-a-uuid/summary", "", 400)
	request(http.MethodGet, base+"/resources/Observation?page=invalid", "", 400)
	actor = owner.Username
	request(http.MethodPost, base+"/grants", `{"delegate_username":"`+delegate.Username+`","access_level":"EDIT"}`, 200)
	actor = delegate.Username
	detail = request(http.MethodGet, path, "", 200)
	require.Equal(t, true, detail["data"].(map[string]interface{})["editable"])
	request(http.MethodPatch, path, payload, 200)
	request(http.MethodPatch, path, `{"resource_raw":{"resourceType":"Observation","id":"changed"}}`, 400)
	request(http.MethodPatch, path, strings.Repeat(" ", 1024*1024+1), 400)
	actor = owner.Username
	request(http.MethodDelete, base+"/grants/"+grantID, "", 200)
	actor = delegate.Username
	request(http.MethodGet, path, "", 404)
	request(http.MethodPatch, path, payload, 404)
	request(http.MethodGet, base+"/summary", "", 404)
	request(http.MethodGet, base+"/resources/Observation", "", 404)
}
