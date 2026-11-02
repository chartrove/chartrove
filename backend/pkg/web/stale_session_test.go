package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

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
