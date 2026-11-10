package models

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// Exercise AfterFunc registration/stop directly, without noisy process-wide
// goroutine counts. Hiding Value prevents the standard library bypassing this
// test-only AfterFunc implementation via its private cancel-context key.
type providerTrackedLifetime struct {
	context.Context
	active atomic.Int32
}

func (parent *providerTrackedLifetime) Value(any) any { return nil }
func (parent *providerTrackedLifetime) AfterFunc(callback func()) func() bool {
	parent.active.Add(1)
	var once sync.Once
	finished := func() { once.Do(func() { parent.active.Add(-1) }) }
	stop := context.AfterFunc(parent.Context, func() { finished(); callback() })
	return func() bool {
		stopped := stop()
		if stopped {
			finished()
		}
		return stopped
	}
}

type providerLifetimeCloseProbe struct {
	io.Reader
	closeErr error
	closes   int
}

func (body *providerLifetimeCloseProbe) Close() error { body.closes++; return body.closeErr }

func TestProviderLifetimeResponseConsumptionAndCleanup(t *testing.T) {
	for _, finish := range []string{"EOF", "close", "transport error", "empty body"} {
		t.Run(finish, func(t *testing.T) {
			baseParent, cancel := context.WithCancel(t.Context())
			defer cancel()
			parent := &providerTrackedLifetime{Context: baseParent}
			const exact = `{"valueQuantity":{"value":1.234567890123456789},"extension":[{"valueString":"unchanged"}]}`
			closeFailure := errors.New("synthetic close failure")
			body := &providerLifetimeCloseProbe{Reader: strings.NewReader(exact), closeErr: closeFailure}
			transportFailure := errors.New("synthetic unchanged transport failure")
			var seen context.Context
			var original *http.Response
			boundary := providerRequestLifetimeTransport{parent: parent, base: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
				seen = request.Context()
				require.EqualValues(t, 1, parent.active.Load())
				if finish == "transport error" {
					return nil, transportFailure
				}
				original = &http.Response{StatusCode: 200, Status: "200 unchanged", Header: http.Header{"X-Synthetic": {"retained"}}, Body: body, ContentLength: int64(len(exact)), Request: request}
				if finish == "empty body" {
					original.Body = http.NoBody
				}
				return original, nil
			})}
			request, err := http.NewRequest(http.MethodGet, "https://synthetic-provider.invalid/fhir", nil)
			require.NoError(t, err)
			response, err := boundary.RoundTrip(request)
			if finish == "transport error" {
				require.Same(t, transportFailure, err)
				require.Nil(t, response)
				require.Zero(t, parent.active.Load())
				return
			}
			require.NoError(t, err)
			if finish == "empty body" {
				require.Same(t, original, response)
				require.Zero(t, parent.active.Load())
				return
			}
			require.NotSame(t, original, response, "do not mutate a transport-owned response")
			require.Same(t, body, original.Body)
			require.NoError(t, seen.Err(), "headers alone do not finish the request lifetime")
			require.EqualValues(t, 1, parent.active.Load())
			require.Equal(t, original.Header, response.Header)
			require.Equal(t, original.ContentLength, response.ContentLength)
			if finish == "EOF" {
				payload, err := io.ReadAll(response.Body)
				require.NoError(t, err)
				require.Equal(t, exact, string(payload))
				require.Zero(t, parent.active.Load(), "EOF unregisters the constructor callback")
			}
			require.Same(t, closeFailure, response.Body.Close())
			require.Equal(t, 1, body.closes)
			require.Zero(t, parent.active.Load(), "Close unregisters even when the underlying close fails")
			require.ErrorIs(t, seen.Err(), context.Canceled)
			require.NoError(t, baseParent.Err(), "release never cancels the caller")
		})
	}
}

func TestProviderLifetimeKeepsBothDeadlinesCancellationAndRequestValues(t *testing.T) {
	type key string
	for _, parentEarlier := range []bool{true, false} {
		parentDuration, requestDuration := time.Minute, 2*time.Minute
		if !parentEarlier {
			parentDuration, requestDuration = requestDuration, parentDuration
		}
		parent, cancelParent := context.WithTimeout(context.WithValue(t.Context(), key("constructor-only"), "not-copied"), parentDuration)
		requestContext, cancelRequest := context.WithTimeout(context.WithValue(t.Context(), key("request-only"), "retained"), requestDuration)
		parentDeadline, _ := parent.Deadline()
		requestDeadline, _ := requestContext.Deadline()
		earliest := parentDeadline
		if requestDeadline.Before(earliest) {
			earliest = requestDeadline
		}
		var joined context.Context
		boundary := providerRequestLifetimeTransport{parent: parent, base: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
			joined = request.Context()
			deadline, ok := joined.Deadline()
			require.True(t, ok)
			require.Equal(t, earliest, deadline)
			require.Equal(t, "retained", joined.Value(key("request-only")))
			require.Nil(t, joined.Value(key("constructor-only")))
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("unchanged")), Request: request}, nil
		})}
		request, err := http.NewRequestWithContext(requestContext, http.MethodGet, "https://synthetic-provider.invalid", nil)
		require.NoError(t, err)
		response, err := boundary.RoundTrip(request)
		require.NoError(t, err)
		require.Same(t, requestContext, request.Context(), "never replace the original request context in place")
		if parentEarlier {
			cancelParent()
		} else {
			cancelRequest()
		}
		select {
		case <-joined.Done():
		case <-time.After(time.Second):
			t.Fatal("joined request did not observe cancellation")
		}
		require.ErrorIs(t, joined.Err(), context.Canceled)
		require.NoError(t, response.Body.Close())
		cancelParent()
		cancelRequest()
	}
}

