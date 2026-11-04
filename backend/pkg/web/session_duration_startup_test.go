package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestStartRejectsInvalidSessionDurationBeforeStorage(t *testing.T) {
	for _, value := range []string{"0s", "-1h", "500ms", "invalid-duration"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", "synthetic-private-signing-key")
			cfg.Set("jwt.session.duration", value)
			cfg.Set("database.type", "synthetic-no-database")
			var output bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&output)
			engine := AppEngine{Config: cfg, Logger: logrus.NewEntry(logger)}
			err = engine.Start()
			require.ErrorContains(t, err, "jwt.session.duration")
			require.Empty(t, output.String(), "invalid lifetime must be rejected before storage initialization")
		})
	}
}

func TestEmbeddedRejectsInvalidSessionDurationBeforeStorage(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("jwt.issuer.key", "synthetic-private-signing-key")
	cfg.Set("jwt.session.duration", "invalid-duration")
	cfg.Set("database.type", "synthetic-no-database")
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&output)
	app, err := NewEmbeddedApplication(cfg, logrus.NewEntry(logger), event_bus.NewNoopEventBusServer(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	require.Nil(t, app)
	require.ErrorContains(t, err, "jwt.session.duration")
	require.Empty(t, output.String())
}

func TestEmbeddedConfiguredSessionsSurviveEncryptedReopen(t *testing.T) {
	t.Setenv("FASTEN_JWT_SESSION_DURATION", "")
	path := filepath.Join(t.TempDir(), "synthetic-sessions.db")
	key := strings.Repeat("synthetic-encryption", 3)
	cfg := embeddedConfig(t, path, true)
	cfg.Set("jwt.session.duration", "5m")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase(key))
	payload := `{"username":"synthetic-session-owner","password":"synthetic-password","join_mailing_list":false}`
	start := time.Now()
	response := embeddedRequest(t, app, http.MethodPost, "/api/auth/signup", payload)
	finish := time.Now()
	require.Equal(t, http.StatusOK, response.Code)
	var signup struct {
		Token string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &signup))
	claims, err := auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), signup.Token)
	require.NoError(t, err)
	require.WithinRange(t, claims.ExpiresAt.Time, start.Add(5*time.Minute).Add(-time.Second), finish.Add(5*time.Minute))
	require.NoError(t, app.Close())
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.False(t, bytes.HasPrefix(raw, []byte("SQLite format 3")), "synthetic account storage must remain encrypted")

	// A changed installation lifetime affects newly issued sessions. It does not
	// reinterpret the existing signed expiry or rotate persistent signing keys.
	reopenedConfig := embeddedConfig(t, path, true)
	reopenedConfig.Set("jwt.session.duration", "2h")
	reopened := openEmbedded(t, reopenedConfig)
	require.NoError(t, reopened.UnlockDatabase(key))
	account := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/secure/account/me", nil)
	request.Header.Set("Authorization", "Bearer "+signup.Token)
	reopened.ServeHTTP(account, request)
	require.Equal(t, http.StatusOK, account.Code)
	require.Contains(t, account.Body.String(), "synthetic-session-owner")
	start = time.Now()
	response = embeddedRequest(t, reopened, http.MethodPost, "/api/auth/signin", payload)
	finish = time.Now()
	require.Equal(t, http.StatusOK, response.Code)
	var signin struct {
		Token string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &signin))
	newClaims, err := auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), signin.Token)
	require.NoError(t, err)
	require.Equal(t, claims.Subject, newClaims.Subject)
	require.Equal(t, claims.Role, newClaims.Role)
	require.WithinRange(t, newClaims.ExpiresAt.Time, start.Add(2*time.Hour).Add(-time.Second), finish.Add(2*time.Hour))
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.GetString("jwt.issuer.key")))
	require.NoError(t, err)
	account = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/secure/account/me", nil)
	request.Header.Set("Authorization", "Bearer "+expired)
	reopened.ServeHTTP(account, request)
	require.Equal(t, http.StatusUnauthorized, account.Code)
}
