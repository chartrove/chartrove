package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedMutationsKeepDuplicateIDsInTheirNativeSource(t *testing.T) {
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-scoped-desktop.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-scoped-desktop-key"))
	request := func(method, path, token, body string, status int) json.RawMessage {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		require.Equal(t, status, w.Code)
		var result struct {
			Data json.RawMessage `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		return result.Data
	}
	signup := func(name string) string {
		var token string
		require.NoError(t, json.Unmarshal(request("POST", "/api/auth/signup", "", fmt.Sprintf(`{"username":%q,"password":"synthetic-password","join_mailing_list":false}`, name), 200), &token))
		return token
	}
	owner := signup("synthetic-scope-native-owner")
	other := signup("synthetic-scope-native-other")
	db := app.engine.deviceRepo
	user, err := db.GetUserByUsername(context.Background(), "synthetic-scope-native-owner")
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
	ctx = context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, user.ID)
	sources := []*models.SourceCredential{{Patient: "synthetic-native-source-A"}, {Patient: "synthetic-native-source-B"}}
	id := "synthetic-native-collision"
	for _, source := range sources {
		require.NoError(t, db.CreateSource(ctx, source))
		_, err := db.UpsertRawResource(ctx, source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: id, ResourceRaw: json.RawMessage(fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"final","code":{"text":"Synthetic original source"}}`, id))})
		require.NoError(t, err)
	}
	base := "/api/secure/resource/fhir/Observation/" + id
	payload := fmt.Sprintf(`{"resource_raw":{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic selected source edit"}}}`, id)
	request("PATCH", base, "", payload, 401)
	request("PATCH", base, owner, payload, 409)
	scoped := strings.TrimSuffix(payload, "}") + fmt.Sprintf(`,"source_id":%q}`, sources[1].ID)
	request("PATCH", base, other, scoped, 404)
	request("PATCH", base, owner, scoped, 200)
	request("DELETE", base, owner, "", 409)
	query := "?source_id=" + sources[1].ID.String()
	request("DELETE", base+query, "", "", 401)
	request("DELETE", base+query, other, "", 404)
	request("DELETE", base+query, owner, "", 200)
	request("GET", "/api/secure/source-access/"+sources[1].ID.String()+"/resources/Observation/"+id, owner, "", 404)
	remaining := request("GET", "/api/secure/source-access/"+sources[0].ID.String()+"/resources/Observation/"+id, owner, "", 200)
	require.Contains(t, string(remaining), "Synthetic original source")
	request("DELETE", base, owner, "", 200) // Backward-compatible unique legacy target.
}
