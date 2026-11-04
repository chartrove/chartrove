package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUserSessionDurationConfiguration(t *testing.T) {
	t.Setenv("FASTEN_JWT_SESSION_DURATION", "")
	for _, tc := range []struct {
		name, value string
		want        time.Duration
		invalid     bool
	}{
		{"missing", "", time.Hour, false},
		{"shorter", "5m", 5 * time.Minute, false},
		{"longer", "2h", 2 * time.Hour, false},
		{"whitespace", " 5m ", 5 * time.Minute, false},
		{"minimum", "1s", time.Second, false},
		{"zero", "0s", 0, true},
		{"negative", "-1h", 0, true},
		{"subsecond", "500ms", 0, true},
		{"unit missing", "3600", 0, true},
		{"malformed", "invalid-duration", 0, true},
		{"overflow", "999999999999999999h", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Create()
			require.NoError(t, err)
			if tc.value != "" {
				cfg.Set("jwt.session.duration", tc.value)
			}
			got, err := UserSessionDuration(cfg)
			if tc.invalid {
				require.ErrorContains(t, err, "jwt.session.duration")
				require.Error(t, cfg.(*configuration).ValidateConfig())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.NoError(t, cfg.(*configuration).ValidateConfig())
		})
	}
}

func TestUserSessionDurationEnvironmentAndFile(t *testing.T) {
	cfg, err := Create()
	require.NoError(t, err)
	file := filepath.Join(t.TempDir(), "synthetic-session.yaml")
	require.NoError(t, os.WriteFile(file, []byte("jwt:\n  session:\n    duration: 2h\n"), 0600))
	t.Setenv("FASTEN_JWT_SESSION_DURATION", "")
	require.NoError(t, cfg.ReadConfig(file))
	lifetime, err := UserSessionDuration(cfg)
	require.NoError(t, err)
	require.Equal(t, 2*time.Hour, lifetime)
	t.Setenv("FASTEN_JWT_SESSION_DURATION", "5m")
	lifetime, err = UserSessionDuration(cfg)
	require.NoError(t, err)
	require.Equal(t, 5*time.Minute, lifetime)
	t.Setenv("FASTEN_JWT_SESSION_DURATION", "0s")
	require.Error(t, cfg.(*configuration).ValidateConfig())
}
