package web

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedManualLabCollectionDaySurvivesImportAndRestart(t *testing.T) {
	blocked := &blockedTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv("TMPDIR", t.TempDir())
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-manual-lab.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-lab-day-key"))
	request := func(method, path, token, body string, status int) json.RawMessage {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		var result struct {
			Data  json.RawMessage
			Error string
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.Equal(t, status, w.Code, "synthetic request error: %s", result.Error)
		return result.Data
	}
	login := func(path, name string) string {
		var token string
		require.NoError(t, json.Unmarshal(request("POST", path, "", `{"username":"`+name+`","password":"synthetic-password","join_mailing_list":false}`, 200), &token))
		return token
	}
	owner := login("/api/auth/signup", "synthetic-lab-day-owner")
	unrelated := login("/api/auth/signup", "synthetic-lab-day-other")
	// Frontend's existing List/contained transport: the asserted day is distinct
	// from the encounter date; no clock time, timezone or issuance is fabricated.
	const list = `{"resourceType":"List","entry":[],"encounter":{"reference":"Encounter/synthetic-encounter"},"contained":[{"resourceType":"Encounter","id":"synthetic-encounter","status":"finished","period":{"start":"2024-12-31"}},{"resourceType":"DiagnosticReport","id":"synthetic-report","status":"final","code":{"text":"Synthetic panel"},"effectiveDateTime":"2020-02-29","result":[{"reference":"Observation/synthetic-observation"}]},{"resourceType":"Observation","id":"synthetic-observation","status":"final","code":{"text":"Synthetic lab"},"effectiveDateTime":"2020-02-29","valueQuantity":{"value":0,"unit":"mg/dL"},"note":[{"text":"Retain café صفر"}]}]}`
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("file", "synthetic-lab-day.json")
	require.NoError(t, err)
	_, err = io.WriteString(file, list)
	require.NoError(t, err)
	require.NoError(t, form.Close())
	r := httptest.NewRequest("POST", "/api/secure/resource/related", &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r.Header.Set("Authorization", "Bearer "+owner)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	require.Equal(t, 200, w.Code)
	var result struct {
		Source struct{ ID string }
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.NotEmpty(t, result.Source.ID)
	base := "/api/secure/source-access/" + result.Source.ID + "/resources/"
	check := func() {
		t.Helper()
		for _, typ := range []string{"Observation", "DiagnosticReport"} {
			id := "synthetic-observation"
			if typ == "DiagnosticReport" {
				id = "synthetic-report"
			}
			var stored models.SharedResource
			require.NoError(t, json.Unmarshal(request("GET", base+typ+"/"+id, owner, "", 200), &stored))
			var raw map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(stored.ResourceRaw, &raw))
			require.Equal(t, `"2020-02-29"`, string(raw["effectiveDateTime"]))
			require.NotContains(t, raw, "issued")
			require.NotNil(t, stored.SortDate)
			require.True(t, stored.SortDate.Equal(time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC)))
			if typ == "Observation" {
				require.JSONEq(t, `{"value":0,"unit":"mg/dL"}`, string(raw["valueQuantity"]))
				require.Contains(t, string(raw["note"]), "café صفر")
			}
			request("GET", base+typ+"/"+id, unrelated, "", 404)
			request("GET", base+typ+"/"+id, "", "", 401)
		}
	}
	check()
	require.NoError(t, app.Close())
	app = openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-lab-day-key"))
	owner = login("/api/auth/signin", "synthetic-lab-day-owner")
	unrelated = login("/api/auth/signin", "synthetic-lab-day-other")
	check()
	require.Zero(t, blocked.requests, "manual lab date ingestion must stay local")
}
