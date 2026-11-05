package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedRegistrationPolicyAndLocalRecovery(t *testing.T) {
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-registration.db"), false)
	cfg.Set("auth.registration.enabled", false)
	app := openEmbedded(t, cfg)
	status := embeddedRequest(t, app, "GET", "/api/auth/registration", "")
	require.Equal(t, 200, status.Code, status.Body.String())
	require.Equal(t, "no-store", status.Header().Get("Cache-Control"))
	require.JSONEq(t, `{"success":true,"data":{"enabled":false,"allowed":true,"bootstrap":true}}`, status.Body.String())
	credentials := `{"username":"synthetic-local-admin","password":"synthetic-password","role":"user"}`
	signup := embeddedRequest(t, app, "POST", "/api/auth/signup", credentials)
	require.Equal(t, 200, signup.Code, signup.Body.String())
	var result struct {
		Data string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(signup.Body.Bytes(), &result))
	claims, err := auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), result.Data)
	require.NoError(t, err)
	require.Equal(t, pkg.UserRoleAdmin, claims.Role)
	adminToken := result.Data
	status = embeddedRequest(t, app, "GET", "/api/auth/registration", "")
	require.JSONEq(t, `{"success":true,"data":{"enabled":false,"allowed":false,"bootstrap":false}}`, status.Body.String())
	denied := embeddedRequest(t, app, "POST", "/api/auth/signup", `{"username":"synthetic-rejected","password":"synthetic-password","role":"admin","registration_enabled":true}`)
	require.Equal(t, 403, denied.Code, denied.Body.String())
	require.Contains(t, denied.Body.String(), "registration_closed")
	require.Equal(t, 401, embeddedRequest(t, app, "POST", "/api/auth/signin", `{"username":"synthetic-rejected","password":"synthetic-password"}`).Code)
	require.Equal(t, 200, embeddedRequest(t, app, "POST", "/api/auth/signin", credentials).Code)
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w
	}
	managed := request("POST", "/api/secure/users", adminToken, `{"username":"synthetic-managed-user","password":"synthetic-password","role":"user"}`)
	require.Equal(t, http.StatusOK, managed.Code, managed.Body.String())
	managedLogin := embeddedRequest(t, app, "POST", "/api/auth/signin", `{"username":"synthetic-managed-user","password":"synthetic-password"}`)
	require.Equal(t, 200, managedLogin.Code, managedLogin.Body.String())
	require.NoError(t, json.Unmarshal(managedLogin.Body.Bytes(), &result))
	managedToken := result.Data
	var ownSources struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	listed := request("GET", "/api/secure/source", managedToken, "")
	require.Equal(t, 200, listed.Code)
	require.NoError(t, json.Unmarshal(listed.Body.Bytes(), &ownSources))
	require.Len(t, ownSources.Data, 1)
	base := "/api/secure/source-access/" + ownSources.Data[0].ID + "/resources"
	private := request("POST", base, managedToken, `{"resource_raw":{"resourceType":"Observation","status":"final","code":{"text":"Synthetic private registration record"}}}`)
	require.Equal(t, 200, private.Code, private.Body.String())
	var record struct {
		Data struct {
			ID string `json:"source_resource_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(private.Body.Bytes(), &record))
	detail := base + "/Observation/" + record.Data.ID
	require.Equal(t, 200, request("GET", detail, managedToken, "").Code)
	require.Equal(t, 404, request("GET", detail, adminToken, "").Code)
	require.Equal(t, 401, request("POST", "/api/secure/users", managedToken, `{"username":"synthetic-escalation","password":"synthetic-password","role":"admin"}`).Code)
	// Re-enabling registration preserves the original ordinary-user contract.
	cfg.Set("auth.registration.enabled", true)
	ordinary := embeddedRequest(t, app, "POST", "/api/auth/signup", `{"username":"synthetic-ordinary","password":"synthetic-password","role":"admin"}`)
	require.Equal(t, 200, ordinary.Code, ordinary.Body.String())
	require.NoError(t, json.Unmarshal(ordinary.Body.Bytes(), &result))
	claims, err = auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), result.Data)
	require.NoError(t, err)
	require.Equal(t, pkg.UserRoleUser, claims.Role)
	cfg.Set("auth.registration.enabled", false)
	require.NoError(t, app.Close())
	reopened := openEmbedded(t, cfg)
	require.Equal(t, 200, embeddedRequest(t, reopened, "POST", "/api/auth/signin", credentials).Code)
	require.Equal(t, 200, embeddedRequest(t, reopened, "POST", "/api/auth/signin", `{"username":"synthetic-managed-user","password":"synthetic-password"}`).Code)
	require.Equal(t, 403, embeddedRequest(t, reopened, "POST", "/api/auth/signup", `{"username":"synthetic-still-closed","password":"synthetic-password"}`).Code)
}
