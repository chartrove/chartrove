package host

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// LocalIdentity is backend-authenticated application identity, not an identity
// asserted by a provider Patient, the renderer or a decoded unsigned claim.
type LocalIdentity struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type IdentityResolver func(context.Context, string) (LocalIdentity, error)

var errLocalIdentity = errors.New("local identity could not be verified")

// NewLocalIdentityResolver uses the same embedded JWT/session/account middleware
// as the API. This is an in-process handler invocation: no socket, credentials in
// a URL, second auth implementation or browser-supplied backend address.
func NewLocalIdentityResolver(backend http.Handler) IdentityResolver {
	return func(parent context.Context, token string) (LocalIdentity, error) {
		if backend == nil || len(token) == 0 || len(token) > 16384 || strings.IndexFunc(token, unicode.IsSpace) >= 0 {
			return LocalIdentity{}, errLocalIdentity
		}
		ctx, cancel := context.WithTimeout(parent, 10*time.Second)
		defer cancel()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://chartrove.invalid/api/secure/account/me", nil)
		if err != nil {
			return LocalIdentity{}, errLocalIdentity
		}
		request.Header.Set("Authorization", "Bearer "+token)
		response := &identityResponse{headers: make(http.Header)}
		backend.ServeHTTP(response, request)
		if ctx.Err() != nil || response.status != http.StatusOK || response.overflow {
			return LocalIdentity{}, errLocalIdentity
		}
		var envelope struct {
			Success bool          `json:"success"`
			Data    LocalIdentity `json:"data"`
		}
		if json.Unmarshal(response.body.Bytes(), &envelope) != nil || !envelope.Success || envelope.Data.ID == uuid.Nil || envelope.Data.Username == "" || len(envelope.Data.Username) > 256 || !utf8.ValidString(envelope.Data.Username) || strings.IndexFunc(envelope.Data.Username, unicode.IsControl) >= 0 {
			return LocalIdentity{}, errLocalIdentity
		}
		return envelope.Data, nil
	}
}

type identityResponse struct {
	headers  http.Header
	status   int
	body     bytes.Buffer
	overflow bool
}

func (w *identityResponse) Header() http.Header { return w.headers }
func (w *identityResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *identityResponse) Write(raw []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	const limit = 64 << 10
	if len(raw) > limit-w.body.Len() {
		w.overflow = true
		return len(raw), nil
	}
	_, _ = w.body.Write(raw)
	return len(raw), nil
}
