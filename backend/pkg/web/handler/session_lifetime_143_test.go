package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	mock_database "github.com/fastenhealth/fasten-onprem/backend/pkg/database/mock"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/handler"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthSessionConfiguredLifetime(t *testing.T) {
	t.Setenv("FASTEN_JWT_SESSION_DURATION", "")
	gin.SetMode(gin.TestMode)
	const signingKey = "synthetic-session-signing-key"
	password, err := bcrypt.GenerateFromPassword([]byte("synthetic-password"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{Username: "synthetic-session-user", Password: string(password), Role: pkg.UserRoleUser}
	for _, tc := range []struct {
		name, duration string
		want           time.Duration
		status         int
	}{
		{"default remains one hour", "", time.Hour, http.StatusOK},
		{"shorter configured session", "5m", 5 * time.Minute, http.StatusOK},
		{"longer configured session", "2h", 2 * time.Hour, http.StatusOK},
		{"zero rejected", "0s", 0, http.StatusInternalServerError},
		{"negative rejected", "-5m", 0, http.StatusInternalServerError},
		{"malformed rejected", "invalid-duration", 0, http.StatusInternalServerError},
		{"subsecond rejected", "1ns", 0, http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", signingKey)
			if tc.duration != "" {
				cfg.Set("jwt.session.duration", tc.duration)
			}
			ctrl := gomock.NewController(t)
			db := mock_database.NewMockDatabaseRepository(ctrl)
			db.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set(pkg.ContextKeyTypeDatabase, db)
			ctx.Set(pkg.ContextKeyTypeConfig, cfg)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBufferString(`{"username":"synthetic-session-user","password":"synthetic-password"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			start := time.Now()
			handler.AuthSignin(ctx)
			finish := time.Now()
			require.Equal(t, tc.status, recorder.Code)
			if tc.status != http.StatusOK {
				require.JSONEq(t, `{"success":false,"error":"Invalid session duration configuration"}`, recorder.Body.String())
				return
			}
			var response struct {
				Success bool   `json:"success"`
				Token   string `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.True(t, response.Success)
			claims, err := auth.JwtValidateFastenToken(signingKey, response.Token)
			require.NoError(t, err)
			require.Equal(t, user.Username, claims.Subject)
			require.Equal(t, "docker-fastenhealth", claims.Issuer)
			require.Equal(t, pkg.UserRoleUser, claims.Role)
			require.Empty(t, claims.TokenType)
			require.Empty(t, claims.ID)
			require.WithinRange(t, claims.ExpiresAt.Time, start.Add(tc.want).Add(-time.Second), finish.Add(tc.want))
		})
	}
}
