package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestStandbyEncryptionKeyResponsesPreventCaching(t *testing.T) {
	for _, basepath := range []string{"", "/synthetic-prefix"} {
		t.Run("basepath="+basepath, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", "synthetic-encryption-cache-signing-key")
			cfg.Set("web.listen.basepath", basepath)
			cfg.Set("web.allow_unsafe_endpoints", false)
			databasePath := filepath.Join(t.TempDir(), "must-not-be-opened.db")
			cfg.Set("database.location", databasePath)
			cfg.Set("database.encryption.enabled", true)
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			engine := &AppEngine{Config: cfg, Logger: logrus.NewEntry(logger), EventBus: event_bus.NewNoopEventBusServer(), StandbyMode: true}
			_, router := engine.Setup()
			for _, test := range []struct {
				name, method, path, body string
				status                   int
			}{
				{"generate", http.MethodGet, "/api/encryption-key", "", http.StatusOK},
				{"set-invalid", http.MethodPost, "/api/encryption-key", "{", http.StatusBadRequest},
				{"set-missing", http.MethodPost, "/api/encryption-key", "{}", http.StatusBadRequest},
				{"validate-invalid", http.MethodPost, "/api/encryption-key/validate", "{", http.StatusBadRequest},
				{"validate-missing", http.MethodPost, "/api/encryption-key/validate", "{}", http.StatusBadRequest},
			} {
				t.Run(test.name, func(t *testing.T) {
					request := httptest.NewRequest(test.method, basepath+test.path, bytes.NewBufferString(test.body))
					request.Header.Set("Content-Type", "application/json")
					response := httptest.NewRecorder()
					router.ServeHTTP(response, request)
					require.Equal(t, test.status, response.Code) // Never dump key-bearing response bodies.
					require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
					if test.method == http.MethodGet {
						var envelope struct {
							Success bool   `json:"success"`
							Data    string `json:"data"`
						}
						if json.Unmarshal(response.Body.Bytes(), &envelope) != nil || !envelope.Success || envelope.Data == "" {
							t.Fatal("generation response contract unavailable")
						}
					}
				})
			}
			if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
				t.Fatal("standby cache fixture unexpectedly opened a database")
			}
		})
	}
}

func TestEmbeddedEncryptionKeyResponseCachePreservesUnlockLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic-encryption-cache.db")
	key := "synthetic-cache-fixture-key-only"
	seed := openEmbedded(t, embeddedConfig(t, path, true))
	require.NoError(t, seed.UnlockDatabase(key))
	require.NoError(t, seed.Close())
	app := openEmbedded(t, embeddedConfig(t, path, true))
	assertLocked := func(expected bool) {
		response := embeddedRequest(t, app, http.MethodGet, "/api/health", "")
		require.Equal(t, http.StatusOK, response.Code)
		var envelope struct {
			Data struct {
				Standby bool `json:"standby_mode"`
			} `json:"data"`
		}
		if json.Unmarshal(response.Body.Bytes(), &envelope) != nil || envelope.Data.Standby != expected {
			t.Fatal("encrypted fixture lock-state contract unavailable")
		}
	}
	assertLocked(true)
	for _, test := range []struct {
		name, route, key string
		status           int
		locked           bool
	}{
		{"validate-correct", "/api/encryption-key/validate", key, http.StatusOK, true},
		{"validate-wrong", "/api/encryption-key/validate", "wrong-synthetic-cache-key", http.StatusBadRequest, true},
		{"unlock-wrong", "/api/encryption-key", "wrong-synthetic-cache-key", http.StatusBadRequest, true},
		{"unlock-correct", "/api/encryption-key", key, http.StatusOK, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"encryption_key": test.key})
			require.NoError(t, err)
			response := embeddedRequest(t, app, http.MethodPost, test.route, string(body))
			require.Equal(t, test.status, response.Code) // Do not print key-bearing request/response bodies.
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			assertLocked(test.locked)
		})
	}
	response := embeddedRequest(t, app, http.MethodGet, "/api/encryption-key", "")
	require.Equal(t, http.StatusNotFound, response.Code, "unlocked application must not expose standby key generation")
}
