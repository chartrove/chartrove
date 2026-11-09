package smart

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
)

func clinicalObservation(id, patient string) string {
	return fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"final","code":{"text":"Synthetic clinical value"},"subject":{"reference":%q},"effectiveDateTime":"2020-02","valueQuantity":{"value":0.1234567890123456789,"unit":"mg/dL"},"note":[{"text":"Keep café صفر 中"}]}`, id, patient)
}

func clinicalPage(resource, next string) string {
	link := ""
	if next != "" {
		link = fmt.Sprintf(`,"link":[{"relation":"next","url":%q}]`, next)
	}
	return `{"resourceType":"Bundle","type":"searchset","entry":[{"resource":` + resource + `}]` + link + `}`
}

func connectClinical(t *testing.T, requested, granted string, handler http.Handler) (*fixture, *Manager, Connection) {
	t.Helper()
	f := newFixture(t)
	f.profile.Scopes = append(f.profile.Scopes, requested)
	f.scope = strings.Join(f.profile.Scopes, " ")
	if granted != "" {
		f.scope = granted
	}
	f.observation = handler
	m := f.manager(t, func(raw string) error {
		q := f.authorization(t, raw)
		if callback(t, f.redirect, url.Values{"state": {q.Get("state")}, "code": {"synthetic-code"}}) != 200 {
			t.Error("synthetic callback rejected")
		}
		return nil
	})
	c, err := m.Connect(context.Background(), f.profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	return f, m, c
}

func TestClinicalReadPreservesPatientAndOpaquePages(t *testing.T) {
	for _, scope := range []string{"patient/Observation.read", "patient/Observation.rs"} {
		t.Run(scope, func(t *testing.T) {
			var calls atomic.Int32
			one := clinicalObservation("aa001", "Patient/synthetic-patient")
			two := clinicalObservation("001", "Patient/synthetic-patient")
			_, m, c := connectClinical(t, scope, "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic-access-value" {
					t.Error("incorrect clinical request method/bearer")
				}
				if calls.Add(1) == 1 {
					if r.URL.Query().Get("patient") != "synthetic-patient" || r.URL.Query().Get("_count") != "100" {
						t.Error("initial search was not bound to the authorized Patient")
					}
					fmt.Fprint(w, clinicalPage(one, "?cursor=synthetic-page-two"))
				} else {
					// Continuations are server-opaque. Do not invent/rewrite parameters.
					if r.URL.RawQuery != "cursor=synthetic-page-two" {
						t.Error("rewrote the opaque continuation")
					}
					fmt.Fprint(w, clinicalPage(two, ""))
				}
			}))
			raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID)
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := clinicalsnapshot.Parse([]byte(raw), m.profiles[c.ProviderID].FHIRBaseURL)
			if err != nil || snapshot.PatientID != c.PatientID || len(snapshot.Entries) != 3 || calls.Load() != 2 {
				t.Fatalf("incomplete or incorrectly bound clinical collection: %v", err)
			}
			if !bytes.Equal(snapshot.Entries[1].Raw, []byte(one)) || !bytes.Equal(snapshot.Entries[2].Raw, []byte(two)) {
				t.Fatal("clinical dates, numeric precision, fields or identities changed")
			}
			if strings.Contains(raw, "synthetic-access-value") || strings.Contains(raw, "synthetic-page-two") {
				t.Fatal("credential or opaque paging state leaked into the snapshot")
			}
		})
	}
}

func TestClinicalReadRequiresActualSearchGrant(t *testing.T) {
	for _, entry := range []struct{ name, requested, granted string }{
		{"server withheld clinical permission", "patient/Observation.rs", "launch/patient patient/Patient.read"},
		{"read alone is not search", "patient/Observation.r", ""},
	} {
		t.Run(entry.name, func(t *testing.T) {
			var calls atomic.Int32
			f, m, c := connectClinical(t, entry.requested, entry.granted, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
			if raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID); err == nil || raw != "" || calls.Load() != 0 || f.patientCalls.Load() != 0 {
				t.Fatal("clinical retrieval admitted without actual search permission")
			}
			if _, err := m.ReadPatient(context.Background(), c.ID); err != nil {
				t.Fatal("existing permitted Patient read was broken")
			}
		})
	}
}

func TestClinicalReadRejectsInvalidPagesWithoutPartialOutput(t *testing.T) {
	for name, page := range map[string]string{
		"foreign Patient":                  clinicalPage(clinicalObservation("synthetic-obs", "Patient/other"), ""),
		"unsupported resource":             clinicalPage(`{"resourceType":"Condition","id":"synthetic"}`, ""),
		"missing subject":                  clinicalPage(`{"resourceType":"Observation","id":"synthetic"}`, ""),
		"non search collection":            `{"resourceType":"Bundle","type":"collection","entry":[]}`,
		"provider warning requires review": clinicalPage(`{"resourceType":"OperationOutcome","issue":[{"severity":"warning","code":"incomplete"}]}`, ""),
		"invalid JSON":                     `{`,
	} {
		t.Run(name, func(t *testing.T) {
			var calls atomic.Int32
			_, m, c := connectClinical(t, "patient/Observation.rs", "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				fmt.Fprint(w, page)
			}))
			if raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID); err == nil || raw != "" || calls.Load() != 1 {
				t.Fatal("did not reject the actual unsupported clinical page without partial results")
			}
		})
	}
}

type clinicalTransport func(*http.Request) (*http.Response, error)

func (f clinicalTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClinicalReadConfinesContinuationBeforeContact(t *testing.T) {
	var attempts atomic.Int32
	trap := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { attempts.Add(1) }))
	defer trap.Close()
	var calls atomic.Int32
	_, m, c := connectClinical(t, "patient/Observation.rs", "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, clinicalPage(clinicalObservation("synthetic-obs", "Patient/synthetic-patient"), trap.URL+"/fhir/Observation?cursor=synthetic"))
	}))
	original := m.client.Transport
	m.client.Transport = clinicalTransport(func(r *http.Request) (*http.Response, error) {
		if strings.HasPrefix(r.URL.String(), trap.URL) {
			attempts.Add(1) // Catches even a refused TLS handshake before dispatch.
		}
		return original.RoundTrip(r)
	})
	if raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID); err == nil || raw != "" || calls.Load() != 1 || attempts.Load() != 0 {
		t.Fatal("did not confine the server-supplied continuation before contact")
	}
}

func TestClinicalReadDisconnectCancelsActivePage(t *testing.T) {
	started := make(chan struct{})
	_, m, c := connectClinical(t, "patient/Observation.rs", "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	type result struct {
		raw string
		err error
	}
	done := make(chan result, 1)
	go func() { raw, err := m.ReadClinicalSnapshot(context.Background(), c.ID); done <- result{raw, err} }()
	select {
	case <-started:
	case <-done:
		t.Fatal("clinical reader returned before starting the permitted page")
	case <-time.After(2 * time.Second):
		t.Fatal("clinical reader did not start")
	}
	m.Disconnect(c.ID)
	select {
	case got := <-done:
		if got.err == nil || got.raw != "" {
			t.Fatal("disconnected read returned clinical data")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("disconnect did not cancel the active clinical request")
	}
}
