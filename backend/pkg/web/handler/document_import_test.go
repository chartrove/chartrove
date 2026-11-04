package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Real local JWT authentication and repository predicates, never a caller-supplied owner.
func TestDocumentImportHTTPAuthenticationAndReview(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-document-http.db"))
	cfg.Set("database.encryption.enabled", false)
	db, err := database.NewRepository(cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	users := []*models.User{{Username: "synthetic-import-owner", Password: "synthetic-password"}, {Username: "synthetic-import-delegate", Password: "synthetic-password"}, {Username: "synthetic-import-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}}
	for _, user := range users {
		require.NoError(t, db.CreateUser(context.Background(), user))
	}
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, users[0].Username)
	sources, err := db.GetSources(ctx)
	require.NoError(t, err)
	source := sources[0].ID
	r := gin.New()
	r.Use(middleware.RepositoryMiddleware(db), middleware.ConfigMiddleware(cfg), middleware.LoggerMiddleware(logrus.NewEntry(logrus.New())))
	r.POST("/api/auth/signin", AuthSignin)
	secure := r.Group("/api/secure").Use(middleware.RequireAuth())
	RegisterDocumentImportRoutes(secure)
	RegisterSourceAccessRoutes(secure)
	request := func(method, path, token, contentType string, body io.Reader, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "/api"+path, body)
		req.Header.Set("Content-Type", contentType)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, status, w.Code, w.Body.String())
		if strings.Contains(path, "/document-import") && status != 401 {
			require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		}
		return w
	}
	tokens := []string{}
	for _, user := range users {
		body, _ := json.Marshal(map[string]string{"username": user.Username, "password": "synthetic-password"})
		w := request("POST", "/auth/signin", "", "application/json", bytes.NewReader(body), 200)
		var result struct {
			Data string `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		tokens = append(tokens, result.Data)
	}
	path := "/secure/document-import/" + source.String()
	request("POST", path, "", "", strings.NewReader("not parsed before authentication"), 401)
	for _, token := range tokens[1:] {
		request("POST", path, token, "not-multipart", strings.NewReader("not parsed before ownership"), 404)
		w := request("GET", "/secure/document-import/sources", token, "", nil, 200)
		require.NotContains(t, w.Body.String(), source.String())
	}
	w := request("GET", "/secure/document-import/sources", tokens[0], "", nil, 200)
	require.Contains(t, w.Body.String(), source.String())
	for _, secret := range []string{"access_token", "refresh_token", "user_id", "password", "encryption"} {
		require.NotContains(t, w.Body.String(), secret)
	}
	post := func(fields [][2]string, file []byte, status int) *httptest.ResponseRecorder {
		t.Helper()
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		part, err := writer.CreateFormFile("file", "synthetic-scan.pdf")
		require.NoError(t, err)
		_, err = part.Write(file)
		require.NoError(t, err)
		for _, field := range fields {
			require.NoError(t, writer.WriteField(field[0], field[1]))
		}
		require.NoError(t, writer.Close())
		return request("POST", path, tokens[0], writer.FormDataContentType(), &buffer, status)
	}
	fields := [][2]string{{"title", "Synthetic reviewed scan"}, {"text", "Synthetic reviewed private text"}, {"reviewed", "true"}}
	original := []byte("%PDF-1.7\n% Synthetic preserved original\n%%EOF")
	post(fields[:2], original, 400)
	post(append(append([][2]string{}, fields...), [2]string{"owner_user_id", users[2].ID.String()}), original, 400)
	post(append(append([][2]string{}, fields...), [2]string{"text", "duplicate"}), original, 400)
	post(fields, []byte("not an accepted document"), 400)
	post(fields, bytes.Repeat([]byte("x"), models.DocumentMaxBytes+1), 400)
	post([][2]string{{"title", "Synthetic"}, {"text", strings.Repeat("x", models.DocumentMaxTranscriptBytes+1)}, {"reviewed", "true"}}, original, 400)
	w = post(fields, original, 200)
	require.NotContains(t, w.Body.String(), fields[1][1])
	require.NotContains(t, w.Body.String(), "resource_raw")
	var result struct {
		Data models.DocumentImportResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.True(t, result.Data.Created)
	repo := db.(database.SourceAccessRepository)
	grant, err := repo.SetSourceGrant(ctx, source, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	base := "/secure/source-access/" + source.String()
	metadata := base + "/metadata/DocumentReference"
	w = request("GET", metadata, tokens[1], "", nil, 200)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.Contains(t, w.Body.String(), result.Data.ResourceID)
	require.NotContains(t, w.Body.String(), "resource_raw")
	require.NotContains(t, w.Body.String(), fields[1][1])
	request("GET", metadata+"?page=-1", tokens[1], "", nil, 400)
	request("GET", metadata, tokens[2], "", nil, 404)
	detail := base + "/resources/DocumentReference/" + result.Data.ResourceID
	w = request("GET", detail, tokens[1], "", nil, 200)
	require.Contains(t, w.Body.String(), `"editable":false`)
	request("PATCH", detail, tokens[1], "application/json", strings.NewReader(`{"resource_raw":{"resourceType":"DocumentReference","id":"`+result.Data.ResourceID+`"}}`), 404)
	request("POST", path, tokens[1], "not-multipart", nil, 404)
	require.NoError(t, repo.RevokeSourceGrant(ctx, source, grant.ID))
	request("GET", metadata, tokens[1], "", nil, 404)
	request("GET", detail, tokens[1], "", nil, 404)
}
