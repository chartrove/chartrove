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

	"github.com/stretchr/testify/require"
)

func TestEmbeddedGraphKeepsSameBusinessIDsInTheirNativeSources(t *testing.T) {
	blocked := &blockedTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv("TMPDIR", t.TempDir())
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-graph-sources.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-graph-source-key"))
	request := func(method, path, token, body string, status int) json.RawMessage {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		var response struct {
			Data  json.RawMessage
			Error string
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, status, w.Code, "synthetic request error: %s", response.Error)
		return response.Data
	}
	signup := func(name string) string {
		var token string
		require.NoError(t, json.Unmarshal(request("POST", "/api/auth/signup", "", `{"username":"`+name+`","password":"synthetic-password","join_mailing_list":false}`, 200), &token))
		return token
	}
	owner, other := signup("synthetic-graph-owner"), signup("synthetic-graph-other")
	var sources []string
	for i, patient := range []string{"synthetic-patient-A", "synthetic-patient-B"} {
		// Same report/result IDs in two genuinely different source namespaces.
		// Bundle alias UUIDs also differ from each resource's business ID.
		bundle := fmt.Sprintf(`{"resourceType":"Bundle","type":"collection","entry":[{"fullUrl":"urn:uuid:11111111-1111-4111-8111-111111111111","resource":{"resourceType":"Patient","id":%q,"name":[{"family":"Synthetic"}],"gender":"unknown"}},{"fullUrl":"urn:uuid:22222222-2222-4222-8222-222222222222","resource":{"resourceType":"DiagnosticReport","id":"synthetic-shared-report","status":"final","code":{"text":"Synthetic report"},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"result":[{"reference":"urn:uuid:33333333-3333-4333-8333-333333333333"}]}},{"fullUrl":"urn:uuid:33333333-3333-4333-8333-333333333333","resource":{"resourceType":"Observation","id":"synthetic-shared-result","status":"final","code":{"coding":[{"system":"http://loinc.org","code":"SYNTHETIC-%d"}]},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"derivedFrom":[{"reference":"urn:uuid:44444444-4444-4444-8444-444444444444"}],"valueInteger":%d}}]}`, patient, i, i)
		// Preserve the generic graph contract, including reciprocal focus links.
		// Adjacency is deliberately not a claim of DiagnosticReport.result membership.
		bundle = strings.TrimSuffix(bundle, "]}") + `,{"fullUrl":"urn:uuid:44444444-4444-4444-8444-444444444444","resource":{"resourceType":"Observation","id":"synthetic-derived-input","status":"final","code":{"coding":[{"system":"http://loinc.org","code":"SYNTHETIC-INPUT"}]},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"valueInteger":99}},{"fullUrl":"urn:uuid:55555555-5555-4555-8555-555555555555","resource":{"resourceType":"Observation","id":"synthetic-focus-result","status":"final","focus":[{"reference":"urn:uuid:22222222-2222-4222-8222-222222222222"}],"code":{"coding":[{"system":"http://loinc.org","code":"SYNTHETIC-FOCUS"}]},"subject":{"reference":"urn:uuid:11111111-1111-4111-8111-111111111111"},"valueInteger":100}}]}`

		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		file, err := form.CreateFormFile("file", "synthetic-graph-bundle.json")
		require.NoError(t, err)
		_, err = io.WriteString(file, bundle)
		require.NoError(t, err)
		require.NoError(t, form.Close())
		r := httptest.NewRequest("POST", "/api/secure/source/manual", &body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("Authorization", "Bearer "+owner)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		var response struct {
			Source struct{ ID string }
			Error  string
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, 200, w.Code, "synthetic import error: %s", response.Error)
		sources = append(sources, response.Source.ID)
	}
	require.NotEmpty(t, sources[0])
	require.NotEqual(t, sources[0], sources[1])
	for i, source := range sources {
		query := fmt.Sprintf(`{"resource_ids":[{"source_id":%q,"source_resource_type":"DiagnosticReport","source_resource_id":"synthetic-shared-report"}]}`, source)
		var graph struct {
			Results map[string][]struct {
				SourceID string          `json:"source_id"`
				Raw      json.RawMessage `json:"resource_raw"`
				Related  []struct {
					SourceID string `json:"source_id"`
					Type     string `json:"source_resource_type"`
					ID       string `json:"source_resource_id"`
					Raw      struct {
						Value int `json:"valueInteger"`
					} `json:"resource_raw"`
				} `json:"related_resources"`
			}
		}
		require.NoError(t, json.Unmarshal(request("POST", "/api/secure/resource/graph/MedicalHistory", owner, query, 200), &graph))
		require.Len(t, graph.Results["DiagnosticReport"], 1)
		root := graph.Results["DiagnosticReport"][0]
		require.Equal(t, source, root.SourceID)
		require.Contains(t, string(root.Raw), "urn:uuid:33333333-3333-4333-8333-333333333333")
		found := 0
		for _, child := range root.Related {
			if child.Type != "Observation" {
				continue
			}
			found++
			require.Equal(t, source, child.SourceID, "graph expansion must retain the selected native source")
			expected := map[string]int{"synthetic-shared-result": i, "synthetic-focus-result": 100}
			require.Contains(t, expected, child.ID)
			require.Equal(t, expected[child.ID], child.Raw.Value)
		}
		require.Equal(t, 2, found, "preserve incoming reciprocal focus links as generic relationships")
		var denied struct{ Results map[string]json.RawMessage }
		require.NoError(t, json.Unmarshal(request("POST", "/api/secure/resource/graph/MedicalHistory", other, query, 200), &denied))
		require.NotContains(t, denied.Results, "DiagnosticReport")
	}
	require.Zero(t, blocked.requests)
}
