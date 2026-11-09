package web

import (
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/stretchr/testify/require"
)

// Exercise exactly the address construction used by startServer without calling
// it: the baseline must not expose a wildcard TCP listener or initialize storage.
func TestStandaloneListenerAddressConstruction(t *testing.T) {
	for _, tc := range []struct {
		name string
		host string
		set  bool
		want string
	}{
		{name: "default", want: "127.0.0.1:8080"},
		{name: "empty", host: "", set: true, want: "127.0.0.1:8080"},
		{name: "blank", host: " \t", set: true, want: "127.0.0.1:8080"},
		{name: "v4 loopback", host: "127.0.0.2", set: true, want: "127.0.0.2:8080"},
		{name: "v4 explicit wildcard", host: "0.0.0.0", set: true, want: "0.0.0.0:8080"},
		{name: "explicit external", host: "198.51.100.42", set: true, want: "198.51.100.42:8080"},
		{name: "hostname", host: "chartrove.invalid", set: true, want: "chartrove.invalid:8080"},
		{name: "v6 loopback", host: "::1", set: true, want: "[::1]:8080"},
		{name: "v6 wildcard", host: "::", set: true, want: "[::]:8080"},
		{name: "existing bracketed v6", host: "[::1]", set: true, want: "[::1]:8080"},
		{name: "existing bracketed wildcard", host: "[::]", set: true, want: "[::]:8080"},
		{name: "empty brackets cannot become wildcard", host: "[]", set: true, want: "127.0.0.1:8080"},
		{name: "v6 zone", host: "fe80::1%synthetic", set: true, want: "[fe80::1%synthetic]:8080"},
		{name: "existing bracketed zone", host: "[fe80::1%synthetic]", set: true, want: "[fe80::1%synthetic]:8080"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("FASTEN_WEB_LISTEN_HOST", "")
			cfg, err := config.Create()
			require.NoError(t, err)
			if tc.set {
				cfg.Set("web.listen.host", tc.host)
			}
			ae := &AppEngine{Config: cfg}
			require.Equal(t, tc.want, ae.listenAddress())
		})
	}
}
