package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"golang.org/x/oauth2"
)

const maxProviderFailureBytes = 4096
const providerFailureCode = "provider_request_failed"

// WithProviderResponsePrivacy bounds HTTP non-success bodies before the pinned
// SDK can include them in errors or standard-log diagnostics. Append it after
// other HTTP-client options. Successful bodies and caller context remain intact.
// This is not a filter for request URLs, response headers or 2xx OAuth errors.
func WithProviderResponsePrivacy() func(*sourceModels.SourceClientOptions) {
	return func(options *sourceModels.SourceClientOptions) {
		if options.Context == nil {
			// Do not replace an absent caller lifetime/actor with Background.
			return
		}
		client := http.DefaultClient
		if configured, ok := options.Context.Value(oauth2.HTTPClient).(*http.Client); ok && configured != nil {
			client = configured
		}
		privateClient := providerPrivateHTTPClient(client)
		options.Context = context.WithValue(options.Context, oauth2.HTTPClient, privateClient)
		if options.TestHttpClient != nil {
			// Preserve explicit test mode without mutating the shared fixture client.
			options.TestHttpClient = providerPrivateHTTPClient(options.TestHttpClient)
		}
	}
}

func providerPrivateHTTPClient(client *http.Client) *http.Client {
	clone := *client
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = providerResponseTransport{base: base}
	return &clone
}

type providerResponseTransport struct{ base http.RoundTripper }

func (transport providerResponseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.base.RoundTrip(request)
	if err != nil || response == nil || (response.StatusCode >= 200 && response.StatusCode < 300) {
		return response, err
	}

	code := providerFailureCode
	if response.Body != nil {
		payload, readErr := io.ReadAll(io.LimitReader(response.Body, maxProviderFailureBytes+1))
		_ = response.Body.Close()
		if readErr == nil && len(payload) <= maxProviderFailureBytes {
			code = providerSafeOAuthResponseCode(payload, response.Header.Get("Content-Type"))
		}
	}
	payload := []byte(fmt.Sprintf(`{"error":%q}`, code))
	clone := *response
	clone.Body = io.NopCloser(bytes.NewReader(payload))
	clone.ContentLength = int64(len(payload))
	clone.Header = response.Header.Clone()
	if clone.Header == nil {
		clone.Header = make(http.Header)
	}
	clone.Header.Set("Content-Type", "application/json")
	for _, name := range []string{"Content-Length", "Content-Encoding", "Transfer-Encoding", "Trailer"} {
		clone.Header.Del(name)
	}
	clone.TransferEncoding = nil
	clone.Trailer = nil
	clone.Uncompressed = false
	clone.Status = strings.TrimSpace(fmt.Sprintf("%d %s", clone.StatusCode, http.StatusText(clone.StatusCode)))
	return &clone, nil
}

func providerSafeOAuthResponseCode(payload []byte, contentType string) string {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	if mediaType != "application/x-www-form-urlencoded" && mediaType != "text/plain" {
		return providerSafeOAuthCode(payload)
	}
	// Match the pinned OAuth client's supported form response formats too.
	values, err := url.ParseQuery(string(payload))
	if err != nil || len(values["error"]) != 1 {
		return providerFailureCode
	}
	for name := range values {
		if strings.EqualFold(name, "error") && name != "error" {
			return providerFailureCode
		}
	}
	return providerAllowedOAuthCode(values["error"][0])
}

// Only an unambiguous exact OAuth error name is retained. Descriptions, unknown
// values, aliases, duplicate error fields and trailing JSON are never echoed.
func providerSafeOAuthCode(payload []byte) string {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return providerFailureCode
	}
	found := false
	var code string
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return providerFailureCode
		}
		name, ok := key.(string)
		if !ok || (strings.EqualFold(name, "error") && name != "error") {
			return providerFailureCode
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return providerFailureCode
		}
		if name == "error" {
			if found || json.Unmarshal(value, &code) != nil {
				return providerFailureCode
			}
			found = true
		}
	}
	if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
		return providerFailureCode
	}
	var trailing json.RawMessage
	if decoder.Decode(&trailing) != io.EOF {
		return providerFailureCode
	}
	return providerAllowedOAuthCode(code)
}

func providerAllowedOAuthCode(code string) string {
	switch code {
	case "invalid_request", "invalid_client", "invalid_grant", "unauthorized_client", "unsupported_grant_type", "invalid_scope", "server_error", "temporarily_unavailable":
		return code
	default:
		return providerFailureCode
	}
}
