package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Boundary fakes deliberately stop source handlers before SDK/network creation.
// The graph fake enforces the same immutable-owner predicate as the real store.
type unsafeContextProbe struct {
	database.DatabaseRepository
	actor                      *models.User
	lookup, operation, cleanup context.Context
}

func (p *unsafeContextProbe) GetUserByUsername(ctx context.Context, name string) (*models.User, error) {
	p.lookup = ctx
	if p.actor == nil || name != p.actor.Username {
		return nil, database.ErrSourceAccessDenied
	}
	return p.actor, nil
}
func (p *unsafeContextProbe) GetSource(ctx context.Context, _ string) (*models.SourceCredential, error) {
	p.operation = ctx
	return nil, errors.New("synthetic stop before provider creation")
}
func (p *unsafeContextProbe) GetFlattenedResourceGraph(ctx context.Context, _ pkg.ResourceGraphType, _ models.ResourceGraphOptions) (map[string][]*models.ResourceBase, error) {
	p.operation = ctx
	if ctx.Value(pkg.ContextKeyTypeAuthUserID) != p.actor.ID {
		return nil, database.ErrSourceAccessDenied
	}
	return map[string][]*models.ResourceBase{}, nil
}
func (p *unsafeContextProbe) UpdateSource(ctx context.Context, _ *models.SourceCredential) error {
	p.cleanup = ctx
	return nil
}

func TestUnsafeDiagnosticsPreserveSelectedOwnerRequest(t *testing.T) {
	for _, test := range []struct {
		name    string
		handler gin.HandlerFunc
		status  int
	}{
		{"raw source", UnsafeRequestSource, http.StatusInternalServerError},
		{"name sync", UnsafeSyncResourceNames, http.StatusInternalServerError},
		{"graph", UnsafeResourceGraph, http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("web.allow_unsafe_endpoints", true)
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			actor := uuid.New()
			probe := &unsafeContextProbe{actor: &models.User{ModelBase: models.ModelBase{ID: actor}, Username: "synthetic-selected"}}
			var original *gin.Context
			router := gin.New()
			router.Use(func(c *gin.Context) {
				original = c
				c.Set(pkg.ContextKeyTypeConfig, cfg)
				c.Set(pkg.ContextKeyTypeDatabase, probe)
				c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
			})
			router.POST("/:username/:sourceId/:graphType", test.handler)
			deadline := time.Now().Add(time.Minute)
			parent, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()
			parent = context.WithValue(parent, pkg.ContextKeyTypeAuthToken, "synthetic-inherited-local-token")
			parent = context.WithValue(parent, pkg.ContextKeyTypeAuthUserID, uuid.New())
			body := `{"resource_ids":[{"source_id":"` + uuid.NewString() + `","source_resource_type":"Patient","source_resource_id":"synthetic-patient"}]}`
			request := httptest.NewRequest(http.MethodPost, "/synthetic-selected/"+uuid.NewString()+"/all?resourceNames=Patient", strings.NewReader(body)).WithContext(parent)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code, response.Body.String())
			require.NotNil(t, probe.operation)
			_, pooled := probe.operation.(*gin.Context)
			require.False(t, pooled, "diagnostic SQL/SDK must not retain pooled Gin")
			actual, ok := probe.operation.Deadline()
			require.True(t, ok)
			require.Equal(t, deadline, actual)
			require.NotNil(t, probe.lookup, "resolve selected active owner exactly through the repository")
			require.Nil(t, probe.lookup.Value(pkg.ContextKeyTypeAuthToken))
			lookupDeadline, ok := probe.lookup.Deadline()
			require.True(t, ok)
			require.Equal(t, deadline, lookupDeadline)
			_, hasClaim := original.Get(pkg.ContextKeyTypeAuthUserID)
			require.False(t, hasClaim, "diagnostic selection must not manufacture Gin authentication claims")
			cancel()
			original.Request = httptest.NewRequest(http.MethodGet, "/another-owner", nil)
			original.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-next")
			original.Set(pkg.ContextKeyTypeAuthUserID, uuid.New())
			probe.actor.ID = uuid.New() // same-name account recreation must not rewrite the selected UUID.
			require.ErrorIs(t, probe.operation.Err(), context.Canceled)
			require.ErrorIs(t, probe.lookup.Err(), context.Canceled)
			require.Equal(t, actor, probe.operation.Value(pkg.ContextKeyTypeAuthUserID))
			require.Equal(t, "synthetic-selected", probe.operation.Value(pkg.ContextKeyTypeAuthUsername))
			require.Nil(t, probe.operation.Value(pkg.ContextKeyTypeAuthToken))
		})
	}
}

