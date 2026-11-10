package models

import (
	"context"
	"io"
	"net/http"
	"sync"

	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"golang.org/x/oauth2"
)

// WithProviderRequestLifetime links SDK-manufactured requests to the supplied
// constructor lifetime without discarding their own cancellation or deadline.
// Append after HTTP-client/privacy options. Detached Background jobs stay
// detached; this does not invent a timeout or copy an HTTP/Gin actor into jobs.
func WithProviderRequestLifetime() func(*sourceModels.SourceClientOptions) {
	return func(options *sourceModels.SourceClientOptions) {
		parent := options.Context
		if parent == nil {
			return
		}
		if _, deadline := parent.Deadline(); parent.Done() == nil && !deadline {
			return
		}
		client := http.DefaultClient
		if configured, ok := parent.Value(oauth2.HTTPClient).(*http.Client); ok && configured != nil {
			client = configured
		}
		options.Context = context.WithValue(parent, oauth2.HTTPClient, providerLifetimeHTTPClient(client, parent))
		if options.TestHttpClient != nil {
			options.TestHttpClient = providerLifetimeHTTPClient(options.TestHttpClient, parent)
		}
	}
}

func providerLifetimeHTTPClient(client *http.Client, parent context.Context) *http.Client {
	clone := *client
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = providerRequestLifetimeTransport{base: base, parent: parent}
	return &clone
}

type providerRequestLifetimeTransport struct {
	base   http.RoundTripper
	parent context.Context
}

func (transport providerRequestLifetimeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := transport.parent.Err(); err != nil {
		return nil, err
	}
	if err := request.Context().Err(); err != nil {
		return nil, err
	}
	joined, cancel := context.WithCancelCause(request.Context())
	var deadlineCancel context.CancelFunc
	if deadline, ok := transport.parent.Deadline(); ok {
		joined, deadlineCancel = context.WithDeadline(joined, deadline)
	}
	stopParent := context.AfterFunc(transport.parent, func() {
		// Err carries the standard cancellation class, not an arbitrary private
		// cancellation cause that a transport could put into diagnostics.
		cancel(transport.parent.Err())
	})
	var once sync.Once
	release := func() {
		once.Do(func() {
			stopParent()
			if deadlineCancel != nil {
				deadlineCancel()
			}
			cancel(context.Canceled)
		})
	}
	response, err := transport.base.RoundTrip(request.Clone(joined))
	if err != nil || response == nil || response.Body == nil || response.Body == http.NoBody {
		release()
		return response, err
	}
	// RoundTrip ends at response headers, not at the end of clinical bytes.
	// Keep both lifetimes linked until consumption finishes or the caller closes.
	clone := *response
	clone.Body = &providerRequestLifetimeBody{ReadCloser: response.Body, release: release}
	return &clone, nil
}

type providerRequestLifetimeBody struct {
	io.ReadCloser
	release func()
}

func (body *providerRequestLifetimeBody) Read(buffer []byte) (int, error) {
	n, err := body.ReadCloser.Read(buffer)
	if err == io.EOF {
		body.release()
	}
	return n, err
}

func (body *providerRequestLifetimeBody) Close() error {
	defer body.release()
	return body.ReadCloser.Close()
}
