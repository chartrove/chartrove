package web

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedEncryptedExactIDSearchUsesNativeAuthentication(t *testing.T) {
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-desktop-mode.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-desktop-mode-recovery-key"))
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
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, status == 200, response.Success)
		return response.Data
	}
	var token string
	require.NoError(t, json.Unmarshal(request("POST", "/api/auth/signup", "", `{"username":"synthetic-desktop-owner","password":"synthetic-password","join_mailing_list":false}`, 200), &token))
	var sources []struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(request("GET", "/api/secure/source", token, "", 200), &sources))
	require.NotEmpty(t, sources)
	base := "/api/secure/source-access/" + sources[0].ID
	var created struct {
		ID string `json:"source_resource_id"`
	}
	require.NoError(t, json.Unmarshal(request("POST", base+"/resources", token, `{"resource_raw":{"resourceType":"Observation","status":"final","code":{"text":"Synthetic desktop observation"}}}`, 200), &created))
	require.NotEmpty(t, created.ID)
	search := func(text, mode string, status int) []struct {
		ID string `json:"source_resource_id"`
	} {
		t.Helper()
		payload, _ := json.Marshal(map[string]interface{}{"text": text, "page": 0, "mode": mode})
		raw := request("POST", base+"/search/Observation", token, string(payload), status)
		var rows []struct {
			ID string `json:"source_resource_id"`
		}
		if status == 200 {
			require.NoError(t, json.Unmarshal(raw, &rows))
		}
		return rows
	}
	require.Len(t, search(created.ID, "exact_id", 200), 1)
	prefix := created.ID[:len(created.ID)-1]
	require.Len(t, search(prefix, "literal", 200), 1)
	require.Empty(t, search(prefix, "exact_id", 200))
	require.Empty(t, search(created.ID+" ", "exact_id", 200))
	search(created.ID, "token_phrase", 400)
	token = ""
	search(created.ID, "exact_id", 401)
	// A second native UUID cannot use a known owner/source/resource tuple.
	require.NoError(t, json.Unmarshal(request("POST", "/api/auth/signup", "", `{"username":"synthetic-desktop-other","password":"synthetic-password","join_mailing_list":false}`, 200), &token))
	search(created.ID, "exact_id", 404)
}
