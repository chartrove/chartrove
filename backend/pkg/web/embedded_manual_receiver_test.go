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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedManualImportExpectedReceiver(t *testing.T) {
	blocked := &blockedTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv("TMPDIR", t.TempDir())
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-receiver.db"), true)
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-manual-receiver-key"))
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
	signup := func(name string) (string, uuid.UUID) {
		t.Helper()
		w := request(http.MethodPost, "/api/auth/signup", "", `{"username":"`+name+`","password":"synthetic-password","join_mailing_list":false}`)
		require.Equal(t, http.StatusOK, w.Code)
		var response struct{ Data string }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.NotEmpty(t, response.Data)
		me := request(http.MethodGet, "/api/secure/account/me", response.Data, "")
		require.Equal(t, http.StatusOK, me.Code)
		var account struct{ Data struct{ ID uuid.UUID } }
		require.NoError(t, json.Unmarshal(me.Body.Bytes(), &account))
		require.NotEqual(t, uuid.Nil, account.Data.ID)
		return response.Data, account.Data.ID
	}
	count := func(token, path string) int {
		t.Helper()
		w := request(http.MethodGet, path, token, "")
		require.Equal(t, http.StatusOK, w.Code)
		var response struct{ Data []json.RawMessage }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		return len(response.Data)
	}
	importFile := func(token, query string, fields [][2]string, fileField string) *httptest.ResponseRecorder {
		t.Helper()
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		for _, field := range fields {
			require.NoError(t, form.WriteField(field[0], field[1]))
		}
		if fileField != "" {
			part, err := form.CreateFormFile(fileField, "synthetic-invalid-field.txt")
			require.NoError(t, err)
			_, err = io.WriteString(part, "synthetic-not-an-owner-field")
			require.NoError(t, err)
		}
		part, err := form.CreateFormFile("file", "synthetic-receiver.json")
		require.NoError(t, err)
		_, err = io.WriteString(part, `{"resourceType":"Bundle","type":"collection","entry":[{"resource":{"resourceType":"Patient","id":"synthetic-receiver-patient","gender":"unknown","name":[{"family":"Synthetic","given":["Receiver"]}]}}]}`)
		require.NoError(t, err)
		require.NoError(t, form.Close())
		r := httptest.NewRequest(http.MethodPost, "/api/secure/source/manual"+query, &body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w
	}
	owner, ownerID := signup("synthetic-receiver-owner")
	foreign, foreignID := signup("synthetic-receiver-foreign")
	ownerSources := count(owner, "/api/secure/source")
	ownerJobs := count(owner, "/api/secure/jobs")
	foreignSources := count(foreign, "/api/secure/source")
	for _, test := range []struct {
		name   string
		fields [][2]string
		query  string
		file   string
		status int
	}{
		{name: "foreign receiver", fields: [][2]string{{"expected_owner_id", foreignID.String()}}, status: http.StatusConflict},
		{name: "malformed", fields: [][2]string{{"expected_owner_id", "synthetic-invalid-owner"}}, status: http.StatusBadRequest},
		{name: "empty", fields: [][2]string{{"expected_owner_id", ""}}, status: http.StatusBadRequest},
		{name: "nil UUID", fields: [][2]string{{"expected_owner_id", uuid.Nil.String()}}, status: http.StatusBadRequest},
		{name: "duplicate", fields: [][2]string{{"expected_owner_id", ownerID.String()}, {"expected_owner_id", ownerID.String()}}, status: http.StatusBadRequest},
		{name: "case alias", fields: [][2]string{{"Expected_Owner_ID", ownerID.String()}}, status: http.StatusBadRequest},
		{name: "file instead of value", file: "expected_owner_id", status: http.StatusBadRequest},
		{name: "query instead of form", query: "?expected_owner_id=" + ownerID.String(), status: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := importFile(owner, test.query, test.fields, test.file)
			assert.Equal(t, test.status, w.Code)
			assert.Equal(t, ownerSources, count(owner, "/api/secure/source"), "rejected receiving authority must not create a source")
			assert.Equal(t, ownerJobs, count(owner, "/api/secure/jobs"), "rejected receiving authority must not start a tracked job")
			assert.Equal(t, foreignSources, count(foreign, "/api/secure/source"))
			assert.NotContains(t, w.Body.String(), "synthetic-receiver-patient")
		})
	}
	valid := importFile(owner, "", [][2]string{{"expected_owner_id", ownerID.String()}}, "")
	require.Equal(t, http.StatusOK, valid.Code)
	require.Equal(t, ownerSources+1, count(owner, "/api/secure/source"))
	require.Equal(t, foreignSources, count(foreign, "/api/secure/source"))
	legacy := importFile(foreign, "", nil, "")
	require.Equal(t, http.StatusOK, legacy.Code, "absent optional field preserves legacy manual clients")
	require.Equal(t, foreignSources+1, count(foreign, "/api/secure/source"))
	require.Equal(t, http.StatusOK, request(http.MethodDelete, "/api/secure/account/me", owner, "").Code)
	replacement, replacementID := signup("synthetic-receiver-owner")
	require.NotEqual(t, ownerID, replacementID)
	replacementSources := count(replacement, "/api/secure/source")
	replacementJobs := count(replacement, "/api/secure/jobs")
	stale := importFile(replacement, "", [][2]string{{"expected_owner_id", ownerID.String()}}, "")
	assert.Equal(t, http.StatusConflict, stale.Code, "a reused display name is not the original immutable receiver")
	assert.Equal(t, replacementSources, count(replacement, "/api/secure/source"))
	assert.Equal(t, replacementJobs, count(replacement, "/api/secure/jobs"))
	assert.Equal(t, http.StatusUnauthorized, importFile(owner, "", [][2]string{{"expected_owner_id", ownerID.String()}}, "").Code)
	require.Equal(t, http.StatusOK, importFile(replacement, "", [][2]string{{"expected_owner_id", replacementID.String()}}, "").Code)
	require.Equal(t, replacementSources+1, count(replacement, "/api/secure/source"))
	require.Zero(t, blocked.requests, "synthetic receiving-authority controls must not fetch external services")
}
