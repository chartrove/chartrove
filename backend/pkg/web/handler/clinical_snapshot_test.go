package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const clinicalAPIBase = "https://synthetic.example.invalid/fhir"
const clinicalAPIPost = "/api/secure/source/clinical-snapshot"
const clinicalAPIRaw = `{
 "resourceType":"Bundle", "type":"collection", "entry":[
 {"resource":{"resourceType":"Patient","id":"synthetic-patient","name":[{"text":"synthetic-private-clinical-name"}]}},
 {"resource":{"resourceType":"Observation","id":"synthetic-observation","status":"final","subject":{"reference":"Patient/synthetic-patient"},"code":{"text":"synthetic-private-clinical-result"},"valueQuantity":{"value":1.234567890123456789,"unit":"mg"},"extension":[{"url":"https://example.invalid/unknown","valueString":"synthetic-private-extension café"}]}}
 ]}`

type clinicalAPIEvents struct {
	event_bus.Interface
	count atomic.Int32
	check func() error
}

func (b *clinicalAPIEvents) PublishMessage(models.EventInterface) error {
	b.count.Add(1)
	if b.check != nil {
		return b.check()
	}
	return nil
}

type clinicalAPIHarness struct {
	router *gin.Engine
	repo   *database.GormRepository
	cfg    config.Interface
	users  []*models.User
	tokens []string
	owner  context.Context
	logs   *bytes.Buffer
	bus    *clinicalAPIEvents
}

func newClinicalAPIHarness(t *testing.T) clinicalAPIHarness {
	t.Helper()
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-api.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-clinical-api-key")
	cfg.Set("jwt.issuer.key", "synthetic-clinical-api-signing-key")
	logs := &bytes.Buffer{}
	logger := logrus.New()
	logger.SetOutput(logs)
	bus := &clinicalAPIEvents{Interface: event_bus.NewNoopEventBusServer()}
	db, err := database.NewRepository(cfg, logger, bus)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	r := db.(*database.GormRepository)
	h := clinicalAPIHarness{repo: r, cfg: cfg, logs: logs, bus: bus}
	for i, name := range []string{"synthetic-api-owner", "synthetic-api-delegate", "synthetic-api-admin"} {
		u := &models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: name, Password: "synthetic-placeholder-not-authenticated", Role: pkg.UserRoleUser}
		if i == 2 {
			u.Role = pkg.UserRoleAdmin
		}
		// Real signed-JWT middleware is exercised; fixture insertion is not a
		// password/OIDC authentication test or a production account operation.
		require.NoError(t, r.GormClient.Session(&gorm.Session{SkipHooks: true}).Create(u).Error)
		h.users = append(h.users, u)
		token, err := auth.JwtGenerateFastenTokenFromUser(*u, cfg.GetString("jwt.issuer.key"))
		require.NoError(t, err)
		h.tokens = append(h.tokens, token)
	}
	h.owner = context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, h.users[0].Username)
	h.owner = context.WithValue(h.owner, pkg.ContextKeyTypeAuthUserID, h.users[0].ID)
	h.router = clinicalAPIRouter(cfg, r, logger, bus)
	return h
}

func clinicalAPIRouter(cfg config.Interface, db database.DatabaseRepository, logger *logrus.Logger, bus event_bus.Interface) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(pkg.ContextKeyTypeConfig, cfg)
		c.Set(pkg.ContextKeyTypeDatabase, db)
		c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
		c.Set(pkg.ContextKeyTypeEventBusServer, bus)
	})
	RegisterClinicalSnapshotRoutes(router.Group("/api/secure").Use(middleware.RequireAuth()))
	return router
}

func (h clinicalAPIHarness) body(t *testing.T, raw string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]interface{}{
		"expected_owner_id": h.users[0].ID.String(), "confirmed": true,
		"origin":        map[string]string{"fhir_base_url": clinicalAPIBase, "display": "Synthetic clinic"},
		"snapshot_json": raw,
	})
	require.NoError(t, err)
	return body
}

func (h clinicalAPIHarness) request(method, path, token string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	h.router.ServeHTTP(response, request)
	return response
}

func clinicalAPIResult(t *testing.T, response *httptest.ResponseRecorder) models.ClinicalSnapshotImportResult {
	t.Helper()
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var wrapper struct {
		Success bool                                `json:"success"`
		Data    models.ClinicalSnapshotImportResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &wrapper))
	require.True(t, wrapper.Success)
	return wrapper.Data
}

