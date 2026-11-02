package middleware_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	mock_database "github.com/fastenhealth/fasten-onprem/backend/pkg/database/mock"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRequireAuthChecksCurrentAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name     string
		user     *models.User
		err      error
		access   bool
		token    *models.AccessToken
		tokenErr error
		status   int
	}{
		{"current session", &models.User{Username: "synthetic-user"}, nil, false, nil, nil, 204},
		{"deleted account", nil, gorm.ErrRecordNotFound, false, nil, nil, 401},
		{"wrapped missing account", nil, fmt.Errorf("lookup: %w", gorm.ErrRecordNotFound), false, nil, nil, 401},
		{"nil account", nil, nil, false, nil, nil, 401},
		{"database unavailable", nil, errors.New("synthetic database failure"), false, nil, nil, 500},
		{"current access token", &models.User{Username: "synthetic-user"}, nil, true, &models.AccessToken{}, nil, 204},
		{"access token missing account", nil, gorm.ErrRecordNotFound, true, nil, nil, 401},
		{"revoked access token", &models.User{Username: "synthetic-user"}, nil, true, nil, gorm.ErrRecordNotFound, 401},
		{"nil access token", &models.User{Username: "synthetic-user"}, nil, true, nil, nil, 401},
		{"access token database unavailable", &models.User{Username: "synthetic-user"}, nil, true, nil, errors.New("synthetic database failure"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", "synthetic-signing-key")
			repo := mock_database.NewMockDatabaseRepository(gomock.NewController(t))
			repo.EXPECT().GetUserByUsername(gomock.Any(), "synthetic-user").Return(tc.user, tc.err)
			if tc.access && tc.user != nil && tc.err == nil {
				repo.EXPECT().GetAccessTokenByTokenIDAndUsername(gomock.Any(), "synthetic-token", "synthetic-user").Return(tc.token, tc.tokenErr)
			}
			var token string
			if tc.access {
				token, err = auth.JwtGenerateAccessToken(models.User{Username: "synthetic-user"}, "synthetic-signing-key", time.Now().Add(time.Hour), "synthetic-token")
			} else {
				token, err = auth.JwtGenerateFastenTokenFromUser(models.User{Username: "synthetic-user"}, "synthetic-signing-key")
			}
			require.NoError(t, err)
			called := false
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set(pkg.ContextKeyTypeConfig, cfg); c.Set(pkg.ContextKeyTypeDatabase, repo) }, middleware.RequireAuth())
			router.GET("/protected", func(c *gin.Context) { called = true; c.Status(http.StatusNoContent) })
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, tc.status, response.Code)
			require.Equal(t, tc.status == http.StatusNoContent, called)
			require.NotContains(t, response.Body.String(), "synthetic database failure")
		})
	}
}

func TestRequireAuthRejectsMalformedHeadersWithoutLoggingCredentials(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("jwt.issuer.key", "synthetic-signing-key")
	validToken, err := auth.JwtGenerateFastenTokenFromUser(models.User{Username: "synthetic-user"}, "synthetic-signing-key")
	require.NoError(t, err)
	var output bytes.Buffer
	original := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(original) })
	for _, header := range []string{"Bearer synthetic-sensitive-token extra", "Basic synthetic-sensitive-token", "synthetic-sensitive-token", "", "Basic " + validToken} {
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set(pkg.ContextKeyTypeConfig, cfg) }, middleware.RequireAuth())
		called := false
		router.GET("/protected", func(c *gin.Context) { called = true; c.Status(204) })
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.Header.Set("Authorization", header)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, 401, response.Code)
		require.False(t, called)
	}
	require.NotContains(t, output.String(), "synthetic-sensitive-token")
}
