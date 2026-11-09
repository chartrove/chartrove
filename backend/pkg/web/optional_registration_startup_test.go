package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type optionalRegistrationUnavailableTransport struct{ directory string }

func (transport optionalRegistrationUnavailableTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodPost || request.URL.String() != "https://api.platform.fastenhealth.com/v1/installation/register" {
		return nil, errors.New("unexpected request in isolated startup fixture")
	}
	if err := os.WriteFile(filepath.Join(transport.directory, "registration-attempt"), []byte("attempted\n"), 0o600); err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusServiceUnavailable, Status: "503 synthetic optional service unavailable", Header: make(http.Header), Body: io.NopCloser(strings.NewReader("synthetic-unavailable-body")), Request: request}, nil
}

func TestStandaloneStartupSurvivesOptionalRegistrationOutage(t *testing.T) {
	if os.Getenv("CHARTROVE_OPTIONAL_STARTUP_CHILD") == "1" {
		directory := os.Getenv("CHARTROVE_OPTIONAL_STARTUP_DIRECTORY")
		require.NotEmpty(t, directory)
		cfg, err := config.Create()
		require.NoError(t, err)
		cfg.Set("jwt.issuer.key", "synthetic-private-signing-key-for-isolated-startup-only")
		cfg.Set("database.location", filepath.Join(directory, "synthetic.db"))
		cfg.Set("database.encryption.enabled", true)
		cfg.Set("database.encryption.key", "synthetic-encryption-key-for-isolated-startup-only")
		cfg.Set("cache.location", filepath.Join(directory, "cache"))
		cfg.Set("web.src.frontend.path", directory)
		cfg.Set("web.listen.host", "127.0.0.1")
		cfg.Set("web.listen.port", os.Getenv("CHARTROVE_OPTIONAL_STARTUP_PORT"))
		cfg.Set("web.listen.https.enabled", false)
		// This mutation exists only in the owned test subprocess. No remote dial.
		http.DefaultClient = &http.Client{Transport: optionalRegistrationUnavailableTransport{directory: directory}}
		logger := logrus.NewEntry(logrus.New())
		engine := AppEngine{Config: cfg, Logger: logger, EventBus: event_bus.NewNoopEventBusServer()}
		err = engine.Start()
		t.Fatalf("isolated startup unexpectedly returned: %v", err)
		return
	}

	directory, err := os.MkdirTemp(os.TempDir(), "chartrove-optional-startup-")
	require.NoError(t, err)
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Failed synthetic fixture retained privately: %s", directory)
			return
		}
		require.NoError(t, os.RemoveAll(directory))
	})
	reservation, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	port := reservation.Addr().(*net.TCPAddr).Port
	require.NoError(t, reservation.Close())
	logfile, err := os.OpenFile(filepath.Join(directory, "child.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStandaloneStartupSurvivesOptionalRegistrationOutage$", "-test.v")
	command.Dir = directory
	command.Env = []string{
		"PATH=" + os.Getenv("PATH"), "HOME=" + directory, "TMPDIR=" + directory,
		"CHARTROVE_OPTIONAL_STARTUP_CHILD=1", "CHARTROVE_OPTIONAL_STARTUP_DIRECTORY=" + directory,
		fmt.Sprintf("CHARTROVE_OPTIONAL_STARTUP_PORT=%d", port),
	}
	command.Stdout, command.Stderr = logfile, logfile
	if err := command.Start(); err != nil {
		cancel()
		_ = logfile.Close()
		t.Fatal(err)
	}
	finished := make(chan struct{})
	var childError error
	go func() { childError = command.Wait(); close(finished) }()
	t.Cleanup(func() {
		cancel() // Only the child created by this CommandContext.
		<-finished
		require.NoError(t, logfile.Close())
	})
	client := &http.Client{Timeout: 250 * time.Millisecond, Transport: &http.Transport{Proxy: nil}}
	defer client.CloseIdleConnections()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-finished:
			t.Fatalf("isolated engine exited before local health readiness: %v (private child log retained)", childError)
		default:
		}
		response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/health", port))
		if err == nil {
			payload, readErr := io.ReadAll(io.LimitReader(response.Body, 65537))
			_ = response.Body.Close()
			require.NoError(t, readErr)
			require.Equal(t, http.StatusOK, response.StatusCode)
			require.JSONEq(t, `{"success":true,"data":{"first_run_wizard":true,"standby_mode":false}}`, string(payload))
			_, err := os.Stat(filepath.Join(directory, "registration-attempt"))
			require.NoError(t, err, "exercise the failed optional registration, not a cached or skipped path")
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("local health readiness deadline exceeded; private child log retained")
}