func TestUnsafeDiagnosticsRejectDisabledMode(t *testing.T) {
	for _, handler := range []gin.HandlerFunc{UnsafeRequestSource, UnsafeSyncResourceNames, UnsafeResourceGraph} {
		cfg, err := config.Create()
		require.NoError(t, err)
		cfg.Set("web.allow_unsafe_endpoints", false)
		probe := &unsafeContextProbe{}
		logger := logrus.New()
		logger.SetOutput(io.Discard)
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(pkg.ContextKeyTypeConfig, cfg)
			c.Set(pkg.ContextKeyTypeDatabase, probe)
			c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
		})
		router.GET("/", handler)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusServiceUnavailable, response.Code)
		require.Nil(t, probe.lookup)
		require.Nil(t, probe.operation)
	}
}

func TestUnsafeDiagnosticOwnerRefusesMissingDeletedAndCanceled(t *testing.T) {
	for _, state := range []string{"missing", "deleted", "nil UUID", "canceled"} {
		t.Run(state, func(t *testing.T) {
			probe := &unsafeContextProbe{actor: &models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: "synthetic-selected"}}
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch state {
			case "missing":
				probe.actor = nil
			case "deleted":
				now := time.Now()
				probe.actor.DeletedAt = &now
			case "nil UUID":
				probe.actor.ID = uuid.Nil
			case "canceled":
				cancel()
			}
			selected, err := resolveUnsafeDiagnosticOwner(parent, probe, "synthetic-selected")
			require.Error(t, err)
			require.Nil(t, selected)
			require.Nil(t, probe.operation)
			if state == "canceled" {
				require.Nil(t, probe.lookup, "do not start owner resolution after HTTP abort")
			}
		})
	}
}

func TestUnsafeCredentialCleanupPreservesBoundedSelectedOwner(t *testing.T) {
	parent, cancelRequest := context.WithCancel(context.Background())
	parent = context.WithValue(parent, pkg.ContextKeyTypeAuthToken, "synthetic-inherited-local-token")
	defer cancelRequest()
	id := uuid.New()
	probe := &unsafeContextProbe{actor: &models.User{ModelBase: models.ModelBase{ID: id}, Username: "synthetic-selected"}}
	selected, err := resolveUnsafeDiagnosticOwner(parent, probe, "synthetic-selected")
	require.NoError(t, err)
	cancelRequest()
	require.ErrorIs(t, selected.Err(), context.Canceled)
	probe.actor.ID = uuid.New()
	before := time.Now()
	cleanup, cancelCleanup := unsafeCredentialCleanupContext(selected)
	defer cancelCleanup()
	require.NoError(t, cleanup.Err(), "a rotated credential still needs bounded final storage after abort")
	deadline, ok := cleanup.Deadline()
	require.True(t, ok)
	require.WithinDuration(t, before.Add(10*time.Second), deadline, time.Second)
	require.Equal(t, id, cleanup.Value(pkg.ContextKeyTypeAuthUserID))
	require.Equal(t, "synthetic-selected", cleanup.Value(pkg.ContextKeyTypeAuthUsername))
	require.Nil(t, cleanup.Value(pkg.ContextKeyTypeAuthToken))
	require.NoError(t, probe.UpdateSource(cleanup, &models.SourceCredential{}))
	require.Same(t, cleanup, probe.cleanup)
	cancelCleanup()
	require.ErrorIs(t, cleanup.Err(), context.Canceled)
}
