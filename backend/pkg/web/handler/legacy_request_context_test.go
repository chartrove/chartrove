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

// Only the boundary is fake: these real HTTP handlers must not lend their pooled
// Gin object to a repository, even if SQL cleanup outlives the response.
type legacyContextProbe struct {
	database.DatabaseRepository
	retained context.Context
	actor    *models.User
}

func (p *legacyContextProbe) GetSources(ctx context.Context) ([]models.SourceCredential, error) {
	p.retained = ctx
	return nil, nil
}
func (p *legacyContextProbe) DeleteSource(ctx context.Context, _ string) (int64, error) {
	p.retained = ctx
	return 0, nil
}
func (p *legacyContextProbe) AddResourceComposition(ctx context.Context, _ string, _ []*models.ResourceBase) error {
	p.retained = ctx
	return nil
}
func (p *legacyContextProbe) RegisterUser(ctx context.Context, _ *models.User, _ bool) error {
	p.retained = ctx
	return errors.New("synthetic registration refusal")
}

func (p *legacyContextProbe) GetCurrentUser(ctx context.Context) (*models.User, error) {
	p.retained = ctx
	return p.actor, nil
}
func (p *legacyContextProbe) GetSummary(ctx context.Context) (*models.Summary, error) {
	p.retained = ctx
	return &models.Summary{}, nil
}
func (p *legacyContextProbe) LoadUserSettings(ctx context.Context) (*models.UserSettings, error) {
	p.retained = ctx
	return &models.UserSettings{}, nil
}
func (p *legacyContextProbe) GetUserAccessTokens(ctx context.Context) ([]models.AccessToken, error) {
	p.retained = ctx
	return nil, nil
}
func (p *legacyContextProbe) GetUserFavorites(ctx context.Context, _, _ string) ([]models.Favorite, error) {
	p.retained = ctx
	return nil, nil
}
func (p *legacyContextProbe) ListBackgroundJobs(ctx context.Context, _ models.BackgroundJobQueryOptions) ([]models.BackgroundJob, error) {
	p.retained = ctx
	return nil, nil
}
func (p *legacyContextProbe) CreateBackgroundJob(ctx context.Context, _ *models.BackgroundJob) error {
	p.retained = ctx
	return nil
}
func (p *legacyContextProbe) GetUsers(ctx context.Context) ([]models.User, error) {
	p.retained = ctx
	return nil, nil
}

func TestLegacyHandlersPreserveRepositoryRequestLifetime(t *testing.T) {
	for _, test := range []struct {
		name      string
		handler   gin.HandlerFunc
		body      string
		anonymous bool
		status    int
	}{
		{"source list", ListSource, "", false, http.StatusOK},
		{"source deletion", DeleteSource, "", false, http.StatusOK},
		{"composition", CreateResourceComposition, `{"title":"Synthetic","resources":[]}`, false, http.StatusOK},
		{"registration", AuthSignup, `{"username":"synthetic-new","password":"synthetic-password"}`, true, http.StatusInternalServerError},
		{"summary", GetSummary, "", false, http.StatusOK},
		{"dashboard settings", GetDashboard, "", false, http.StatusOK},
		{"access tokens", GetAccessTokens, "", false, http.StatusOK},
		{"OIDC link password lookup", func(c *gin.Context) {
			// Empty fixture hash refuses proof after the real helper's repository lookup.
			// This checks request lifetime, not successful password/JWT authentication.
			oidcPasswordUser(c, c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository))
		}, `{"password":"synthetic-password"}`, false, http.StatusUnauthorized},
		{"practitioner favorites", GetUserFavoritePractitioners, "", false, http.StatusOK},
		{"background job list", ListBackgroundJobs, "", false, http.StatusOK},
		{"client failure record", CreateBackgroundJobError, `{}`, false, http.StatusOK},
		{"admin user list", GetUsers, "", false, http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("auth.registration.enabled", true)
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			probe := &legacyContextProbe{}
			actor := uuid.New()
			probe.actor = &models.User{ModelBase: models.ModelBase{ID: actor}, Username: "synthetic-original", Role: pkg.UserRoleAdmin}
			var original *gin.Context
			router := gin.New()
			router.Use(func(c *gin.Context) {
				original = c
				c.Set(pkg.ContextKeyTypeDatabase, probe)
				c.Set(pkg.ContextKeyTypeConfig, cfg)
				c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
				if !test.anonymous {
					c.Set(pkg.ContextKeyTypeAuthUserID, actor)
					c.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-original")
					c.Set(pkg.ContextKeyTypeAuthToken, "synthetic-private-local-token")
				}
			})
			router.POST("/boundary/:sourceId", test.handler)
			deadline := time.Now().Add(time.Minute)
			parent, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()
			request := httptest.NewRequest(http.MethodPost, "/boundary/"+uuid.NewString(), strings.NewReader(test.body)).WithContext(parent)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
			require.NotNil(t, probe.retained)
			_, pooled := probe.retained.(*gin.Context)
			require.False(t, pooled, "SQL must not retain a reusable Gin context")
			actual, ok := probe.retained.Deadline()
			require.True(t, ok, "repository must receive the HTTP deadline")
			require.Equal(t, deadline, actual)
			cancel()
			// Deterministically model the pool's next request without a timing race.
			original.Request = httptest.NewRequest(http.MethodGet, "/another-profile", nil)
			original.Set(pkg.ContextKeyTypeAuthUserID, uuid.New())
			original.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-next")
			require.ErrorIs(t, probe.retained.Err(), context.Canceled)
			require.Nil(t, probe.retained.Value(pkg.ContextKeyTypeAuthToken))
			if !test.anonymous {
				require.Equal(t, actor, probe.retained.Value(pkg.ContextKeyTypeAuthUserID))
				require.Equal(t, "synthetic-original", probe.retained.Value(pkg.ContextKeyTypeAuthUsername))
			}
		})
	}
}
