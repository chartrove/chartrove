package models

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/jwk"
	"github.com/fastenhealth/fasten-sources/catalog"
	definitions "github.com/fastenhealth/fasten-sources/definitions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type providerPrivacyTransport func(*http.Request) (*http.Response, error)

func (f providerPrivacyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type providerPrivacyBody struct {
	io.Reader
	reads  int
	closed bool
}

func (b *providerPrivacyBody) Read(p []byte) (int, error) {
	b.reads++
	return b.Reader.Read(p)
}
func (b *providerPrivacyBody) Close() error { b.closed = true; return nil }

// Use only the pinned module's embedded catalog metadata to exercise the real
// discovery path. The HTTP client below intercepts every request: no socket,
// provider account or real credential is used.
func providerPrivacyEndpoint(t *testing.T) uuid.UUID {
	t.Helper()
	endpoints, err := catalog.GetEndpoints(nil)
	require.NoError(t, err)
	ids := make([]string, 0, len(endpoints))
	for id, endpoint := range endpoints {
		if string(endpoint.GetPlatformType()) == "epic" {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	require.NotEmpty(t, ids, "embedded provider regression fixture disappeared")
	definition, err := definitions.GetSourceDefinition(definitions.GetSourceConfigOptions{EndpointId: ids[0]})
	require.NoError(t, err)
	require.NotEmpty(t, definition.DynamicClientRegistrationMode)
	require.NotEmpty(t, definition.RegistrationEndpoint)
	require.NotEmpty(t, definition.TokenEndpoint)
	id, err := uuid.Parse(ids[0])
	require.NoError(t, err)
	return id
}

func TestDynamicClientResponseBodyPrivacy(t *testing.T) {
	endpointID := providerPrivacyEndpoint(t)
	key, err := jwk.JWKGenerate()
	require.NoError(t, err)
	serialized, err := jwk.JWKSerialize(key)
	require.NoError(t, err)
	for _, operation := range []string{"registration", "token"} {
		for _, status := range []int{http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError} {
			t.Run(fmt.Sprintf("%s/%d", operation, status), func(t *testing.T) {
				const marker = "synthetic-private-provider-reply"
				payload := `{"error":"` + marker + `"}`
				if status == http.StatusOK {
					if operation == "registration" {
						payload = `{"client_id":"synthetic-registered-client"}`
					} else {
						payload = `{"access_token":"synthetic-refreshed-access","expires_in":3600}`
					}
				}
				body := &providerPrivacyBody{Reader: strings.NewReader(payload)}
				calls := 0
				originalClient := http.DefaultClient
				http.DefaultClient = &http.Client{Transport: providerPrivacyTransport(func(r *http.Request) (*http.Response, error) {
					calls++
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "https", r.URL.Scheme)
					if operation == "registration" {
						require.Equal(t, "Bearer synthetic-original-access", r.Header.Get("Authorization"))
					} else {
						require.NoError(t, r.ParseForm())
						require.NotEmpty(t, r.PostForm.Get("assertion"), "exercise actual synthetic signing")
					}
					return &http.Response{StatusCode: status, Body: body, Header: make(http.Header), Request: r}, nil
				})}
				t.Cleanup(func() { http.DefaultClient = originalClient })
				var output bytes.Buffer
				originalWriter := log.Writer()
				log.SetOutput(&output)
				t.Cleanup(func() { log.SetOutput(originalWriter) })
				credential := SourceCredential{
					EndpointID: endpointID, ClientId: "synthetic-software", DynamicClientId: "synthetic-existing-client",
					AccessToken: "synthetic-original-access", ExpiresAt: 123,
					DynamicClientJWKS: []map[string]string{serialized},
				}
				var result error
				if operation == "registration" {
					result = credential.RegisterDynamicClient()
				} else {
					result = credential.RefreshDynamicClientAccessToken()
				}
				require.Equal(t, 1, calls)
				require.True(t, body.closed, "release the response even on a rejected status")
				require.NotContains(t, output.String(), marker)
				if status != http.StatusOK {
					require.ErrorContains(t, result, fmt.Sprint(status), "retain non-payload status diagnostics")
					require.NotContains(t, result.Error(), marker)
					require.Zero(t, body.reads, "do not read an untrusted, potentially unbounded error body")
					require.Equal(t, "synthetic-existing-client", credential.DynamicClientId)
					require.Equal(t, "synthetic-original-access", credential.AccessToken)
					require.EqualValues(t, 123, credential.ExpiresAt)
					return
				}
				require.NoError(t, result)
				require.Positive(t, body.reads, "successful responses must still be decoded")
				if operation == "registration" {
					require.Equal(t, "synthetic-registered-client", credential.DynamicClientId)
					require.NotEmpty(t, credential.DynamicClientJWKS)
				} else {
					require.Equal(t, "synthetic-refreshed-access", credential.AccessToken)
					require.Greater(t, credential.ExpiresAt, time.Now().Unix())
				}
			})
		}
	}
}
