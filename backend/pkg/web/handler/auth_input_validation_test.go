package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestAuthSignupRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body, message string
		username, password  string
	}{
		{name: "empty object", body: `{}`, message: "Username and password are required"},
		{name: "null", body: `null`, message: "Username and password are required"},
		{name: "wizard fields only", body: `{"join_mailing_list":false}`, message: "Username and password are required"},
		{name: "empty body", body: ``, message: "Invalid sign-up request"},
		{name: "malformed JSON", body: `{`, message: "Invalid sign-up request"},
		{name: "array instead of object", body: `[]`, message: "Invalid sign-up request"},
		{name: "wrong credential type", body: `{"username":12,"password":"synthetic-password"}`, message: "Invalid sign-up request"},
		{name: "username absent", body: `{"password":"synthetic-password"}`, message: "Username and password are required"},
		{name: "password absent", body: `{"username":"synthetic-user"}`, message: "Username and password are required"},
		{name: "empty username", body: `{"username":"","password":"synthetic-password"}`, message: "Username and password are required"},
		{name: "whitespace username", body: `{"username":"  ","password":"synthetic-password"}`, message: "Username and password are required"},
		{name: "whitespace password", body: `{"username":"synthetic-user","password":"  "}`, message: "Username and password are required"},
		{name: "service name", body: `{"username":"support","password":"synthetic-password"}`, message: "Username is reserved; choose another name"},
		{name: "case and padding", body: `{"username":"  SyStEm  ","password":"synthetic-password"}`, message: "Username is reserved; choose another name"},
		{name: "product name", body: `{"username":"chartrove","password":"synthetic-password"}`, message: "Username is reserved; choose another name"},
		{name: "existing maintainer name remains available", body: `{"username":"admin","password":"synthetic-password","role":"admin"}`, username: "admin", password: "synthetic-password"},
		{name: "nonblank literal credentials are preserved", body: `{"username":" Synthetic User ","password":" synthetic password "}`, username: " Synthetic User ", password: " synthetic password "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", "synthetic-input-test-key")
			cfg.Set("jwt.session.duration", "1h")
			ctrl := gomock.NewController(t)
			db := mock_database.NewMockDatabaseRepository(ctrl)
			countReads, creations := 0, 0
			var received *models.User
			db.EXPECT().GetUserCount(gomock.Any()).DoAndReturn(func(_ interface{}) (int, error) {
				countReads++
				return 1, nil
			}).AnyTimes()
			db.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, user *models.User) error {
				creations++
				copy := *user
				received = &copy
				return nil
			}).AnyTimes()
			var recovery bytes.Buffer
			router := gin.New()
			router.Use(gin.RecoveryWithWriter(&recovery))
			router.POST("/signup", func(c *gin.Context) {
				c.Set(pkg.ContextKeyTypeDatabase, db)
				c.Set(pkg.ContextKeyTypeConfig, cfg)
				handler.AuthSignup(c)
			})
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(tc.body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(response, request)
			if tc.message != "" {
				require.Equal(t, http.StatusBadRequest, response.Code)
				expected, err := json.Marshal(map[string]interface{}{"success": false, "error": tc.message})
				require.NoError(t, err)
				require.JSONEq(t, string(expected), response.Body.String())
				require.Empty(t, recovery.String(), "bad input must not rely on panic recovery")
				require.Zero(t, countReads)
				require.Zero(t, creations)
				return
			}
			require.Equal(t, http.StatusOK, response.Code)
			require.Empty(t, recovery.String())
			require.Equal(t, 1, countReads)
			require.Equal(t, 1, creations)
			require.Equal(t, tc.username, received.Username)
			require.Equal(t, tc.password, received.Password)
			require.Equal(t, pkg.UserRoleUser, received.Role, "role follows actual account policy, not supplied role or username")
			var result struct {
				Token string `json:"data"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
			claims, err := auth.JwtValidateFastenToken("synthetic-input-test-key", result.Token)
			require.NoError(t, err)
			require.Equal(t, tc.username, claims.Subject)
			require.Equal(t, pkg.UserRoleUser, claims.Role)
		})
	}
}
