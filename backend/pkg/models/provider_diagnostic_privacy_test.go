package models

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
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

// A real pinned-SDK failed FHIR read with both accepted transport guards. Every
// request is intercepted in memory; the source location itself is synthetic.
func TestProviderSDKRequestLocationDiagnosticsArePrivate(t *testing.T) {
	const marker = "synthetic-private-sdk-request-location"
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(&output)
	calls := 0
	body := &providerDiagnosticBody{Reader: strings.NewReader(`{"resourceType":"OperationOutcome"}`)}
	client := &http.Client{Transport: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		require.Equal(t, "/fhir/Patient/"+marker, request.URL.Path)
		require.Equal(t, marker, request.URL.Query().Get("subject.name"))
		require.Equal(t, "Bearer synthetic-diagnostic-only-access", request.Header.Get("Authorization"))
		return &http.Response{StatusCode: 500, Status: "500 Internal Server Error", Header: http.Header{"Content-Type": {"application/fhir+json"}}, Body: body, Request: request}, nil
	})}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	sdk, err := factory.GetSourceClientWithDefinition(sourcePkg.FastenLighthouseEnvSandbox, ctx, providerDiagnosticLogger(logger),
		&syntheticOAuthCredential{&SourceCredential{AccessToken: "synthetic-diagnostic-only-access", ExpiresAt: time.Now().Add(time.Hour).Unix()}},
		&definitionModels.LighthouseSourceDefinition{PatientAccessEndpoint: &catalog.PatientAccessEndpoint{Url: "https://synthetic-provider.invalid/fhir", TokenEndpoint: "https://synthetic-provider.invalid/token"}},
		sourceModels.WithHttpClient(client), WithProviderResponsePrivacy(), WithProviderRequestLifetime())
	require.NoError(t, err)
	sdk = &providerDiagnosticClient{SourceClient: sdk}
	var patient map[string]any
	_, err = sdk.GetRequest("Patient/"+marker+"?subject.name="+marker, &patient)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), marker, "outward operation errors must not expose provider request locations")
	assert.NotContains(t, output.String(), marker, "SDK-instance debug diagnostics must not expose provider request locations")
	require.Equal(t, 1, calls)
	require.Nil(t, patient)
	require.True(t, body.closed)
}

type providerDiagnosticBody struct {
	io.Reader
	closed bool
}

func (body *providerDiagnosticBody) Close() error { body.closed = true; return nil }

func TestProviderSDKWarningIsPrivateWithoutChangingSuccessfulBinary(t *testing.T) {
	const marker = "synthetic-private-sdk-binary-location"
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(&output)
	payload := []byte("synthetic exact binary:\x00 1.234567890123456789")
	body := &providerDiagnosticBody{Reader: bytes.NewReader(payload)}
	client := &http.Client{Transport: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Status: "200 OK", Header: http.Header{"Content-Type": {"application/x-" + marker}}, Body: body, Request: request}, nil
	})}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	sdk, err := factory.GetSourceClientWithDefinition(sourcePkg.FastenLighthouseEnvSandbox, ctx, providerDiagnosticLogger(logger),
		&syntheticOAuthCredential{&SourceCredential{AccessToken: "synthetic-diagnostic-only-access", ExpiresAt: time.Now().Add(time.Hour).Unix()}},
		&definitionModels.LighthouseSourceDefinition{PatientAccessEndpoint: &catalog.PatientAccessEndpoint{Url: "https://synthetic-provider.invalid/fhir", TokenEndpoint: "https://synthetic-provider.invalid/token"}},
		sourceModels.WithHttpClient(client), WithProviderResponsePrivacy(), WithProviderRequestLifetime())
	require.NoError(t, err)
	sdk = &providerDiagnosticClient{SourceClient: sdk}
	var binary map[string]any
	location, err := sdk.GetRequest("Binary/"+marker, &binary)
	require.NoError(t, err)
	require.Equal(t, "https://synthetic-provider.invalid/fhir/Binary/"+marker, location)
	require.Equal(t, "Binary", binary["resourceType"])
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte(location)), binary["id"])
	require.Equal(t, base64.StdEncoding.EncodeToString(payload), binary["data"])
	require.Equal(t, "application/x-"+marker, binary["contentType"])
	require.True(t, body.closed)
	require.Contains(t, output.String(), `"level":"warning"`, "actual SDK warning remains operationally visible")
	require.Contains(t, output.String(), "provider SDK diagnostic")
	require.NotContains(t, output.String(), marker)
}

func TestProviderProductionFactoryOwnsManualDiagnostics(t *testing.T) {
	const marker = "synthetic-private-manual-file"
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(&output)
	credential := &SourceCredential{PlatformType: sourcePkg.PlatformTypeManual}
	client, err := NewProviderSourceClient(sourcePkg.FastenLighthouseEnvSandbox, t.Context(), logger, credential)
	require.NoError(t, err)
	require.IsType(t, &providerDiagnosticClient{}, client)
	require.Same(t, credential, client.GetSourceCredential())
	file, err := os.CreateTemp(t.TempDir(), marker+"-*.json")
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	_, err = file.WriteString(`{"malformed":"` + marker + `"`)
	require.NoError(t, err)
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)
	_, _, err = client.ExtractPatientId(file)
	require.Error(t, err)
	require.IsType(t, &providerOperationError{}, err)
	require.NotContains(t, err.Error(), marker)
	require.NotContains(t, output.String(), marker)
}
