package models

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-sources/clients/factory"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	definitionModels "github.com/fastenhealth/fasten-sources/definitions/models"
	sourcePkg "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/fastenhealth/fasten-sources/pkg/models/catalog"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Actual pinned SDK non-dynamic refresh, not test mode. Every HTTP request is
// intercepted in memory; no socket, provider discovery/account or database.
func TestLegacyProviderOAuthRejectionBodyPrivacy(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			const bodyMarker = "synthetic-private-legacy-provider-error-description"
			const statusMarker = "synthetic-private-provider-status-reason"
			var output bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&output)
			t.Cleanup(func() { log.SetOutput(previous) })
			logger := logrus.New()
			logger.SetLevel(logrus.DebugLevel)
			logger.SetOutput(&output)
			var bodies []*providerPrivacyBody
			client := &http.Client{Transport: providerPrivacyTransport(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "https://synthetic-provider.invalid/token", r.URL.String())
				require.NoError(t, r.ParseForm())
				require.Equal(t, "refresh_token", r.PostForm.Get("grant_type"))
				body := &providerPrivacyBody{Reader: strings.NewReader(`{"error":"invalid_grant","error_description":"` + bodyMarker + `"}`)}
				bodies = append(bodies, body)
				return &http.Response{StatusCode: status, Status: fmt.Sprintf("%d %s", status, statusMarker), Header: http.Header{"Content-Type": {"application/json"}}, Body: body, Request: r}, nil
			})}
			expiry := time.Now().Add(-time.Hour).Unix()
			credential := &syntheticOAuthCredential{&SourceCredential{
				ClientId: "synthetic-client", AccessToken: "synthetic-original-access",
				RefreshToken: "synthetic-original-refresh", ExpiresAt: expiry,
			}}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			_, err := factory.GetSourceClientWithDefinition(
				sourcePkg.FastenLighthouseEnvSandbox, ctx, logger, credential,
				&definitionModels.LighthouseSourceDefinition{PatientAccessEndpoint: &catalog.PatientAccessEndpoint{
					Url: "https://synthetic-provider.invalid/fhir", TokenEndpoint: "https://synthetic-provider.invalid/token",
				}}, sourceModels.WithHttpClient(client), WithProviderResponsePrivacy(), WithProviderRequestLifetime(),
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid_grant", "retain the standard safe failure classification")
			assert.NotContains(t, err.Error(), bodyMarker)
			assert.NotContains(t, err.Error(), statusMarker)
			assert.NotContains(t, output.String(), bodyMarker)
			assert.NotContains(t, output.String(), statusMarker)
			require.NotEmpty(t, bodies, "exercise actual SDK refresh, not a skipped test-mode exchange")
			for _, body := range bodies {
				assert.True(t, body.closed)
			}
			assert.Equal(t, "synthetic-original-access", credential.AccessToken)
			assert.Equal(t, "synthetic-original-refresh", credential.RefreshToken)
			assert.Equal(t, expiry, credential.ExpiresAt)
		})
	}
}

func TestLegacyProviderFHIRRejectionBodyPrivacy(t *testing.T) {
	const marker = "synthetic-private-fhir-operation-outcome"
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&output)
	logger.SetLevel(logrus.DebugLevel)
	body := &providerFailureTrackedBody{Reader: strings.NewReader(`{"resourceType":"OperationOutcome","issue":[{"diagnostics":"` + marker + `"}]}`)}
	calls := 0
	client := &http.Client{Transport: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		require.Equal(t, http.MethodGet, request.Method)
		require.Equal(t, "https://synthetic-provider.invalid/fhir/Patient/synthetic", request.URL.String())
		require.Equal(t, "Bearer synthetic-original-access", request.Header.Get("Authorization"))
		return &http.Response{StatusCode: http.StatusInternalServerError, Status: "500 " + marker, Header: http.Header{"Content-Type": {"application/fhir+json"}}, Body: body, Request: request}, nil
	})}
	credential := &syntheticOAuthCredential{&SourceCredential{AccessToken: "synthetic-original-access", ExpiresAt: time.Now().Add(time.Hour).Unix()}}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	sdk, err := factory.GetSourceClientWithDefinition(
		sourcePkg.FastenLighthouseEnvSandbox, ctx, logger, credential,
		&definitionModels.LighthouseSourceDefinition{PatientAccessEndpoint: &catalog.PatientAccessEndpoint{
			Url: "https://synthetic-provider.invalid/fhir", TokenEndpoint: "https://synthetic-provider.invalid/token",
		}}, sourceModels.WithHttpClient(client), WithProviderResponsePrivacy(), WithProviderRequestLifetime(),
	)
	require.NoError(t, err)
	var patient map[string]interface{}
	_, err = sdk.GetRequest("Patient/synthetic", &patient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
	require.NotContains(t, err.Error(), marker)
	require.NotContains(t, output.String(), marker)
	require.Nil(t, patient, "a rejected response is not a successful clinical projection")
	require.Equal(t, 1, calls)
	require.True(t, body.closed)
}
