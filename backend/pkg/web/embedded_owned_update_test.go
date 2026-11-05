package web

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedOwnedEditsPreserveNativeFamilyAuthorization(t *testing.T) {
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-owned-desktop.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-owned-desktop-recovery-key"))
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
		var response struct {
			Data json.RawMessage `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		return response.Data
	}
	signup := func(name string) string {
		t.Helper()
		var token string
		require.NoError(t, json.Unmarshal(request("POST", "/api/auth/signup", "", fmt.Sprintf(`{"username":%q,"password":"synthetic-password","join_mailing_list":false}`, name), 200), &token))
		return token
	}
	owner := signup("synthetic-native-family-owner")
	delegate := signup("synthetic-native-family-delegate")
	other := signup("synthetic-native-family-other")
	var sources []struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(request("GET", "/api/secure/source", owner, "", 200), &sources))
	require.NotEmpty(t, sources)
	base := "/api/secure/source-access/" + sources[0].ID
	var local struct {
		ID string `json:"source_resource_id"`
	}
	require.NoError(t, json.Unmarshal(request("POST", base+"/resources", owner, `{"resource_raw":{"resourceType":"Observation","status":"final","code":{"text":"Synthetic family record"}}}`, 200), &local))
	var grant struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(request("POST", base+"/grants", owner, `{"delegate_username":"synthetic-native-family-delegate","access_level":"EDIT"}`, 200), &grant))
	payload := fmt.Sprintf(`{"resource_raw":{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic changed family record"}}}`, local.ID)
	legacy := "/api/secure/resource/fhir/Observation/" + local.ID
	shared := base + "/resources/Observation/" + local.ID
	request("PATCH", legacy, "", payload, 401)
	request("GET", shared, other, "", 404)
	request("PATCH", legacy, owner, payload, 200)
	var detail struct {
		Editable bool `json:"editable"`
	}
	require.NoError(t, json.Unmarshal(request("GET", shared, delegate, "", 200), &detail))
	require.True(t, detail.Editable)
	request("PATCH", shared, delegate, payload, 200)
	request("DELETE", base+"/grants/"+grant.ID, owner, "", 200)
	request("PATCH", shared, delegate, payload, 404)
	request("GET", shared, delegate, "", 404)
}
