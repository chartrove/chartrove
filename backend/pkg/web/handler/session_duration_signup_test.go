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
)

func TestSignupConfiguredSessionLifetime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, valid := range []bool{true, false} {
		name := "invalid lifetime prevents account creation"
		if valid {
			name = "configured lifetime reaches signup token"
		}
		t.Run(name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", "synthetic-session-key")
			cfg.Set("jwt.session.duration", "invalid-duration")
			ctrl := gomock.NewController(t)
			db := mock_database.NewMockDatabaseRepository(ctrl)
			if valid {
				cfg.Set("jwt.session.duration", "5m")
				db.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), true).DoAndReturn(func(_ interface{}, user *models.User, _ bool) error {
					user.Role = pkg.UserRoleUser
					require.Equal(t, "synthetic-signup", user.Username)
					return nil
				})
			}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set(pkg.ContextKeyTypeDatabase, db)
			ctx.Set(pkg.ContextKeyTypeConfig, cfg)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBufferString(`{"username":"synthetic-signup","password":"synthetic-password"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			start := time.Now()
			handler.AuthSignup(ctx)
			finish := time.Now()
			if !valid {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.JSONEq(t, `{"success":false,"error":"Invalid session duration configuration"}`, recorder.Body.String())
				return
			}
			require.Equal(t, http.StatusOK, recorder.Code)
			var response struct {
				Token string `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			claims, err := auth.JwtValidateFastenToken("synthetic-session-key", response.Token)
			require.NoError(t, err)
			require.Equal(t, "synthetic-signup", claims.Subject)
			require.Equal(t, pkg.UserRoleUser, claims.Role)
			require.WithinRange(t, claims.ExpiresAt.Time, start.Add(5*time.Minute).Add(-time.Second), finish.Add(5*time.Minute))
		})
	}
}
