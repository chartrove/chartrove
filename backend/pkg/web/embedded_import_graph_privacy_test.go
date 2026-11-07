package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedSyntheticImportAndGraphKeepClinicalIdentifiersOutOfProcessOutput(t *testing.T) {
	blocked := &blockedTransport{}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	t.Setenv("TMPDIR", t.TempDir())
	capture, err := os.CreateTemp(t.TempDir(), "synthetic-process-output-")
	require.NoError(t, err)
	previousStdout, previousLog := os.Stdout, log.Writer()
	os.Stdout = capture
	log.SetOutput(capture)
	t.Cleanup(func() { os.Stdout = previousStdout; log.SetOutput(previousLog); capture.Close() })
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-private-import.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg) // Its configured application logger already discards output.
	require.NoError(t, app.UnlockDatabase("synthetic-private-import-key"))
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
	signup := func(name string) string {
		var token string
		require.NoError(t, json.Unmarshal(request("POST", "/api/auth/signup", "", `{"username":"`+name+`","password":"synthetic-password","join_mailing_list":false}`, 200), &token))
		return token
	}
	owner, other := signup("synthetic-private-owner"), signup("synthetic-private-other")
	// All data is generated here. Bundle aliases intentionally differ from IDs;
	// a second result shares the Patient but is not part of this report.
	const bundle = `{"resourceType":"Bundle","type":"collection","entry":[{"fullUrl":"urn:uuid:11111111-1111-4111-8111-111111111111","resource":{"resourceType":"Patient","id":"synthetic-private-patient","name":[{"family":"Synthetic","given":["Private"]}],"gender":"unknown"}},{"fullUrl":"urn:uuid:22222222-2222-4222-8222-222222222222","resource":{"resourceType":"DiagnosticReport","id":"synthetic-private-report","status":"final","code":{"text":"Synthetic private report"},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"result":[{"reference":"urn:uuid:33333333-3333-4333-8333-333333333333"}]}},{"fullUrl":"urn:uuid:33333333-3333-4333-8333-333333333333","resource":{"resourceType":"Observation","id":"synthetic-private-result","status":"final","code":{"coding":[{"system":"http://loinc.org","code":"SYNTHETIC-CODE"}]},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"valueQuantity":{"value":0}}},{"fullUrl":"urn:uuid:44444444-4444-4444-8444-444444444444","resource":{"resourceType":"Observation","id":"synthetic-private-unrelated","status":"final","code":{"text":"Synthetic other result"},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"valueString":"Synthetic unrelated"}}]}`
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	const filename = "synthetic-private-clinical-filename.json"
	file, err := form.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = io.WriteString(file, bundle)
	require.NoError(t, err)
	require.NoError(t, form.Close())
	r := httptest.NewRequest("POST", "/api/secure/source/manual", &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r.Header.Set("Authorization", "Bearer "+owner)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	var imported struct {
		Source struct{ ID string }
		Error  string
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &imported))
	require.Equal(t, 200, w.Code, "synthetic import error: %s", imported.Error)
	require.NotEmpty(t, imported.Source.ID)
	detail := "/api/secure/source-access/" + imported.Source.ID + "/resources/DiagnosticReport/synthetic-private-report"
	var report struct {
		Raw struct{ Result []struct{ Reference string } } `json:"resource_raw"`
	}
	require.NoError(t, json.Unmarshal(request("GET", detail, owner, "", 200), &report))
	require.Len(t, report.Raw.Result, 1)
	require.Equal(t, "urn:uuid:33333333-3333-4333-8333-333333333333", report.Raw.Result[0].Reference)
	request("GET", detail, other, "", 404)
	query := fmt.Sprintf(`{"resource_ids":[{"source_id":%q,"source_resource_type":"DiagnosticReport","source_resource_id":"synthetic-private-report"}]}`, imported.Source.ID)
	collect := func(raw json.RawMessage) []string {
		var value any
		require.NoError(t, json.Unmarshal(raw, &value))
		var ids []string
		var visit func(any)
		visit = func(v any) {
			switch x := v.(type) {
			case map[string]any:
				if x["source_resource_type"] == "Observation" {
					ids = append(ids, fmt.Sprint(x["source_resource_id"]))
				}
				for _, child := range x {
					visit(child)
				}
			case []any:
				for _, child := range x {
					visit(child)
				}
			}
		}
		visit(value)
		return ids
	}
	require.Equal(t, []string{"synthetic-private-result"}, collect(request("POST", "/api/secure/resource/graph/MedicalHistory", owner, query, 200)))
	require.Empty(t, collect(request("POST", "/api/secure/resource/graph/MedicalHistory", other, query, 200)))
	require.Zero(t, blocked.requests)
	require.NoError(t, capture.Sync())
	output, err := os.ReadFile(capture.Name())
	require.NoError(t, err)
	// Audit data remains encrypted in the repository. Process diagnostics must
	// not bypass the configured logger and disclose clinical/native identities.
	for _, marker := range []string{filename, imported.Source.ID, "synthetic-private-patient", "synthetic-private-report", "synthetic-private-result", "synthetic-private-unrelated"} {
		require.NotContains(t, string(output), marker)
	}
}
