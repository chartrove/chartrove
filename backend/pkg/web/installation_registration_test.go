package web

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type installationSettingsProbe struct {
	database.DatabaseRepository
	settings models.SystemSettings
	loadErr  error
	saveErr  error
	saves    int
}

func (probe *installationSettingsProbe) LoadSystemSettings(context.Context) (*models.SystemSettings, error) {
	copy := probe.settings
	return &copy, probe.loadErr
}
func (probe *installationSettingsProbe) SaveSystemSettings(_ context.Context, settings *models.SystemSettings) error {
	probe.saves++
	if probe.saveErr != nil {
		return probe.saveErr
	}
	probe.settings = *settings
	return nil
}

type installationRegistrationTransport func(*http.Request) (*http.Response, error)

func (transport installationRegistrationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

type installationReplyBody struct {
	io.Reader
	readBytes int
	closed    bool
}

func (body *installationReplyBody) Read(buffer []byte) (int, error) {
	n, err := body.Reader.Read(buffer)
	body.readBytes += n
	return n, err
}
func (body *installationReplyBody) Close() error { body.closed = true; return nil }

func TestInstallationRegistrationRemoteContractAndPrivacy(t *testing.T) {
	const marker = "synthetic-private-installation-response"
	for _, test := range []struct {
		name   string
		status int
		body   string
		valid  bool
		secret string
	}{
		{"success", 200, `{"success":true,"data":{"installation_id":"synthetic-id","installation_secret":"synthetic-exact-secret"}}`, true, "synthetic-exact-secret"},
		{"optional secret", 200, `{"success":true,"data":{"installation_id":"synthetic-id"}}`, true, ""},
		{"unavailable", 503, marker, false, ""},
		{"rejected", 401, marker, false, ""},
		{"not confirmed", 200, `{"success":false,"error":"` + marker + `"}`, false, ""},
		{"missing id", 200, `{"success":true,"data":{"installation_secret":"` + marker + `"}}`, false, ""},
		{"malformed", 200, marker, false, ""},
		{"trailing data", 200, `{"success":true,"data":{"installation_id":"synthetic-id"}} {}`, false, ""},
		{"oversize", 200, strings.Repeat("x", maxInstallationRegistrationResponse+20), false, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			probe := &installationSettingsProbe{}
			var output bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&output)
			engine := AppEngine{Logger: logrus.NewEntry(logger), deviceRepo: probe}
			body := &installationReplyBody{Reader: strings.NewReader(test.body)}
			client := &http.Client{Transport: installationRegistrationTransport(func(request *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodPost, request.Method)
				require.Equal(t, "https://api.platform.fastenhealth.com/v1/installation/register", request.URL.String())
				require.Equal(t, "application/json", request.Header.Get("Content-Type"))
				deadline, ok := request.Context().Deadline()
				require.True(t, ok)
				require.WithinDuration(t, time.Now().Add(installationRegistrationTimeout), deadline, time.Second)
				return &http.Response{StatusCode: test.status, Status: marker, Header: make(http.Header), Body: body, Request: request}, nil
			})}
			err := engine.setupInstallationRegistration(client)
			require.True(t, body.closed)
			require.LessOrEqual(t, body.readBytes, maxInstallationRegistrationResponse+1)
			require.Zero(t, client.Timeout, "do not mutate the shared caller client")
			require.NotContains(t, output.String(), marker)
			require.NotContains(t, output.String(), "synthetic-id")
			require.NotContains(t, output.String(), "synthetic-exact-secret")
			if test.valid {
				require.NoError(t, err)
				require.Equal(t, 1, probe.saves)
				require.Equal(t, models.SystemSettings{InstallationID: "synthetic-id", InstallationSecret: test.secret}, probe.settings)
				return
			}
			require.ErrorIs(t, err, errOptionalInstallationRegistration)
			require.NotContains(t, err.Error(), marker)
			require.Zero(t, probe.saves)
			require.Empty(t, probe.settings.InstallationID)
			if test.status != http.StatusOK {
				require.Zero(t, body.readBytes, "rejected status bodies need not be read at all")
			}
		})
	}
}

func TestInstallationRegistrationKeepsStorageFailuresCriticalAndCachedIdentityLocal(t *testing.T) {
	failure := errors.New("synthetic local settings failure")
	for _, stage := range []string{"load", "save", "cached"} {
		t.Run(stage, func(t *testing.T) {
			probe := &installationSettingsProbe{}
			if stage == "load" {
				probe.loadErr = failure
			} else if stage == "save" {
				probe.saveErr = failure
			} else {
				probe.settings = models.SystemSettings{InstallationID: "synthetic-cached-id", InstallationSecret: "synthetic-cached-secret"}
			}
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			engine := AppEngine{Logger: logrus.NewEntry(logger), deviceRepo: probe}
			calls := 0
			client := &http.Client{Transport: installationRegistrationTransport(func(request *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"success":true,"data":{"installation_id":"synthetic-new-id","installation_secret":"synthetic-new-secret"}}`)), Request: request}, nil
			})}
			err := engine.setupInstallationRegistration(client)
			if stage == "cached" {
				require.NoError(t, err)
				require.Zero(t, calls)
				require.Zero(t, probe.saves)
				return
			}
			require.ErrorIs(t, err, failure)
			require.NotContains(t, err.Error(), failure.Error(), "startup diagnostics must not echo a potentially credential-bearing storage cause")
			require.NotErrorIs(t, err, errOptionalInstallationRegistration)
			if stage == "load" {
				require.Zero(t, calls)
			} else {
				require.Equal(t, 1, calls)
				require.Equal(t, 1, probe.saves)
				require.Equal(t, models.SystemSettings{}, probe.settings, "a storage failure is not a saved installation identity")
			}
		})
	}
}

func TestInstallationRegistrationPreservesTighterClientTimeout(t *testing.T) {
	probe := &installationSettingsProbe{}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	engine := AppEngine{Logger: logrus.NewEntry(logger), deviceRepo: probe}
	client := &http.Client{Timeout: 40 * time.Millisecond, Transport: installationRegistrationTransport(func(request *http.Request) (*http.Response, error) {
		deadline, ok := request.Context().Deadline()
		require.True(t, ok)
		require.True(t, deadline.Before(time.Now().Add(time.Second)))
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	err := engine.setupInstallationRegistration(client)
	require.ErrorIs(t, err, errOptionalInstallationRegistration)
	require.Equal(t, 40*time.Millisecond, client.Timeout)
	require.Zero(t, probe.saves)
}
