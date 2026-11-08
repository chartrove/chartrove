package smart

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
)

const clinicalMaxPages = 64
const clinicalPageBytes = 2 << 20

var errClinicalRead = errors.New("SMART clinical snapshot could not be completed; no partial snapshot is available")

// ReadClinicalSnapshot prepares only a bounded Patient/Observation collection.
// It neither persists records nor exposes a host bridge/import route. Paging
// queries are opaque; bearer credentials stay on the pinned Observation endpoint.
func (m *Manager) ReadClinicalSnapshot(parent context.Context, id string) (string, error) {
	m.mu.Lock()
	c, exists := m.connections[id]
	if m.closed || !exists {
		m.mu.Unlock()
		return "", errClinicalRead
	}
	if time.Now().Unix() >= c.summary.ExpiresAt {
		delete(m.connections, id)
		c.cancel()
		m.mu.Unlock()
		return "", errClinicalRead
	}
	// SMART v1 .read includes search; v2 .r alone does not. Requested
	// permissions and the public summary are never substitutes for the grant.
	if !slices.Contains(c.scopes, "patient/Observation.read") && !slices.Contains(c.scopes, "patient/Observation.rs") {
		m.mu.Unlock()
		return "", errClinicalRead
	}
	p := m.profiles[c.summary.ProviderID]
	m.active.Add(1)
	m.mu.Unlock()
	defer m.active.Done()
	endpoint, err := url.Parse(p.FHIRBaseURL + "/Observation")
	if err != nil || len(p.FHIRBaseURL) > 2048 || !httpsURL(p.FHIRBaseURL) || endpoint.ForceQuery {
		return "", errClinicalRead
	}
	deadline := time.Now().Add(2 * time.Minute)
	if expiry := time.Unix(c.summary.ExpiresAt, 0); expiry.Before(deadline) {
		deadline = expiry
	}
	ctx, cancel := context.WithDeadline(parent, deadline)
	defer cancel()
	stop := context.AfterFunc(c.ctx, cancel)
	defer stop()
	patient, err := m.ReadPatient(ctx, id)
	if err != nil {
		return "", errClinicalRead
	}
	patientURI := p.FHIRBaseURL + "/Patient/" + url.PathEscape(c.summary.PatientID)
	records := []clinicalsnapshot.Entry{}
	current := *endpoint
	current.RawQuery = url.Values{"patient": {c.summary.PatientID}, "_count": {"100"}}.Encode()
	visited := map[string]bool{}
	total := len(patient)
	for pageNumber := 0; pageNumber < clinicalMaxPages; pageNumber++ {
		if ctx.Err() != nil || visited[current.String()] || total >= clinicalsnapshot.MaxBytes {
			return "", errClinicalRead
		}
		visited[current.String()] = true
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, current.String(), nil)
		if err != nil {
			return "", errClinicalRead
		}
		req.Header.Set("Authorization", "Bearer "+c.access)
		req.Header.Set("Accept", "application/fhir+json")
		limit := min(clinicalPageBytes, clinicalsnapshot.MaxBytes-total)
		raw, err := m.readClinicalPage(req, int64(limit))
		if err != nil {
			return "", errClinicalRead
		}
		total += len(raw)
		entries, next, err := clinicalsnapshot.SearchPage(raw)
		if err != nil || len(entries)+len(records) > clinicalsnapshot.MaxObservations {
			return "", errClinicalRead
		}
		if len(entries) > 0 {
			checked, err := clinicalsnapshot.Parse(clinicalCollection(patient, patientURI, entries), p.FHIRBaseURL)
			if err != nil || checked.PatientID != c.summary.PatientID {
				return "", errClinicalRead
			}
			records = append(records, checked.Entries[1:]...)
		}
		if next == "" {
			result := clinicalCollection(patient, patientURI, records)
			checked, err := clinicalsnapshot.Parse(result, p.FHIRBaseURL)
			if err != nil || checked.PatientID != c.summary.PatientID || ctx.Err() != nil {
				return "", errClinicalRead
			}
			return string(result), nil
		}
		if len(next) > 4096 || strings.ContainsAny(next, "\r\n\t #") {
			return "", errClinicalRead
		}
		reference, err := url.Parse(next)
		if err != nil {
			return "", errClinicalRead
		}
		u := current.ResolveReference(reference)
		if u.Scheme != endpoint.Scheme || !strings.EqualFold(u.Host, endpoint.Host) || u.EscapedPath() != endpoint.EscapedPath() || u.User != nil || u.Fragment != "" || u.Opaque != "" || u.RawPath != "" {
			return "", errClinicalRead
		}
		current = *u
	}
	return "", errClinicalRead
}

// Count the actual page body, including padding, against the whole-read bound.
func (m *Manager) readClinicalPage(req *http.Request, limit int64) ([]byte, error) {
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, errClinicalRead
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errClinicalRead
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, errClinicalRead
	}
	return raw, nil
}

// Resource JSON is copied verbatim, rather than decoding clinical numbers into
// float64 or normalizing date/text fields. Only the framing URI is serialized.
func clinicalCollection(patient, patientURI string, observations []clinicalsnapshot.Entry) []byte {
	var b bytes.Buffer
	b.WriteString(`{"resourceType":"Bundle","type":"collection","entry":[`)
	write := func(uri string, raw []byte) {
		encoded, _ := json.Marshal(uri)
		b.WriteString(`{"fullUrl":`)
		b.Write(encoded)
		b.WriteString(`,"resource":`)
		b.Write(raw)
		b.WriteByte('}')
	}
	write(patientURI, []byte(patient))
	for _, entry := range observations {
		b.WriteByte(',')
		if entry.FullURL == "" {
			// Omitted search fullUrl is derived by the shared parser, not guessed
			// from an unrelated record/source. Keep it omitted during validation.
			b.WriteString(`{"resource":`)
			b.Write(entry.Raw)
			b.WriteByte('}')
		} else {
			write(entry.FullURL, entry.Raw)
		}
	}
	b.WriteString(`]}`)
	return b.Bytes()
}
