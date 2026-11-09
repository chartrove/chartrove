package shell

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
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web"
	"github.com/karamhouse/fasten-desktop/internal/host"
	"github.com/karamhouse/fasten-desktop/internal/smart"
	"github.com/karamhouse/fasten-desktop/internal/smarttest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// This is an in-process native-service/API flow, not a Wails GUI acceptance test.
// Only window context is synthetic; PKCE/TLS, local signup/JWT authority, reader,
// real encrypted storage, API original downloads and database restart are real.
func TestNativeClinicalFlowEncryptedRestart(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-native-clinical.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-native-clinical-key")
	cfg.Set("jwt.issuer.key", "synthetic-native-clinical-signing-key")
	cfg.Set("web.listen.basepath", "")
	cfg.Set("web.allow_unsafe_endpoints", false)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	open := func() *web.EmbeddedApplication {
		app, err := web.NewEmbeddedApplication(cfg, logrus.NewEntry(logger), event_bus.NewNoopEventBusServer(), http.NotFoundHandler())
		require.NoError(t, err)
		return app
	}
	backend := open()
	t.Cleanup(func() { require.NoError(t, backend.Close()) })
	call := func(method, path, token string, body []byte) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		backend.ServeHTTP(response, request)
		return response
	}
	signup := func(username string) string {
		body, err := json.Marshal(map[string]string{"username": username, "password": "synthetic-native-password"})
		require.NoError(t, err)
		response := call(http.MethodPost, "/api/auth/signup", "", body)
		require.Equal(t, http.StatusOK, response.Code)
		var result struct {
			Success bool   `json:"success"`
			Data    string `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
		require.True(t, result.Success)
		return result.Data
	}
	token, foreign := signup("synthetic-native-owner"), signup("synthetic-native-foreign")
	fixture, err := smarttest.New()
	require.NoError(t, err)
	t.Cleanup(fixture.Close)
	manager, err := smart.NewManagerWithHTTPClient(fixture.Profiles(), fixture.OpenBrowser, fixture.Client())
	require.NoError(t, err)
	t.Cleanup(manager.Close)
	service := NewSMARTService(manager, host.NewLocalIdentityResolver(backend))
	ctx := smartTestContext("main")
	profiles := fixture.Profiles()
	connection, err := service.Connect(ctx, profiles[0].ID, token)
	require.NoError(t, err)
	foreignRaw, err := service.ReadClinicalSnapshot(ctx, connection.ID, foreign)
	require.Error(t, err)
	require.Empty(t, foreignRaw)
	raw, err := service.ReadClinicalSnapshot(ctx, connection.ID, token)
	require.NoError(t, err)
	require.Contains(t, raw, smarttest.PatientJSON)
	require.Contains(t, raw, "3.500000000000000001")
	post := func(raw, token, owner, base string) *httptest.ResponseRecorder {
		body, err := json.Marshal(map[string]interface{}{"snapshot_json": raw, "confirmed": true, "expected_owner_id": owner, "origin": map[string]string{"fhir_base_url": base, "display": "Synthetic native clinical"}})
		require.NoError(t, err)
		return call(http.MethodPost, "/api/secure/source/clinical-snapshot", token, body)
	}
	result := func(response *httptest.ResponseRecorder) models.ClinicalSnapshotImportResult {
		require.Equal(t, http.StatusOK, response.Code)
		var envelope struct {
			Success bool                                `json:"success"`
			Data    models.ClinicalSnapshotImportResult `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		require.True(t, envelope.Success)
		return envelope.Data
	}
	require.Equal(t, http.StatusForbidden, post(raw, foreign, connection.Owner.ID.String(), profiles[0].FHIRBaseURL).Code)
	first := result(post(raw, token, connection.Owner.ID.String(), profiles[0].FHIRBaseURL))
	require.Equal(t, 3, first.Created)
	repeated, err := service.ReadClinicalSnapshot(ctx, connection.ID, token)
	require.NoError(t, err)
	require.Equal(t, raw, repeated)
	replay := result(post(repeated, token, connection.Owner.ID.String(), profiles[0].FHIRBaseURL))
	require.True(t, replay.AlreadyImported)
	changed, err := service.ReadClinicalSnapshot(ctx, connection.ID, token)
	require.NoError(t, err)
	require.Contains(t, changed, "4.250000000000000003")
	updated := result(post(changed, token, connection.Owner.ID.String(), profiles[0].FHIRBaseURL))
	require.Equal(t, first.SourceID, updated.SourceID)
	require.Equal(t, 1, updated.Updated)
	require.Equal(t, 2, updated.Unchanged)
	require.NotEqual(t, first.SnapshotID, updated.SnapshotID)
	historical := result(post(raw, token, connection.Owner.ID.String(), profiles[0].FHIRBaseURL))
	require.True(t, historical.AlreadyImported)
	require.Equal(t, first.SnapshotID, historical.SnapshotID)
	require.NoError(t, service.Disconnect(ctx, connection.ID))
	otherOrigin, err := service.Connect(ctx, profiles[1].ID, token)
	require.NoError(t, err)
	otherRaw, err := service.ReadClinicalSnapshot(ctx, otherOrigin.ID, token)
	require.NoError(t, err)
	other := result(post(otherRaw, token, otherOrigin.Owner.ID.String(), profiles[1].FHIRBaseURL))
	require.NotEqual(t, first.SourceID, other.SourceID, "same Patient ID from distinct origins must not merge")
	require.NoError(t, service.Disconnect(ctx, otherOrigin.ID))
	require.NoError(t, backend.Close())
	backend = open()
	owner, err := host.NewLocalIdentityResolver(backend)(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, connection.Owner.ID, owner.ID)
	for _, item := range []struct {
		result models.ClinicalSnapshotImportResult
		raw    string
	}{{first, raw}, {updated, changed}, {other, otherRaw}} {
		path := "/api/secure/source/" + item.result.SourceID.String() + "/clinical-snapshots/" + item.result.SnapshotID.String() + "/original"
		download := call(http.MethodGet, path, token, nil)
		require.Equal(t, http.StatusOK, download.Code)
		require.Equal(t, item.raw, download.Body.String())
		require.Equal(t, http.StatusForbidden, call(http.MethodGet, path, foreign, nil).Code)
	}
}
