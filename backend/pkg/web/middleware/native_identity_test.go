package middleware_test

import (
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Sign the old and new payloads directly, independently of the token generator.
func TestRequireAuthNativeIdentityAndLegacyReauthentication(t *testing.T) {
	userID := uuid.New()
	issued := time.Now().Add(-time.Minute).Truncate(time.Second)
	for _, tc := range []struct {
		name     string
		nativeID string
		created  time.Time
		iat      *jwt.NumericDate
		allowed  bool
	}{
		{"bound account", userID.String(), time.Now(), jwt.NewNumericDate(issued), true},
		{"different UUID", uuid.NewString(), issued.Add(-time.Hour), jwt.NewNumericDate(issued), false},
		{"malformed UUID", "invalid", issued.Add(-time.Hour), jwt.NewNumericDate(issued), false},
		{"legacy established account", "", issued.Add(-time.Hour), jwt.NewNumericDate(issued), false},
		{"legacy replacement account", "", issued.Add(time.Second), jwt.NewNumericDate(issued), false},
		{"legacy same second", "", issued.Add(time.Nanosecond), jwt.NewNumericDate(issued), false},
		{"legacy equal timestamp", "", issued, jwt.NewNumericDate(issued), false},
		{"legacy missing issued at", "", issued.Add(-time.Hour), nil, false},
		{"legacy unknown creation", "", time.Time{}, jwt.NewNumericDate(issued), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", "synthetic-identity-key")
			repo := mock_database.NewMockDatabaseRepository(gomock.NewController(t))
			repo.EXPECT().GetUserByUsername(gomock.Any(), "synthetic-user").Return(&models.User{ModelBase: models.ModelBase{ID: userID, CreatedAt: tc.created}, Username: "synthetic-user"}, nil)
			claims := jwt.MapClaims{"sub": "synthetic-user", "exp": time.Now().Add(time.Hour).Unix()}
			if tc.nativeID != "" {
				claims["native_user_id"] = tc.nativeID
			}
			if tc.iat != nil {
				claims["iat"] = tc.iat.Unix()
			}
			signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("synthetic-identity-key"))
			require.NoError(t, err)
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set(pkg.ContextKeyTypeConfig, cfg); c.Set(pkg.ContextKeyTypeDatabase, repo) }, middleware.RequireAuth())
			router.GET("/protected", func(c *gin.Context) { c.Status(204) })
			r := httptest.NewRequest("GET", "/protected", nil)
			r.Header.Set("Authorization", "Bearer "+signed)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			want := http.StatusUnauthorized
			if tc.allowed {
				want = http.StatusNoContent
			}
			require.Equal(t, want, w.Code)
		})
	}
}

func TestNativeSessionGeneratorIncludesImmutableUUID(t *testing.T) {
	user := models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: "synthetic-user"}
	for _, kind := range []string{"password", "access", "oidc"} {
		t.Run(kind, func(t *testing.T) {
			var token string
			var err error
			switch kind {
			case "password":
				token, err = auth.JwtGenerateFastenTokenFromUser(user, "synthetic-identity-key")
			case "access":
				token, err = auth.JwtGenerateAccessToken(user, "synthetic-identity-key", time.Now().Add(time.Hour), "synthetic-token")
			case "oidc":
				token, err = auth.JwtGenerateOIDCSession(user, "synthetic-identity-key", time.Now().Add(time.Hour), uuid.NewString())
			}
			require.NoError(t, err)
			var payload jwt.MapClaims
			_, _, err = jwt.NewParser().ParseUnverified(token, &payload)
			require.NoError(t, err)
			require.Equal(t, user.ID.String(), payload["native_user_id"])
		})
	}
}
