package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	mock_database "github.com/fastenhealth/fasten-onprem/backend/pkg/database/mock"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCreateReconnectSourceBindingErrorsDoNotEchoInput(t *testing.T) {
	const marker = "synthetic-private-invalid-date"
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&output)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
	ctx.Set(pkg.ContextKeyTypeDatabase, mock_database.NewMockDatabaseRepository(gomock.NewController(t)))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/source", bytes.NewBufferString(`{"created_at":"`+marker+`"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	CreateReconnectSource(ctx)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.NotEmpty(t, output.String(), "retain the parsing failure diagnostic")
	require.NotContains(t, output.String(), marker)
	require.NotContains(t, response.Body.String(), marker)
}

// Exercise the real handler, including accepted credential decoding and failed
// local catalog resolution, without storage or any external provider request.
func TestCreateReconnectSourceDoesNotLogCredentials(t *testing.T) {
	for _, level := range []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel} {
		t.Run(level.String(), func(t *testing.T) {
			canaries := []string{
				"synthetic-private-access", "synthetic-private-refresh", "synthetic-private-id-token",
				"synthetic-private-verifier", "synthetic-private-jwk", "synthetic-private-patient",
				"synthetic-private-display", "synthetic-private-user-password",
			}
			credential := models.SourceCredential{
				AccessToken: canaries[0], RefreshToken: canaries[1], IdToken: canaries[2],
				CodeVerifier: canaries[3], DynamicClientJWKS: []map[string]string{{"d": canaries[4]}},
				Patient: canaries[5], Display: canaries[6], User: &models.User{Password: canaries[7]},
			}
			body, err := json.Marshal(credential)
			require.NoError(t, err)
			var output bytes.Buffer
			logger := logrus.New()
			logger.SetLevel(level)
			logger.SetOutput(&output)
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logger))
			// No calls are permitted: missing endpoint must fail before storage.
			ctx.Set(pkg.ContextKeyTypeDatabase, mock_database.NewMockDatabaseRepository(gomock.NewController(t)))
			ctx.Request = httptest.NewRequest(http.MethodPost, "/source", bytes.NewReader(body))
			ctx.Request.Header.Set("Content-Type", "application/json")
			CreateReconnectSource(ctx)
			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, output.String(), "retrieving source definition", "retain actionable non-payload diagnostics")
			for _, canary := range canaries {
				require.NotContains(t, output.String(), canary, "credential and patient payloads must not enter logs")
				require.NotContains(t, response.Body.String(), canary)
			}
		})
	}
}
