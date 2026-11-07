package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedManualImportKeepsPatientSourceAndGrantIdentity(t *testing.T) {
	blocked := &blockedTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv("TMPDIR", t.TempDir())
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-patient-identity.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-manual-patient-key"))
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w
	}
	login := func(path, name string) string {
		t.Helper()
		w := request(http.MethodPost, path, "", `{"username":"`+name+`","password":"synthetic-password","join_mailing_list":false}`)
		require.Equal(t, http.StatusOK, w.Code)
		var response struct{ Data string }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.NotEmpty(t, response.Data)
		return response.Data
	}
	owner := login("/api/auth/signup", "synthetic-patient-owner")
	delegate := login("/api/auth/signup", "synthetic-patient-delegate")
	importPatient := func(patient string, value int) models.SourceCredential {
		t.Helper()
		bundle := fmt.Sprintf(`{"resourceType":"Bundle","type":"collection","entry":[{"resource":{"resourceType":"Patient","id":%q,"name":[{"family":"Synthetic","given":["Identity"]}],"gender":"unknown"}},{"resource":{"resourceType":"Observation","id":"synthetic-shared-observation","status":"final","code":{"text":"Synthetic patient identity"},"subject":{"reference":"Patient/%s"},"effectiveDateTime":"2020-01-02","valueInteger":%d,"note":[{"text":"Retain café صفر"}]}}]}`, patient, patient, value)
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		file, err := form.CreateFormFile("file", "synthetic-patient.json")
		require.NoError(t, err)
		_, err = io.WriteString(file, bundle)
		require.NoError(t, err)
		require.NoError(t, form.Close())
		r := httptest.NewRequest(http.MethodPost, "/api/secure/source/manual", &body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("Authorization", "Bearer "+owner)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
		var response struct{ Source models.SourceCredential }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.NotEqual(t, uuid.Nil, response.Source.ID)
		return response.Source
	}
	readObservation := func(source models.SourceCredential, token string) models.SharedResource {
		t.Helper()
		w := request(http.MethodGet, "/api/secure/source-access/"+source.ID.String()+"/resources/Observation/synthetic-shared-observation", token, "")
		require.Equal(t, http.StatusOK, w.Code)
		var response struct{ Data models.SharedResource }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, source.ID, response.Data.SourceID)
		require.NotEmpty(t, response.Data.ResourceRaw)
		return response.Data
	}

	// Both are valid, distinct FHIR IDs. Leading letters are identity, not a cutset.
	first := importPatient("aa001", 1)
	assert.Equal(t, "aa001", first.Patient)
	before := readObservation(first, owner)
	grant := request(http.MethodPost, "/api/secure/source-access/"+first.ID.String()+"/grants", owner, `{"delegate_username":"synthetic-patient-delegate","access_level":"VIEW"}`)
	require.Equal(t, http.StatusOK, grant.Code)
	var grantResponse struct{ Data models.SourceGrant }
	require.NoError(t, json.Unmarshal(grant.Body.Bytes(), &grantResponse))
	require.NotEqual(t, uuid.Nil, grantResponse.Data.ID)
	assert.JSONEq(t, string(before.ResourceRaw), string(readObservation(first, delegate).ResourceRaw))

	second := importPatient("001", 2)
	assert.Equal(t, "001", second.Patient)
	assert.NotEqual(t, first.ID, second.ID, "different patients must not share one manual-source/grant identity")
	assert.JSONEq(t, string(before.ResourceRaw), string(readObservation(first, owner).ResourceRaw), "a later patient's import must not overwrite an earlier patient's observation")
	assert.JSONEq(t, string(before.ResourceRaw), string(readObservation(first, delegate).ResourceRaw), "an existing source grant must not expose a later patient's clinical value")
	denied := request(http.MethodGet, "/api/secure/source-access/"+second.ID.String()+"/resources/Patient/001", delegate, "")
	assert.Equal(t, http.StatusNotFound, denied.Code, "the second patient's source was never shared")
	assert.Equal(t, http.StatusUnauthorized, request(http.MethodGet, "/api/secure/source-access/"+second.ID.String()+"/resources/Patient/001", "", "").Code)

	revoked := request(http.MethodDelete, "/api/secure/source-access/"+first.ID.String()+"/grants/"+grantResponse.Data.ID.String(), owner, "")
	require.Equal(t, http.StatusOK, revoked.Code)
	assert.Equal(t, http.StatusNotFound, request(http.MethodGet, "/api/secure/source-access/"+first.ID.String()+"/resources/Observation/synthetic-shared-observation", delegate, "").Code)
	require.NoError(t, app.Close())
	app = openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-manual-patient-key"))
	owner = login("/api/auth/signin", "synthetic-patient-owner")
	delegate = login("/api/auth/signin", "synthetic-patient-delegate")
	assert.JSONEq(t, string(before.ResourceRaw), string(readObservation(first, owner).ResourceRaw))
	assert.Equal(t, http.StatusNotFound, request(http.MethodGet, "/api/secure/source-access/"+first.ID.String()+"/resources/Observation/synthetic-shared-observation", delegate, "").Code)
	require.Zero(t, blocked.requests, "synthetic manual imports must remain local")
}
