package models

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-sources/clients/factory"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	definitionModels "github.com/fastenhealth/fasten-sources/definitions/models"
	sourcePkg "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/fastenhealth/fasten-sources/pkg/models/catalog"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLegacyProviderRequestConstructorLifetime(t *testing.T) {
	for _, canceled := range []bool{true, false} {
		name := "deadline and complete success"
		if canceled {
			name = "canceled before request"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			parentDeadline, _ := ctx.Deadline()
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			calls := 0
			observedDeadline := false
			client := &http.Client{Transport: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
				calls++
				deadline, ok := request.Context().Deadline()
				observedDeadline = ok && !deadline.After(parentDeadline)
				return &http.Response{StatusCode: 200, Status: "200 OK", Header: http.Header{"Content-Type": {"application/fhir+json"}}, Body: io.NopCloser(strings.NewReader(`{"resourceType":"Patient","id":"synthetic"}`)), Request: request}, nil
			})}
			sdk, err := factory.GetSourceClientWithDefinition(sourcePkg.FastenLighthouseEnvSandbox, ctx, logger,
				&syntheticOAuthCredential{&SourceCredential{AccessToken: "synthetic-lifetime-only-access", ExpiresAt: time.Now().Add(time.Hour).Unix()}},
				&definitionModels.LighthouseSourceDefinition{PatientAccessEndpoint: &catalog.PatientAccessEndpoint{Url: "https://synthetic-provider.invalid/fhir", TokenEndpoint: "https://synthetic-provider.invalid/token"}},
				sourceModels.WithHttpClient(client), WithProviderResponsePrivacy(), WithProviderRequestLifetime())
			require.NoError(t, err)
			if canceled {
				cancel()
			}
			var patient map[string]any
			_, err = sdk.GetRequest("Patient/synthetic", &patient)
			if canceled {
				require.ErrorIs(t, err, context.Canceled)
				require.Zero(t, calls, "a canceled constructor must not start provider network work")
				require.Nil(t, patient)
				return
			}
			require.NoError(t, err, "do not cancel the body when RoundTrip returns")
			require.Equal(t, 1, calls)
			require.Equal(t, map[string]any{"resourceType": "Patient", "id": "synthetic"}, patient)
			require.True(t, observedDeadline, "SDK-generated request must retain the constructor deadline")
		})
	}
}

type providerLifetimeBlockingBody struct {
	ctx         context.Context
	started     chan struct{}
	force       <-chan struct{}
	release     func()
	startedOnce sync.Once
	liveAtRead  atomic.Bool
	closed      atomic.Bool
}

func (body *providerLifetimeBlockingBody) Read([]byte) (int, error) {
	body.startedOnce.Do(func() {
		body.liveAtRead.Store(body.ctx.Err() == nil)
		close(body.started)
	})
	select {
	case <-body.ctx.Done():
		return 0, body.ctx.Err()
	case <-body.force:
		return 0, io.EOF
	}
}
func (body *providerLifetimeBlockingBody) Close() error {
	body.closed.Store(true)
	body.release()
	return nil
}

func TestLegacyProviderResponseReadConstructorCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	force := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(force) }) }
	started := make(chan struct{})
	var body *providerLifetimeBlockingBody
	client := &http.Client{Transport: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
		body = &providerLifetimeBlockingBody{ctx: request.Context(), started: started, force: force, release: release}
		return &http.Response{StatusCode: 200, Status: "200 OK", Header: http.Header{"Content-Type": {"application/fhir+json"}}, Body: body, Request: request}, nil
	})}
	sdk, err := factory.GetSourceClientWithDefinition(sourcePkg.FastenLighthouseEnvSandbox, ctx, logger,
		&syntheticOAuthCredential{&SourceCredential{AccessToken: "synthetic-lifetime-only-access", ExpiresAt: time.Now().Add(time.Hour).Unix()}},
		&definitionModels.LighthouseSourceDefinition{PatientAccessEndpoint: &catalog.PatientAccessEndpoint{Url: "https://synthetic-provider.invalid/fhir", TokenEndpoint: "https://synthetic-provider.invalid/token"}},
		sourceModels.WithHttpClient(client), WithProviderResponsePrivacy(), WithProviderRequestLifetime())
	require.NoError(t, err)
	finished := make(chan struct{})
	var requestError error
	go func() {
		var patient map[string]any
		_, requestError = sdk.GetRequest("Patient/synthetic", &patient)
		close(finished)
	}()
	t.Cleanup(func() {
		release()
		select {
		case <-finished:
		case <-time.After(2 * time.Second):
			t.Error("owned synthetic SDK request did not finish after fixture release")
		}
	})
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("synthetic response read did not start")
	}
	require.True(t, body.liveAtRead.Load(), "RoundTrip return must not prematurely cancel response consumption")
	cancel()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Error("SDK response read ignored constructor cancellation")
		release() // Controlled negative must not strand a request goroutine.
		<-finished
	}
	require.ErrorIs(t, requestError, context.Canceled)
	require.True(t, body.closed.Load())
}
