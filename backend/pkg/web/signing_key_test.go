package web

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// An unsupported database type ensures these tests cannot open any database.
func TestStartRejectsUnsafeSigningKeyBeforeDatabaseSetup(t *testing.T) {
	for _, tc := range []struct{ name, key string }{
		{"missing", ""},
		{"whitespace", "  "},
		{"historical default", "thisismysupersecuressessionsecretlength"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("jwt.issuer.key", tc.key)
			cfg.Set("database.type", "synthetic-no-database")
			var output bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&output)
			engine := AppEngine{Config: cfg, Logger: logrus.NewEntry(logger)}
			err = engine.Start()
			require.ErrorContains(t, err, "jwt.issuer.key")
			require.Empty(t, output.String(), "unsafe signing keys must be rejected before database initialization")
		})
	}
}

func TestStartUsesConfiguredEnvironmentSigningKey(t *testing.T) {
	t.Setenv("FASTEN_JWT_ISSUER_KEY", strings.Repeat("synthetic", 5))
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "synthetic-no-database")
	var output bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&output)
	engine := AppEngine{Config: cfg, Logger: logrus.NewEntry(logger)}
	err = engine.Start()
	require.ErrorContains(t, err, "failed to initialize database repository")
	require.NotContains(t, err.Error(), "jwt.issuer.key")
	require.NotContains(t, output.String(), strings.Repeat("synthetic", 5))
}
