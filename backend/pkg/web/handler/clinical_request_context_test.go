package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type capturedClinicalContext struct {
	method string
	ctx    context.Context
}

// Use actual route/envelope/error logic with a fake repository and synthetic
// server-identity middleware. Existing encrypted real-JWT HTTP tests remain the
// storage/authentication/authority controls.
type clinicalRequestContextProbe struct {
	database.DatabaseRepository
	actor    models.User
	captured []capturedClinicalContext
	onImport func(context.Context) error
}

func (p *clinicalRequestContextProbe) GetCurrentUser(ctx context.Context) (*models.User, error) {
	p.captured = append(p.captured, capturedClinicalContext{"actor", ctx})
	return &p.actor, nil
}
func (p *clinicalRequestContextProbe) ImportClinicalSnapshot(ctx context.Context, _, _ string, _ []byte) (*models.ClinicalSnapshotImportResult, error) {
	p.captured = append(p.captured, capturedClinicalContext{"import", ctx})
	if p.onImport != nil {
		return nil, p.onImport(ctx)
	}
	return nil, database.ErrClinicalSnapshotStorage
}
func (p *clinicalRequestContextProbe) ListClinicalSnapshots(ctx context.Context, _ uuid.UUID, _, _ int) ([]models.ClinicalSnapshotRecord, error) {
	p.captured = append(p.captured, capturedClinicalContext{"history", ctx})
	return nil, nil
}
func (p *clinicalRequestContextProbe) GetClinicalSnapshotOriginal(ctx context.Context, _, _ uuid.UUID) ([]byte, error) {
	p.captured = append(p.captured, capturedClinicalContext{"original", ctx})
	return []byte(`{}`), nil
}

func clinicalBoundaryRouter(probe *clinicalRequestContextProbe, original **gin.Context) *gin.Engine {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		*original = c
		c.Set(pkg.ContextKeyTypeDatabase, probe)
		c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
		c.Set(pkg.ContextKeyTypeAuthUserID, probe.actor.ID)
		c.Set(pkg.ContextKeyTypeAuthUsername, probe.actor.Username)
		c.Set(pkg.ContextKeyTypeAuthToken, "synthetic-private-local-token")
	})
	RegisterClinicalSnapshotRoutes(router)
	return router
}

func clinicalBoundaryBody(t *testing.T, owner uuid.UUID) string {
	t.Helper()
	body, err := json.Marshal(map[string]interface{}{
		"expected_owner_id": owner.String(), "confirmed": true,
		"origin":        map[string]string{"fhir_base_url": "https://clinical.example.invalid/fhir", "display": "Synthetic"},
		"snapshot_json": `{}`,
	})
	require.NoError(t, err)
	return string(body)
}

func TestClinicalSnapshotRoutesPreserveRequestContext(t *testing.T) {
	for _, test := range []struct {
		name, method, path string
		status             int
	}{
		{"import", http.MethodPost, "/source/clinical-snapshot", http.StatusInternalServerError},
		{"history", http.MethodGet, "/source/" + uuid.NewString() + "/clinical-snapshots", http.StatusOK},
		{"original", http.MethodGet, "/source/" + uuid.NewString() + "/clinical-snapshots/" + uuid.NewString() + "/original", http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			probe := &clinicalRequestContextProbe{actor: models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: "synthetic-original"}}
			var original *gin.Context
			router := clinicalBoundaryRouter(probe, &original)
			deadline := time.Now().Add(10 * time.Second)
			parent, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(clinicalBoundaryBody(t, probe.actor.ID))).WithContext(parent)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
			require.NotEmpty(t, probe.captured)
			// Model pool reuse deterministically, including timed child contexts
			// whose dynamic type alone would hide their retained Gin parent.
			cancel()
			original.Request = httptest.NewRequest(http.MethodGet, "/another-profile", nil)
			original.Set(pkg.ContextKeyTypeAuthUserID, uuid.New())
			original.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-next")
			for _, call := range probe.captured {
				t.Run(call.method, func(t *testing.T) {
					actual, ok := call.ctx.Deadline()
					require.True(t, ok, "repository must preserve the HTTP deadline")
					require.Equal(t, deadline, actual)
					require.ErrorIs(t, call.ctx.Err(), context.Canceled)
					require.Equal(t, probe.actor.ID, call.ctx.Value(pkg.ContextKeyTypeAuthUserID))
					require.Equal(t, probe.actor.Username, call.ctx.Value(pkg.ContextKeyTypeAuthUsername))
					require.Nil(t, call.ctx.Value(pkg.ContextKeyTypeAuthToken))
				})
			}
		})
	}
}

func TestClinicalSnapshotImportObservesHTTPAbort(t *testing.T) {
	entered, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	probe := &clinicalRequestContextProbe{actor: models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: "synthetic-original"}}
	probe.onImport = func(ctx context.Context) error {
		close(entered)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
			return database.ErrClinicalSnapshotStorage
		}
	}
	var original *gin.Context
	router := clinicalBoundaryRouter(probe, &original)
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	request := httptest.NewRequest(http.MethodPost, "/source/clinical-snapshot", strings.NewReader(clinicalBoundaryBody(t, probe.actor.ID))).WithContext(parent)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	go func() { router.ServeHTTP(response, request); close(finished) }()
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		select {
		case <-finished:
		case <-time.After(2 * time.Second):
			t.Error("owned synthetic HTTP handler did not exit after cleanup")
		}
	})
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("synthetic import boundary was not reached")
	}
	cancel()
	select {
	case <-finished:
		require.Equal(t, http.StatusRequestTimeout, response.Code)
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP cancellation did not reach the clinical import repository")
	}
}