func TestProviderLifetimeOptionCopiesClientsWithoutInventingJobLifetime(t *testing.T) {
	for _, parent := range []context.Context{nil, context.Background()} {
		fixture := &http.Client{Timeout: time.Minute}
		options := &sourceModels.SourceClientOptions{Context: parent, TestHttpClient: fixture}
		WithProviderRequestLifetime()(options)
		require.Same(t, fixture, options.TestHttpClient)
		require.Equal(t, parent, options.Context)
		require.False(t, options.TestMode)
	}
	parent, cancel := context.WithCancel(t.Context())
	defer cancel()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	base := providerPrivacyTransport(func(*http.Request) (*http.Response, error) { return nil, errors.New("unused") })
	redirect := func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	client := &http.Client{Transport: base, Timeout: 17 * time.Second, Jar: jar, CheckRedirect: redirect}
	options := &sourceModels.SourceClientOptions{Context: context.WithValue(parent, oauth2.HTTPClient, client), TestHttpClient: client}
	WithProviderRequestLifetime()(options)
	copied := options.Context.Value(oauth2.HTTPClient).(*http.Client)
	require.NotSame(t, client, copied)
	require.NotSame(t, client, options.TestHttpClient)
	require.Equal(t, client.Timeout, copied.Timeout)
	require.Same(t, jar, copied.Jar)
	require.ErrorIs(t, copied.CheckRedirect(nil, nil), http.ErrUseLastResponse)
	require.IsType(t, base, client.Transport, "caller transport is not overwritten")
	require.False(t, options.TestMode, "an explicit fixture client does not enable test mode")
	cancel()
	require.ErrorIs(t, options.Context.Err(), context.Canceled)
}

func TestProviderLifetimeCompletedScopesNeverStartTransport(t *testing.T) {
	for _, scope := range []string{"constructor", "request", "expired deadline"} {
		t.Run(scope, func(t *testing.T) {
			parent, cancelParent := context.WithCancelCause(t.Context())
			defer cancelParent(context.Canceled)
			requestContext, cancelRequest := context.WithCancel(t.Context())
			defer cancelRequest()
			var constructor context.Context = parent
			want := error(context.Canceled)
			if scope == "constructor" {
				cancelParent(errors.New("synthetic-private-constructor-cause"))
			} else if scope == "request" {
				cancelRequest()
			} else {
				expired, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
				defer cancel()
				constructor = expired
				want = context.DeadlineExceeded
			}
			calls := 0
			boundary := providerRequestLifetimeTransport{parent: constructor, base: providerPrivacyTransport(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("unexpected transport call")
			})}
			request, err := http.NewRequestWithContext(requestContext, http.MethodGet, "https://synthetic-provider.invalid", nil)
			require.NoError(t, err)
			response, err := boundary.RoundTrip(request)
			require.ErrorIs(t, err, want)
			require.NotContains(t, err.Error(), "synthetic-private-constructor-cause")
			require.Nil(t, response)
			require.Zero(t, calls)
		})
	}
}

func TestProviderLifetimeCauseIsStandardAndDeadlineSurvivesBody(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		parent, cancel := context.WithCancelCause(t.Context())
		var constructor context.Context = parent
		deadlineCancel := func() {}
		want := error(context.Canceled)
		if deadline {
			constructor, deadlineCancel = context.WithTimeout(parent, 40*time.Millisecond)
			want = context.DeadlineExceeded
		}
		var joined context.Context
		boundary := providerRequestLifetimeTransport{parent: constructor, base: providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
			joined = request.Context()
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("unchanged")), Request: request}, nil
		})}
		request, err := http.NewRequest(http.MethodGet, "https://synthetic-provider.invalid", nil)
		require.NoError(t, err)
		response, err := boundary.RoundTrip(request)
		require.NoError(t, err)
		if !deadline {
			cancel(errors.New("synthetic-private-constructor-cause"))
		}
		select {
		case <-joined.Done():
		case <-time.After(time.Second):
			t.Fatal("constructor lifetime did not remain linked through the body")
		}
		require.ErrorIs(t, context.Cause(joined), want)
		require.NotContains(t, context.Cause(joined).Error(), "synthetic-private-constructor-cause")
		require.NoError(t, response.Body.Close())
		deadlineCancel()
		cancel(context.Canceled)
	}
}
