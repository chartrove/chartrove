package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type accessTokenFailureProbe struct {
	database.DatabaseRepository
	user *models.User
	ctx  context.Context
}

func (p *accessTokenFailureProbe) GetCurrentUser(context.Context) (*models.User, error) {
	return p.user, nil
}

func (p *accessTokenFailureProbe) CreateAccessToken(ctx context.Context, _ *models.AccessToken) error {
	p.ctx = ctx
	return errors.New("synthetic storage failure: private-access-token-error-canary")
}

// This boundary-only fake isolates handler error privacy; it is not a SQL logger,
// password authentication, durable database or JWT middleware acceptance test.
func TestCreateAccessTokenPersistenceFailureIsPrivate(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("jwt.issuer.key", "synthetic-access-token-signing-key")
	user := &models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: "synthetic-token-owner"}
	probe := &accessTokenFailureProbe{user: user}
	var logs bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&logs)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodPost, "/synthetic-access-token", strings.NewReader(`{"name":"synthetic companion","expiration":1}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(pkg.ContextKeyTypeConfig, cfg)
	c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
	c.Set(pkg.ContextKeyTypeDatabase, probe)
	c.Set(pkg.ContextKeyTypeAuthUserID, user.ID)
	c.Set(pkg.ContextKeyTypeAuthUsername, user.Username)
	c.Set(pkg.ContextKeyTypeAuthToken, "synthetic-private-application-token")
	CreateAccessToken(c)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, false, body["success"])
	_, hasCredential := body["data"]
	assert.False(t, hasCredential, "failed persistence must not publish a signed credential")
	assert.Equal(t, "Failed to store access token", body["error"])
	assert.NotContains(t, logs.String(), "private-access-token-error-canary")
	assert.NotContains(t, response.Body.String(), "private-access-token-error-canary")
	require.NotNil(t, probe.ctx)
	_, retainedGin := probe.ctx.(*gin.Context)
	require.False(t, retainedGin)
	require.Equal(t, user.ID, probe.ctx.Value(pkg.ContextKeyTypeAuthUserID))
	require.Nil(t, probe.ctx.Value(pkg.ContextKeyTypeAuthToken))
}

// Real SQLCipher storage and signed-JWT middleware, with a targeted GORM create
// failure. Fixture accounts bypass password hashing, not a password/IdP test.
func TestAccessTokenEncryptedPersistenceAndRevocation(t *testing.T) {
	h := newClinicalAPIHarness(t)
	secure := h.router.Group("/synthetic-token").Use(middleware.RequireAuth())
	secure.POST("", CreateAccessToken)
	secure.POST("/delete", DeleteAccessToken)
	secure.GET("/probe", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	var failed models.AccessToken
	const callback = "synthetic_access_token_create_failure"
	require.NoError(t, h.repo.GormClient.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if token, ok := tx.Statement.Dest.(*models.AccessToken); ok {
			failed = *token
			tx.AddError(errors.New("synthetic access storage failure"))
		}
	}))
	var removeOnce sync.Once
	removeFailure := func() {
		removeOnce.Do(func() { require.NoError(t, h.repo.GormClient.Callback().Create().Remove(callback)) })
	}
	t.Cleanup(removeFailure)

	refused := h.request(http.MethodPost, "/synthetic-token", h.tokens[0], []byte(`{"name":"synthetic companion","expiration":1}`))
	assert.Equal(t, http.StatusInternalServerError, refused.Code)
	assert.Equal(t, "no-store", refused.Header().Get("Cache-Control"))
	var failedBody map[string]any
	require.NoError(t, json.Unmarshal(refused.Body.Bytes(), &failedBody))
	assert.Equal(t, false, failedBody["success"])
	_, hasCredential := failedBody["data"]
	assert.False(t, hasCredential, "real failed storage must not publish a credential")
	require.NotEmpty(t, failed.TokenID)
	rows, err := h.repo.GetUserAccessTokens(h.owner)
	require.NoError(t, err)
	require.Empty(t, rows)
	// A correctly signed synthetic credential for that absent row is refused.
	// This preserves the distinction from an authentication-bypass allegation.
	unstored, err := auth.JwtGenerateAccessToken(*h.users[0], h.cfg.GetString("jwt.issuer.key"), failed.ExpiresAt, failed.TokenID)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, h.request(http.MethodGet, "/synthetic-token/probe", unstored, nil).Code)

	removeFailure()
	created := h.request(http.MethodPost, "/synthetic-token", h.tokens[0], []byte(`{"name":"synthetic companion","expiration":1}`))
	require.Equal(t, http.StatusOK, created.Code)
	assert.Equal(t, "no-store", created.Header().Get("Cache-Control"))
	var issued struct {
		Success bool   `json:"success"`
		Data    string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &issued))
	require.True(t, issued.Success)
	claims, err := auth.JwtValidateFastenToken(h.cfg.GetString("jwt.issuer.key"), issued.Data)
	require.NoError(t, err)
	require.Equal(t, "access", claims.TokenType)
	require.Equal(t, h.users[0].Username, claims.Subject)
	rows, err = h.repo.GetUserAccessTokens(h.owner)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, h.users[0].ID, rows[0].UserID)
	require.Equal(t, rows[0].TokenID, claims.ID)
	require.Equal(t, "synthetic companion", rows[0].Name)
	require.WithinDuration(t, time.Now().Add(24*time.Hour), rows[0].ExpiresAt, time.Minute)
	require.Equal(t, http.StatusNoContent, h.request(http.MethodGet, "/synthetic-token/probe", issued.Data, nil).Code)

	deletion, err := json.Marshal(DeleteTokenPayload{TokenID: rows[0].TokenID})
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, h.request(http.MethodPost, "/synthetic-token/delete", h.tokens[1], deletion).Code)
	require.Equal(t, http.StatusNoContent, h.request(http.MethodGet, "/synthetic-token/probe", issued.Data, nil).Code)
	require.Equal(t, http.StatusOK, h.request(http.MethodPost, "/synthetic-token/delete", h.tokens[0], deletion).Code)
	require.Equal(t, http.StatusUnauthorized, h.request(http.MethodGet, "/synthetic-token/probe", issued.Data, nil).Code)
}
