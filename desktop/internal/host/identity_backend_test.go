package host

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLocalIdentityResolverUsesRealEmbeddedAuthentication(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-native-owner.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-native-identity-key")
	cfg.Set("jwt.issuer.key", "synthetic-native-identity-signing-key")
	cfg.Set("web.listen.basepath", "")
	cfg.Set("web.allow_unsafe_endpoints", false)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	entry := logrus.NewEntry(logger)
	backend, err := web.NewEmbeddedApplication(cfg, entry, event_bus.NewNoopEventBusServer(), http.NotFoundHandler())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, backend.Close()) })
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		backend.ServeHTTP(response, request)
		return response
	}
	signup := func() string {
		response := call(http.MethodPost, "/api/auth/signup", "", `{"username":"synthetic-native-owner","password":"synthetic-native-password"}`)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var envelope struct {
			Success bool   `json:"success"`
			Data    string `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		require.True(t, envelope.Success)
		return envelope.Data
	}
	token := signup()
	resolve := NewLocalIdentityResolver(backend)
	owner, err := resolve(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "synthetic-native-owner", owner.Username)
	_, err = resolve(context.Background(), token+"synthetic-invalid-signature")
	require.Error(t, err)
	deleted := call(http.MethodDelete, "/api/secure/account/me", token, "")
	require.Equal(t, http.StatusOK, deleted.Code, deleted.Body.String())
	_, err = resolve(context.Background(), token)
	require.Error(t, err)
	fresh := signup()
	replacement, err := resolve(context.Background(), fresh)
	require.NoError(t, err)
	require.Equal(t, owner.Username, replacement.Username)
	require.NotEqual(t, owner.ID, replacement.ID)
	_, err = resolve(context.Background(), token)
	require.Error(t, err, "username recreation cannot revive the previous native owner")
}
