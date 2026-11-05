package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedDeletedAccountRejectsStaleSession(t *testing.T) {
	app := openEmbedded(t, embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-stale-session.db"), false))
	response := embeddedRequest(t, app, http.MethodPost, "/api/auth/signup", `{"username":"synthetic-user","password":"synthetic-password","full_name":"Synthetic User"}`)
	require.Equal(t, 200, response.Code)
	var signup struct {
		Data string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &signup))
	require.NotEmpty(t, signup.Data)
	request := func(method, path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		r.Header.Set("Authorization", "Bearer "+signup.Data)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w
	}
	require.Equal(t, 200, request(http.MethodDelete, "/api/secure/account/me").Code)
	for _, path := range []string{"/api/secure/account/me", "/api/secure/source"} {
		require.Equal(t, 401, request(http.MethodGet, path).Code)
	}
}

func TestEmbeddedNativeIdentityReplacementDeniesPreviousSession(t *testing.T) {
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-identity-replacement.db"), false)
	app := openEmbedded(t, cfg)
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if body != "" {
			req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		return w
	}
	signup := func(password string) string {
		body, _ := json.Marshal(map[string]string{"username": "synthetic-reused-name", "password": password})
		w := request("POST", "/api/auth/signup", "", string(body))
		require.Equal(t, 200, w.Code)
		var result struct {
			Data string `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		return result.Data
	}
	old := signup("synthetic-first-password")
	oldClaims, err := auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), old)
	require.NoError(t, err)
	require.NotEmpty(t, oldClaims.NativeUserID)
	// A validly signed, unexpired old-format password JWT still needs a fresh
	// local login, even before deletion. Historical timestamps are not proof.
	legacyClaims := *oldClaims
	legacyClaims.NativeUserID = ""
	legacy, err := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyClaims).SignedString([]byte(cfg.GetString("jwt.issuer.key")))
	require.NoError(t, err)
	require.Equal(t, 401, request("GET", "/api/secure/account/me", legacy, "").Code)
	require.Equal(t, 200, request("GET", "/api/secure/account/me", old, "").Code)
	require.Equal(t, 200, request("DELETE", "/api/secure/account/me", old, "").Code)
	replacement := signup("synthetic-second-password")
	newClaims, err := auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), replacement)
	require.NoError(t, err)
	require.NotEqual(t, oldClaims.NativeUserID, newClaims.NativeUserID)
	require.Equal(t, 200, request("GET", "/api/secure/account/me", replacement, "").Code)
	for _, token := range []string{old, legacy} {
		for _, path := range []string{"/api/secure/account/me", "/api/secure/source"} {
			require.Equal(t, 401, request("GET", path, token, "").Code, "a previous native session must not authenticate a different account that reuses a username")
		}
	}
	require.Equal(t, 401, request("POST", "/api/auth/signin", "", `{"username":"synthetic-reused-name","password":"synthetic-first-password"}`).Code)
	login := request("POST", "/api/auth/signin", "", `{"username":"synthetic-reused-name","password":"synthetic-second-password"}`)
	require.Equal(t, 200, login.Code)
	var fresh struct {
		Data string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(login.Body.Bytes(), &fresh))
	require.Equal(t, 200, request("GET", "/api/secure/account/me", fresh.Data, "").Code)
}
