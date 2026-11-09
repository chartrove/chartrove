package smart

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
)

func TestClinicalReadFailureBoundsAndNoPartialSuccess(t *testing.T) {
	for _, scenario := range []string{"cycle", "page limit", "duplicate record", "second page failure", "page too large", "wrong endpoint", "duplicate envelope"} {
		t.Run(scenario, func(t *testing.T) {
			var calls atomic.Int32
			_, m, c := connectClinical(t, "patient/Observation.rs", "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := calls.Add(1)
				obs := clinicalObservation(fmt.Sprintf("synthetic-%d", n), "Patient/synthetic-patient")
				switch scenario {
				case "cycle":
					fmt.Fprint(w, clinicalPage(obs, "?cycle=one"))
				case "page limit":
					fmt.Fprint(w, clinicalPage(obs, fmt.Sprintf("?cursor=%d", n)))
				case "duplicate record":
					next := ""
					if n == 1 {
						next = "?cursor=two"
					}
					fmt.Fprint(w, clinicalPage(clinicalObservation("same", "Patient/synthetic-patient"), next))
				case "second page failure":
					if n == 1 {
						fmt.Fprint(w, clinicalPage(obs, "?cursor=two"))
					} else {
						w.WriteHeader(http.StatusServiceUnavailable)
					}
				case "page too large":
					fmt.Fprint(w, clinicalPage(obs, ""), strings.Repeat(" ", clinicalPageBytes))
				case "wrong endpoint":
					fmt.Fprint(w, clinicalPage(obs, "Patient?cursor=two"))
				case "duplicate envelope":
					fmt.Fprint(w, strings.Replace(clinicalPage(obs, ""), `"entry":`, `"entry":[],"entry":`, 1))
				}
			}))
			if raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID); err == nil || raw != "" {
				t.Fatal("failed/partial clinical read returned data")
			}
			expected := int32(1)
			if scenario == "cycle" || scenario == "second page failure" {
				expected = 2
			}
			if scenario == "duplicate record" {
				expected = 2
			}
			if scenario == "page limit" {
				expected = clinicalMaxPages
			}
			if calls.Load() != expected {
				t.Fatalf("request bound: got %d want %d", calls.Load(), expected)
			}
		})
	}
}

// Each response fits the per-page contract. Only the accumulated response-body
// or resource count crosses the whole-read limit; no partial collection may pass.
func TestClinicalReadAggregateLimits(t *testing.T) {
	for _, scenario := range []string{"response bytes", "resource count"} {
		t.Run(scenario, func(t *testing.T) {
			var calls atomic.Int32
			_, m, c := connectClinical(t, "patient/Observation.rs", "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := calls.Add(1)
				next := fmt.Sprintf("?cursor=%d", n)
				if scenario == "response bytes" {
					if n == 9 {
						next = ""
					}
					page := clinicalPage(clinicalObservation(fmt.Sprintf("synthetic-%d", n), "Patient/synthetic-patient"), next)
					fmt.Fprint(w, page, strings.Repeat(" ", clinicalPageBytes-1-len(page)))
					return
				}
				const perPage = clinicalsnapshot.MaxObservations/10 + 1
				if n == 10 {
					next = ""
				}
				var entries []string
				for i := 0; i < perPage; i++ {
					entries = append(entries, `{"resource":`+clinicalObservation(fmt.Sprintf("synthetic-%d-%d", n, i), "Patient/synthetic-patient")+`}`)
				}
				page := fmt.Sprintf(`{"resourceType":"Bundle","type":"searchset","entry":[%s]`, strings.Join(entries, ","))
				if next != "" {
					page += fmt.Sprintf(`,"link":[{"relation":"next","url":%q}]`, next)
				}
				fmt.Fprint(w, page+"}")
			}))
			if raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID); err == nil || raw != "" {
				t.Fatal("aggregate limit returned a partial or oversized collection")
			}
			if scenario == "response bytes" {
				if calls.Load() != 8 {
					t.Fatalf("aggregate response-byte bound: got %d requests, want 8", calls.Load())
				}
			} else if calls.Load() != 10 {
				t.Fatalf("aggregate record bound: got %d requests, want 10", calls.Load())
			}
		})
	}
}

func TestClinicalReadUnavailableConnectionMakesNoRequests(t *testing.T) {
	for _, scenario := range []string{"disconnect", "expiry", "cancellation", "close", "invalid base"} {
		t.Run(scenario, func(t *testing.T) {
			var calls atomic.Int32
			f, m, c := connectClinical(t, "patient/Observation.rs", "", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch scenario {
			case "disconnect":
				m.Disconnect(c.ID)
			case "expiry":
				m.mu.Lock()
				entry := m.connections[c.ID]
				entry.summary.ExpiresAt = time.Now().Unix() - 1
				m.connections[c.ID] = entry
				m.mu.Unlock()
			case "cancellation":
				cancel()
			case "close":
				m.Close()
			case "invalid base":
				m.mu.Lock()
				profile := m.profiles[c.ProviderID]
				profile.FHIRBaseURL += "#"
				m.profiles[c.ProviderID] = profile
				m.mu.Unlock()
			}
			if raw, err := m.ReadClinicalSnapshot(ctx, c.ID); err == nil || raw != "" || calls.Load() != 0 || f.patientCalls.Load() != 0 {
				t.Fatal("unavailable connection made clinical requests")
			}
		})
	}
}
