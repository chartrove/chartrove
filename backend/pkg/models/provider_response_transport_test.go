package models

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type providerFailureTrackedBody struct {
	io.Reader
	bytesRead int
	closed    bool
}

func (body *providerFailureTrackedBody) Read(buffer []byte) (int, error) {
	n, err := body.Reader.Read(buffer)
	body.bytesRead += n
	return n, err
}
func (body *providerFailureTrackedBody) Close() error { body.closed = true; return nil }

func TestProviderFailureCodeIsOnlyAnUnambiguousAllowlistedName(t *testing.T) {
	for _, code := range []string{"invalid_request", "invalid_client", "invalid_grant", "unauthorized_client", "unsupported_grant_type", "invalid_scope", "server_error", "temporarily_unavailable"} {
		require.Equal(t, code, providerSafeOAuthCode([]byte(`{"error":"`+code+`","error_description":"synthetic-private-description"}`)))
	}
	for _, payload := range []string{
		``, `null`, `[]`, `"invalid_grant"`, `{`, `{"error":null}`, `{"error":123}`,
		`{"error":"private-canary"}`, `{"Error":"invalid_grant"}`,
		`{"error":"invalid_grant","error":"invalid_grant"}`,
		`{"error":"invalid_grant","ERROR":"invalid_client"}`,
		`{"error":"invalid_grant"} {}`, `{"error":"invalid_grant",}`,
		`{"error_description":"private-canary"}`, `{"error":"invalid_grant private-canary"}`,
	} {
		require.Equal(t, providerFailureCode, providerSafeOAuthCode([]byte(payload)), "ambiguous/untrusted reply must not become a classification")
	}
}

func TestProviderFailureFormClassification(t *testing.T) {
	for _, contentType := range []string{"application/x-www-form-urlencoded; charset=UTF-8", "text/plain"} {
		require.Equal(t, "invalid_grant", providerSafeOAuthResponseCode([]byte("error=invalid_grant&error_description=synthetic-private-description"), contentType))
		for _, payload := range []string{
			"error=invalid_grant&error=invalid_grant", "error=invalid_grant&ERROR=invalid_client",
			"error=private-canary", "error=invalid_grant%zz", "error_description=private-canary",
		} {
			require.Equal(t, providerFailureCode, providerSafeOAuthResponseCode([]byte(payload), contentType))
		}
	}
	require.Equal(t, providerFailureCode, providerSafeOAuthResponseCode([]byte("error=invalid_grant"), "application/json"))
}

func TestProviderFailureTransportPreservesSuccessAndBoundsRejection(t *testing.T) {
	const original = `{"resourceType":"Observation","valueQuantity":{"value":1.234567890123456789},"private":"synthetic-body-canary"}`
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://synthetic.invalid/fhir", nil)
	require.NoError(t, err)
	for _, status := range []int{200, 201, 204, 206, 299, 300, 400, 401, 500, 599} {
		body := &providerFailureTrackedBody{Reader: strings.NewReader(original)}
		response := &http.Response{StatusCode: status, Status: "synthetic-private-reason", Body: body, Header: http.Header{"Content-Type": {"text/plain"}, "Location": {"/next"}}, ContentLength: int64(len(original))}
		transport := providerResponseTransport{base: providerPrivacyTransport(func(got *http.Request) (*http.Response, error) {
			require.Same(t, request, got)
			return response, nil
		})}
		got, err := transport.RoundTrip(request)
		require.NoError(t, err)
		if status >= 200 && status < 300 {
			require.Same(t, response, got)
			require.Zero(t, body.bytesRead)
			require.False(t, body.closed)
			payload, err := io.ReadAll(got.Body)
			require.NoError(t, err)
			require.Equal(t, original, string(payload), "never normalize successful clinical decimal lexemes")
			require.NoError(t, got.Body.Close())
			continue
		}
		require.NotSame(t, response, got)
		require.True(t, body.closed)
		require.Equal(t, "synthetic-private-reason", response.Status, "do not mutate the transport's response object")
		require.Equal(t, "text/plain", response.Header.Get("Content-Type"))
		require.Equal(t, "/next", got.Header.Get("Location"), "do not silently alter redirect policy")
		payload, err := io.ReadAll(got.Body)
		require.NoError(t, err)
		require.Equal(t, `{"error":"provider_request_failed"}`, string(payload))
		require.NotContains(t, got.Status, "synthetic-private-reason")
		require.EqualValues(t, len(payload), got.ContentLength)
		require.NoError(t, got.Body.Close())
	}
	large := &providerFailureTrackedBody{Reader: strings.NewReader(strings.Repeat("x", maxProviderFailureBytes*4))}
	transport := providerResponseTransport{base: providerPrivacyTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: large}, nil
	})}
	got, err := transport.RoundTrip(request)
	require.NoError(t, err)
	require.Equal(t, maxProviderFailureBytes+1, large.bytesRead)
	require.True(t, large.closed)
	require.NoError(t, got.Body.Close())

	failure := errors.New("synthetic transport failure")
	transport.base = providerPrivacyTransport(func(*http.Request) (*http.Response, error) { return nil, failure })
	got, err = transport.RoundTrip(request)
	require.Nil(t, got)
	require.Same(t, failure, err, "transport errors are not silently reclassified as HTTP replies")
}

func TestProviderPrivacyOptionPreservesClientAndCallerLifetime(t *testing.T) {
	type ownerKey struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(t.Context(), ownerKey{}, "synthetic-immutable-owner"), time.Minute)
	defer cancel()
	base := providerPrivacyTransport(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, "synthetic-immutable-owner", request.Context().Value(ownerKey{}))
		return nil, request.Context().Err()
	})
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	redirectFailure := errors.New("synthetic redirect policy")
	client := &http.Client{Transport: base, Timeout: 23 * time.Second, Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return redirectFailure }}
	options := &sourceModels.SourceClientOptions{Context: context.WithValue(ctx, oauth2.HTTPClient, client), TestHttpClient: client, TestMode: true}
	WithProviderResponsePrivacy()(options)
	private := options.Context.Value(oauth2.HTTPClient).(*http.Client)
	require.NotSame(t, client, private)
	require.Equal(t, client.Timeout, private.Timeout)
	require.Same(t, jar, private.Jar)
	require.Same(t, redirectFailure, private.CheckRedirect(nil, nil))
	require.IsType(t, providerPrivacyTransport(nil), client.Transport, "never replace shared/default transports globally")
	require.NotSame(t, client, options.TestHttpClient)
	require.Equal(t, 23*time.Second, client.Timeout, "SDK test-mode timeout mutation stays on the copy")
	require.True(t, options.TestMode)
	deadline, ok := options.Context.Deadline()
	require.True(t, ok)
	parentDeadline, _ := ctx.Deadline()
	require.Equal(t, parentDeadline, deadline)
	cancel()
	require.ErrorIs(t, options.Context.Err(), context.Canceled)
	request, err := http.NewRequestWithContext(options.Context, http.MethodGet, "https://synthetic.invalid", nil)
	require.NoError(t, err)
	_, err = private.Transport.RoundTrip(request)
	require.ErrorIs(t, err, context.Canceled)
}