func TestClinicalSnapshotHTTPFlow(t *testing.T) {
	h := newClinicalAPIHarness(t)
	body := h.body(t, clinicalAPIRaw)
	require.Equal(t, http.StatusUnauthorized, h.request(http.MethodPost, clinicalAPIPost, "", body).Code)
	// Notification callback sees a fully committed collection and original.
	h.bus.check = func() error {
		var count int64
		if err := h.repo.GormClient.Model(&models.ClinicalSnapshotRecord{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("synthetic observer saw uncommitted original")
		}
		return nil
	}
	first := clinicalAPIResult(t, h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], body))
	require.Equal(t, 2, first.Created)
	require.EqualValues(t, 1, h.bus.count.Load())
	prefix := "/api/secure/source/" + first.SourceID.String() + "/clinical-snapshots"
	metadata := h.request(http.MethodGet, prefix, h.tokens[0], nil)
	require.Equal(t, http.StatusOK, metadata.Code)
	require.Equal(t, "no-store", metadata.Header().Get("Cache-Control"))
	require.NotContains(t, metadata.Body.String(), "synthetic-private-")
	require.Contains(t, metadata.Body.String(), models.ClinicalSnapshotOwnerAcquisition)
	originalPath := prefix + "/" + first.SnapshotID.String() + "/original"
	original := h.request(http.MethodGet, originalPath, h.tokens[0], nil)
	require.Equal(t, http.StatusOK, original.Code, original.Body.String())
	require.Equal(t, clinicalAPIRaw, original.Body.String(), "HTTP original download must preserve exact lexical bytes")
	require.Equal(t, "application/fhir+json", original.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", original.Header().Get("X-Content-Type-Options"))
	require.Contains(t, original.Header().Get("Content-Disposition"), "attachment;")
	require.Equal(t, "no-store", original.Header().Get("Cache-Control"))
	for _, token := range h.tokens[1:] {
		require.Equal(t, http.StatusForbidden, h.request(http.MethodGet, prefix, token, nil).Code)
		require.Equal(t, http.StatusForbidden, h.request(http.MethodGet, originalPath, token, nil).Code)
		require.Equal(t, http.StatusForbidden, h.request(http.MethodPost, clinicalAPIPost, token, body).Code, "stale receiving owner must not silently change")
	}
	_, err := h.repo.SetSourceGrant(h.owner, first.SourceID, h.users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, h.request(http.MethodGet, originalPath, h.tokens[1], nil).Code)
	repeated := clinicalAPIResult(t, h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], body))
	require.True(t, repeated.AlreadyImported)
	require.EqualValues(t, 1, h.bus.count.Load(), "historical acknowledgement is not a new commit notification")
	h.bus.check = func() error { return fmt.Errorf("synthetic-private-notification-failure") }
	changedRaw := strings.Replace(clinicalAPIRaw, "1.234567890123456789", "2.345678901234567891", 1)
	changed := clinicalAPIResult(t, h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], h.body(t, changedRaw)))
	require.Equal(t, 1, changed.Updated)
	require.EqualValues(t, 2, h.bus.count.Load())
	require.NotContains(t, h.logs.String(), "synthetic-private-notification-failure")
	for _, marker := range []string{"synthetic-private-clinical-name", "synthetic-private-clinical-result", "1.234567890123456789"} {
		require.NotContains(t, h.logs.String(), marker)
	}
	for _, query := range []string{"?limit=0", "?limit=201", "?offset=-1", "?limit=invalid", "?offset=1000001"} {
		require.Equal(t, http.StatusBadRequest, h.request(http.MethodGet, prefix+query, h.tokens[0], nil).Code)
	}
}

func TestClinicalSnapshotHTTPStrictEnvelope(t *testing.T) {
	h := newClinicalAPIHarness(t)
	valid := string(h.body(t, clinicalAPIRaw))
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"absent confirmation", strings.Replace(valid, `"confirmed":true,`, "", 1), 400},
		{"false confirmation", strings.Replace(valid, `"confirmed":true`, `"confirmed":false`, 1), 400},
		{"duplicate confirmation", strings.Replace(valid, `"confirmed":true`, `"confirmed":true,"confirmed":false`, 1), 400},
		{"case alias", strings.Replace(valid, `"confirmed"`, `"Confirmed"`, 1), 400},
		{"unknown credential", strings.Replace(valid, `"confirmed":true`, `"confirmed":true,"access_token":"synthetic-private-request-token"`, 1), 400},
		{"unknown origin", strings.Replace(valid, `"display":"Synthetic clinic"`, `"display":"Synthetic clinic","code_verifier":"synthetic-private-verifier"`, 1), 400},
		{"duplicate origin", strings.Replace(valid, `"display":"Synthetic clinic"`, `"display":"Synthetic clinic","display":"different"`, 1), 400},
		{"null envelope", "null", 400},
		{"array", "[]", 400},
		{"extra document", valid + " {}", 400},
		{"malformed JSON", "{", 400},
		{"invalid UTF8", string([]byte{'{', '"', 0xff, '"', ':', '0', '}'}), 400},
		{"different recipient", strings.Replace(valid, h.users[0].ID.String(), h.users[1].ID.String(), 1), 403},
		{"credential-bearing origin", strings.Replace(valid, clinicalAPIBase, "https://synthetic-private-user:synthetic-private-password@example.invalid/fhir", 1), 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], []byte(tc.body))
			require.Equal(t, tc.status, response.Code, response.Body.String())
			require.NotContains(t, response.Body.String(), "synthetic-private-")
		})
	}
	oversized := h.body(t, strings.Repeat(" ", clinicalsnapshot.MaxBytes+1))
	require.Equal(t, http.StatusBadRequest, h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], oversized).Code)
	var count int64
	require.NoError(t, h.repo.GormClient.Model(&models.ClinicalSnapshotRecord{}).Count(&count).Error)
	require.Zero(t, count)
	require.Zero(t, h.bus.count.Load())
	require.NotContains(t, h.logs.String(), "synthetic-private-")
	// Same exact positive prerequisites must still reach real storage.
	require.Equal(t, 2, clinicalAPIResult(t, h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], []byte(valid))).Created)
}

