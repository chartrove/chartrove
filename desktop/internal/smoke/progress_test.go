package smoke

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeProgressKeepsOnlyFixedOrderedLabels(t *testing.T) {
	var session Session
	handler := session.Wrap(http.NotFoundHandler())
	post := func(body string) int {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/__desktop_smoke__/control?action=progress", strings.NewReader(body)))
		return response.Code
	}
	body, err := json.Marshal(map[string]any{"checks": nativeSmokeChecks[:6]})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, post(string(body)))
	require.Equal(t, http.StatusNoContent, post(`{"checks":["native binding round trip"]}`))
	for _, bad := range []string{
		`{"checks":["PRIVATE_RECORD_CANARY"]}`,
		`{"checks":["local sign in"]}`,
		`{"checks":[],"private":"PRIVATE_RECORD_CANARY"}`,
		`{"checks":[]} {"private":"PRIVATE_RECORD_CANARY"}`,
		`{"checks":["` + strings.Repeat("x", 2049) + `"]}`,
	} {
		require.Equal(t, http.StatusBadRequest, post(bad))
	}
	report := session.TimeoutReport()
	require.False(t, report.OK)
	require.Equal(t, "native smoke test timed out", report.Error)
	require.Equal(t, nativeSmokeChecks[:6], report.Checks)
	report.Checks[0] = "synthetic snapshot mutation"
	require.Equal(t, nativeSmokeChecks[:6], session.TimeoutReport().Checks)
	encoded, err := json.Marshal(session.TimeoutReport())
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "PRIVATE_RECORD_CANARY")
	for _, label := range nativeSmokeChecks {
		require.Contains(t, page, "'"+label+"'", "Go allowlist must match actual page labels")
	}
}

func TestNativeProgressOutOfOrderConcurrentArrival(t *testing.T) {
	var session Session
	handler := session.Wrap(http.NotFoundHandler())
	var wait sync.WaitGroup
	for length := 0; length <= len(nativeSmokeChecks); length++ {
		body, err := json.Marshal(map[string]any{"checks": nativeSmokeChecks[:length]})
		require.NoError(t, err)
		wait.Add(1)
		go func() {
			defer wait.Done()
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/__desktop_smoke__/control?action=progress", strings.NewReader(string(body))))
			if response.Code != http.StatusNoContent {
				t.Errorf("progress status %d", response.Code)
			}
		}()
	}
	wait.Wait()
	require.Equal(t, nativeSmokeChecks, session.TimeoutReport().Checks)
	require.False(t, session.TimeoutReport().OK, "all labels without final result still cannot make timeout a success")
}
