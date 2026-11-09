package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type oidcRequestContextProbe struct {
	database.DatabaseRepository
	database.OIDCRepository
	retained context.Context
}

func (p *oidcRequestContextProbe) GetCurrentUser(ctx context.Context) (*models.User, error) {
	p.retained = ctx
	return nil, errors.New("synthetic stop after current actor lookup")
}

// These are actual registered routes with synthetic identity middleware, not JWT,
// successful password proof or identity-provider acceptance. The fake stops before
// any password, session or external work and retains only the repository boundary.
func TestOIDCSecureRoutesPreserveRepositoryRequestContext(t *testing.T) {
	for _, test := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodPost, "/auth/oidc/link", http.StatusUnauthorized},
		{http.MethodGet, "/auth/oidc/identities", http.StatusInternalServerError},
		{http.MethodGet, "/auth/oidc/audit", http.StatusInternalServerError},
		{http.MethodPost, "/auth/oidc/identities/11111111-1111-4111-8111-111111111111/revoke", http.StatusUnauthorized},
		{http.MethodPost, "/auth/oidc/logout", http.StatusInternalServerError},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			probe := &oidcRequestContextProbe{}
			actor := uuid.New()
			var original *gin.Context
			router := gin.New()
			router.Use(func(c *gin.Context) {
				original = c
				c.Set(pkg.ContextKeyTypeDatabase, probe)
				c.Set(pkg.ContextKeyTypeConfig, cfg)
				c.Set(pkg.ContextKeyTypeAuthUserID, actor)
				c.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-original-oidc")
				c.Set(pkg.ContextKeyTypeAuthToken, "synthetic-private-local-jwt")
			})
			RegisterOIDCRoutes(router.Group("/api"), router.Group("/api/secure"), auth.NewOIDCManager(cfg))
			deadline := time.Now().Add(time.Minute)
			parent, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()
			request := httptest.NewRequest(test.method, "/api/secure"+test.path, strings.NewReader(`{"password":"synthetic-password"}`)).WithContext(parent)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			require.NotNil(t, probe.retained)
			_, pooled := probe.retained.(*gin.Context)
			require.False(t, pooled, "OIDC SQL must not retain pooled Gin")
			actual, ok := probe.retained.Deadline()
			require.True(t, ok)
			require.Equal(t, deadline, actual)
			cancel()
			original.Request = httptest.NewRequest(http.MethodGet, "/different-profile", nil)
			original.Set(pkg.ContextKeyTypeAuthUserID, uuid.New())
			original.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-next-oidc")
			require.ErrorIs(t, probe.retained.Err(), context.Canceled)
			require.Equal(t, actor, probe.retained.Value(pkg.ContextKeyTypeAuthUserID))
			require.Equal(t, "synthetic-original-oidc", probe.retained.Value(pkg.ContextKeyTypeAuthUsername))
			require.Nil(t, probe.retained.Value(pkg.ContextKeyTypeAuthToken))
		})
	}
}
