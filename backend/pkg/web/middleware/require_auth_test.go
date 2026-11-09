package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	mock_config "github.com/fastenhealth/fasten-onprem/backend/pkg/config/mock"
	mock_database "github.com/fastenhealth/fasten-onprem/backend/pkg/database/mock"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRequireAuthRequiresActiveImmutableAccount(t *testing.T) {
	for _, deleted := range []bool{false, true} {
		name := "active"
		if deleted {
			name = "soft-deleted"
		}
		t.Run(name, func(t *testing.T) {
			user := &models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: "synthetic-native-owner"}
			if deleted {
				now := time.Now().UTC()
				user.DeletedAt = &now
			}
			cfg := mock_config.NewMockInterface(gomock.NewController(t))
			cfg.EXPECT().GetString("jwt.issuer.key").Return("synthetic-signing-key")
			repo := mock_database.NewMockDatabaseRepository(gomock.NewController(t))
			repo.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
			token, err := auth.JwtGenerateFastenTokenFromUser(*user, "synthetic-signing-key")
			require.NoError(t, err)
			called := false
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set(pkg.ContextKeyTypeConfig, cfg); c.Set(pkg.ContextKeyTypeDatabase, repo) }, middleware.RequireAuth())
			router.GET("/protected", func(c *gin.Context) { called = true; c.Status(http.StatusNoContent) })
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if deleted {
				require.Equal(t, http.StatusUnauthorized, response.Code)
				require.False(t, called, "immutable identity alone cannot revive a soft-deleted account")
			} else {
				require.Equal(t, http.StatusNoContent, response.Code)
				require.True(t, called)
			}
		})
	}
}

func TestRequireAuthRejectsInvalidTokenBeforeProtectedHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nativeID := uuid.New()
	for _, tt := range []struct {
		name   string
		key    string
		expiry *jwt.NumericDate
		status int
	}{
		{"valid", "synthetic-signing-key", jwt.NewNumericDate(time.Now().Add(time.Hour)), http.StatusNoContent},
		{"missing expiry", "synthetic-signing-key", nil, http.StatusUnauthorized},
		{"expired", "synthetic-signing-key", jwt.NewNumericDate(time.Now().Add(-time.Hour)), http.StatusUnauthorized},
		{"empty signing key", "", jwt.NewNumericDate(time.Now().Add(time.Hour)), http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := mock_config.NewMockInterface(gomock.NewController(t))
			config.EXPECT().GetString("jwt.issuer.key").Return(tt.key)
			repo := mock_database.NewMockDatabaseRepository(gomock.NewController(t))
			if tt.status == http.StatusNoContent {
				repo.EXPECT().GetUserByUsername(gomock.Any(), "synthetic-user").Return(&models.User{ModelBase: models.ModelBase{ID: nativeID}, Username: "synthetic-user"}, nil)
			}
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.UserRegisteredClaims{NativeUserID: nativeID.String(), RegisteredClaims: jwt.RegisteredClaims{
				Subject: "synthetic-user", ExpiresAt: tt.expiry,
			}}).SignedString([]byte(tt.key))
			require.NoError(t, err)

			called := false
			router := gin.New()
			router.Use(gin.Recovery(), func(c *gin.Context) { c.Set(pkg.ContextKeyTypeConfig, config); c.Set(pkg.ContextKeyTypeDatabase, repo) }, middleware.RequireAuth())
			router.GET("/protected", func(c *gin.Context) {
				called = true
				require.Equal(t, "synthetic-user", c.GetString(pkg.ContextKeyTypeAuthUsername))
				require.Equal(t, nativeID, c.MustGet(pkg.ContextKeyTypeAuthUserID))
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, tt.status, response.Code)
			require.Equal(t, tt.status == http.StatusNoContent, called)
		})
	}
}