func TestClinicalSnapshotHTTPRollbackDoesNotNotify(t *testing.T) {
	h := newClinicalAPIHarness(t)
	name := "synthetic_clinical_http_final_failure"
	require.NoError(t, h.repo.GormClient.Callback().Create().Before("gorm:create").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "source_access_audits" {
			tx.AddError(errors.New("synthetic-private-final-audit-failure"))
		}
	}))
	response := h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], h.body(t, clinicalAPIRaw))
	require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
	require.NotContains(t, response.Body.String(), "synthetic-private-")
	require.NotContains(t, h.logs.String(), "synthetic-private-")
	require.Zero(t, h.bus.count.Load(), "failed final write cannot emit a success event")
	for _, model := range []interface{}{&models.ClinicalSnapshotRecord{}, &models.SourceCredential{}} {
		var count int64
		require.NoError(t, h.repo.GormClient.Model(model).Count(&count).Error)
		require.Zero(t, count)
	}
	require.NoError(t, h.repo.GormClient.Callback().Create().Remove(name))
	require.Equal(t, 2, clinicalAPIResult(t, h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], h.body(t, clinicalAPIRaw))).Created)
}

type clinicalAPIBlockingRepository struct {
	database.DatabaseRepository
	database.ClinicalSnapshotRepository
	entered chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (r *clinicalAPIBlockingRepository) ImportClinicalSnapshot(context.Context, string, string, []byte) (*models.ClinicalSnapshotImportResult, error) {
	if r.calls.Add(1) == 1 {
		close(r.entered)
		<-r.release
	}
	return &models.ClinicalSnapshotImportResult{AlreadyImported: true}, nil
}

func TestClinicalSnapshotHTTPBusyAndWireLimits(t *testing.T) {
	h := newClinicalAPIHarness(t)
	blocked := &clinicalAPIBlockingRepository{DatabaseRepository: h.repo, ClinicalSnapshotRepository: h.repo, entered: make(chan struct{}), release: make(chan struct{})}
	logger := logrus.New()
	logger.SetOutput(h.logs)
	h.router = clinicalAPIRouter(h.cfg, blocked, logger, h.bus)
	responses := make(chan *httptest.ResponseRecorder, 1)
	body := h.body(t, clinicalAPIRaw)
	go func() { responses <- h.request(http.MethodPost, clinicalAPIPost, h.tokens[0], body) }()
	unblock := sync.OnceFunc(func() { close(blocked.release) })
	defer unblock()
	select {
	case <-blocked.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("first authenticated request did not reach the optional repository")
	}
	busy := h.request(http.MethodPost, clinicalAPIPost, h.tokens[1], body)
	require.Equal(t, http.StatusServiceUnavailable, busy.Code)
	require.EqualValues(t, 1, blocked.calls.Load(), "busy requests must not parse or start another expensive projection")
	// Release and await the handler before teardown, including failure paths.
	unblock()
	select {
	case response := <-responses:
		require.Equal(t, http.StatusOK, response.Code)
	case <-time.After(5 * time.Second):
		t.Fatal("released import did not finish")
	}
	// The bounded reader must reject actual bytes even when Content-Length
	// is absent. A native-sized collection still has its separate decoded cap.
	h.router = clinicalAPIRouter(h.cfg, h.repo, logger, h.bus)
	request := httptest.NewRequest(http.MethodPost, clinicalAPIPost, strings.NewReader(strings.Repeat(" ", 2*clinicalsnapshot.MaxBytes+(8<<10)+1)))
	request.ContentLength = -1
	request.Header.Set("Authorization", "Bearer "+h.tokens[0])
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	h.router.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Zero(t, h.bus.count.Load())
}
