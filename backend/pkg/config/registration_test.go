package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPublicRegistrationPolicyConfiguration(t *testing.T) {
	cfg, err := Create()
	require.NoError(t, err)
	enabled, err := PublicRegistrationEnabled(cfg)
	require.NoError(t, err)
	require.True(t, enabled)
	for _, value := range []interface{}{false, "false", "0"} {
		cfg.Set("auth.registration.enabled", value)
		enabled, err = PublicRegistrationEnabled(cfg)
		require.NoError(t, err)
		require.False(t, enabled)
	}
	cfg.Set("auth.registration.enabled", "invalid")
	_, err = PublicRegistrationEnabled(cfg)
	require.Error(t, err)
	require.Error(t, cfg.(*configuration).ValidateConfig())
}

func TestPublicRegistrationEnvironmentOverride(t *testing.T) {
	t.Setenv("FASTEN_AUTH_REGISTRATION_ENABLED", "false")
	cfg, err := Create()
	require.NoError(t, err)
	enabled, err := PublicRegistrationEnabled(cfg)
	require.NoError(t, err)
	require.False(t, enabled)
}
