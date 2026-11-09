package host

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLocalIdentityResolverIsFixedAndBounded(t *testing.T) {
	owner := LocalIdentity{ID: uuid.New(), Username: "synthetic-owner"}
	calls := 0
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/secure/account/me", r.URL.Path)
		require.Empty(t, r.URL.RawQuery)
		require.Equal(t, "Bearer synthetic-local-token", r.Header.Get("Authorization"))
		deadline, ok := r.Context().Deadline()
		require.True(t, ok)
		require.LessOrEqual(t, time.Until(deadline), 10*time.Second)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": owner}))
	})
	resolve := NewLocalIdentityResolver(backend)
	actual, err := resolve(context.Background(), "synthetic-local-token")
	require.NoError(t, err)
	require.Equal(t, owner, actual)
	for _, token := range []string{"", "synthetic\r\nprivate-token", strings.Repeat("x", 16385)} {
		_, err := resolve(context.Background(), token)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "private-token")
	}
	require.Equal(t, 1, calls)
	_, err = NewLocalIdentityResolver(nil)(context.Background(), "synthetic-local-token")
	require.Error(t, err)
}

func TestLocalIdentityResolverNeverReturnsPrivateErrors(t *testing.T) {
	owner := uuid.New().String()
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"rejected", 401, `synthetic-private-rejection`},
		{"closed", 503, `synthetic-private-closed`},
		{"bad JSON", 200, `synthetic-private-not-json`},
		{"false", 200, fmt.Sprintf(`{"success":false,"data":{"id":%q,"username":"synthetic-private-name"}}`, owner)},
		{"nil ID", 200, `{"success":true,"data":{"id":"00000000-0000-0000-0000-000000000000","username":"synthetic-private-name"}}`},
		{"large", 200, strings.Repeat("synthetic-private-response", 4096)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolve := NewLocalIdentityResolver(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			actual, err := resolve(context.Background(), "synthetic-token")
			require.Error(t, err)
			require.Empty(t, actual.Username)
			require.NotContains(t, err.Error(), "synthetic-private-")
		})
	}
}
