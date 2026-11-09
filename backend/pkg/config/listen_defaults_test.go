package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// These are configuration/source contracts, not OS listener or container-network
// acceptance. In particular, the baseline never opens a wildcard TCP socket.
func TestStandaloneListenerDefaultIsLoopback(t *testing.T) {
	t.Setenv("FASTEN_WEB_LISTEN_HOST", "")
	cfg, err := Create()
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", cfg.GetString("web.listen.host"))
	require.Equal(t, "8080", cfg.GetString("web.listen.port"))
	require.False(t, cfg.GetBool("web.allow_unsafe_endpoints"))
}

func TestRepositoryStandaloneSampleIsLoopback(t *testing.T) {
	t.Setenv("FASTEN_WEB_LISTEN_HOST", "")
	cfg, err := Create()
	require.NoError(t, err)
	require.NoError(t, cfg.ReadConfig("../../../config.yaml"))
	require.Equal(t, "127.0.0.1", cfg.GetString("web.listen.host"))
	// Do not weaken the sample's existing transport/storage choices to make
	// binding tests pass, or supply a shared signing/encryption key.
	require.True(t, cfg.GetBool("web.listen.https.enabled"))
	require.True(t, cfg.GetBool("database.encryption.enabled"))
	require.Empty(t, cfg.GetString("jwt.issuer.key"))
	require.Empty(t, cfg.GetString("database.encryption.key"))
}

func TestExplicitListenerChoicesRemainAvailable(t *testing.T) {
	for _, host := range []string{"0.0.0.0", "198.51.100.42", "localhost", "::1", "::"} {
		t.Run(host, func(t *testing.T) {
			t.Setenv("FASTEN_WEB_LISTEN_HOST", host)
			cfg, err := Create()
			require.NoError(t, err)
			require.NoError(t, cfg.ReadConfig("../../../config.yaml"))
			require.Equal(t, host, cfg.GetString("web.listen.host"), "explicit environment overrides the sample")
			cfg.Set("web.listen.host", "127.0.0.2")
			require.Equal(t, "127.0.0.2", cfg.GetString("web.listen.host"), "existing programmatic override precedence remains")
		})
	}
}

func TestDistributionKeepsExplicitContainerSideBind(t *testing.T) {
	t.Setenv("FASTEN_WEB_LISTEN_HOST", "")
	dockerfile, err := os.ReadFile("../../../Dockerfile")
	require.NoError(t, err)
	require.Contains(t, string(dockerfile), "sh scripts/container-config.sh config.yaml /opt/fasten/config/config.yaml")
	require.Contains(t, string(dockerfile), "COPY --from=backend-build  /opt/fasten/ /opt/fasten/")
	require.NotContains(t, string(dockerfile), "COPY config.yaml /opt/fasten/config/config.yaml")
	require.NotContains(t, string(dockerfile), "ENV FASTEN_WEB_LISTEN_HOST", "image ENV would override an operator-mounted configuration")

	source, err := os.ReadFile("../../../config.yaml")
	require.NoError(t, err)
	out := filepath.Join(t.TempDir(), "container.yaml")
	log, err := exec.Command("sh", "../../../scripts/container-config.sh", "../../../config.yaml", out).CombinedOutput()
	require.NoError(t, err, "%s", log)
	generated, err := os.ReadFile(out)
	require.NoError(t, err)
	require.Equal(t, strings.Replace(string(source), "    host: 127.0.0.1", "    host: 0.0.0.0", 1), string(generated), "only the deliberate container-side host changes")
	unchanged, err := os.ReadFile("../../../config.yaml")
	require.NoError(t, err)
	require.Equal(t, source, unchanged)
	cfg, err := Create()
	require.NoError(t, err)
	require.NoError(t, cfg.ReadConfig(out))
	require.Equal(t, "0.0.0.0", cfg.GetString("web.listen.host"))
	require.True(t, cfg.GetBool("database.encryption.enabled"))
	require.True(t, cfg.GetBool("web.listen.https.enabled"))

	_, err = exec.Command("sh", "../../../scripts/container-config.sh", "../../../config.yaml", out).CombinedOutput()
	require.Error(t, err, "never replace existing configuration")
	preserved, err := os.ReadFile(out)
	require.NoError(t, err)
	require.Equal(t, generated, preserved)
}

func TestComposeHostPublicationDefaultIsLoopback(t *testing.T) {
	for _, path := range []string{"../../../docker-compose.yml", "../../../docker-compose-prod.yml"} {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Contains(t, string(contents), `"${FASTEN_BIND_HOST:-127.0.0.1}:${PORT:-9090}:8080"`, "host publication stays separate from the deliberate container-side wildcard")
	}
}

func TestContainerConfigProducerFailsClosed(t *testing.T) {
	for _, input := range []string{"web:\n  listen:\n    host: 0.0.0.0\n", "    host: 127.0.0.1\n    host: 127.0.0.1\n"} {
		directory := t.TempDir()
		source := filepath.Join(directory, "source.yaml")
		destination := filepath.Join(directory, "new.yaml")
		require.NoError(t, os.WriteFile(source, []byte(input), 0600))
		_, err := exec.Command("sh", "../../../scripts/container-config.sh", source, destination).CombinedOutput()
		require.Error(t, err)
		_, err = os.Stat(destination)
		require.True(t, os.IsNotExist(err), "ambiguous input must not create output")
		preserved, err := os.ReadFile(source)
		require.NoError(t, err)
		require.Equal(t, input, string(preserved))
	}
	log, err := exec.Command("sh", "../../../scripts/container-config.sh", "--help").CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(log), "NEW_DESTINATION")
	_, err = exec.Command("sh", "../../../scripts/container-config.sh").CombinedOutput()
	require.Error(t, err)
}
