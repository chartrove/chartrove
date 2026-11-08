// Package clinicalsnapshot guards the bounded raw Patient/Observation collection
// shared by native retrieval and the subsequent owner-authorized import layer.
// Validation here never establishes native account/source authorization.
package clinicalsnapshot

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const MaxBytes = 16 << 20
const MaxObservations = 5000

var ErrInvalid = errors.New("invalid clinical snapshot")

type Entry struct {
	Type    string
	ID      string
	FullURL string
	Raw     json.RawMessage
}

type Snapshot struct {
	PatientID string
	Entries   []Entry
}

var resourceID = regexp.MustCompile(`^[A-Za-z0-9.-]{1,64}$`)
var uuidURN = regexp.MustCompile(`^urn:uuid:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Parse validates identity and provenance without interpreting clinical fields.
// Raw resource bytes and entry order are retained. This is not authorization.
func Parse(raw []byte, fhirBase string) (*Snapshot, error) {
	fhirBase = strings.TrimRight(fhirBase, "/")
	u, err := url.Parse(fhirBase)
	if err != nil || len(fhirBase) > 2048 || strings.ContainsAny(fhirBase, "\r\n\t #") || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawPath != "" || u.Opaque != "" || len(raw) > MaxBytes || !utf8.Valid(raw) {
		return nil, ErrInvalid
	}
	bundle, err := object(raw, "resourceType", "type", "entry")
	if err != nil || text(bundle["resourceType"]) != "Bundle" || text(bundle["type"]) != "collection" {
		return nil, ErrInvalid
	}
	var entries []json.RawMessage
	if json.Unmarshal(bundle["entry"], &entries) != nil || len(entries) < 2 || len(entries) > MaxObservations+1 {
		return nil, ErrInvalid
	}
	snapshot := &Snapshot{}
	seen, uris := map[string]bool{}, map[string]bool{}
	for _, rawEntry := range entries {
		entry, err := object(rawEntry, "resource", "fullUrl")
		if err != nil {
			return nil, ErrInvalid
		}
		resource, err := object(entry["resource"], "resourceType", "id", "subject")
		kind, id := text(resource["resourceType"]), text(resource["id"])
		if err != nil || (kind != "Patient" && kind != "Observation") || !resourceID.MatchString(id) || seen[kind+"/"+id] {
			return nil, ErrInvalid
		}
		seen[kind+"/"+id] = true
		if kind == "Patient" {
			if snapshot.PatientID != "" {
				return nil, ErrInvalid
			}
			snapshot.PatientID = id
		}
		uri := fhirBase + "/" + kind + "/" + id
		if provided, exists := entry["fullUrl"]; exists {
			value := text(provided)
			if value != uri && !uuidURN.MatchString(value) {
				return nil, ErrInvalid
			}
			uri = value
		}
		if uris[uri] {
			return nil, ErrInvalid
		}
		uris[uri] = true
		snapshot.Entries = append(snapshot.Entries, Entry{Type: kind, ID: id, FullURL: uri, Raw: entry["resource"]})
	}
	if snapshot.PatientID == "" {
		return nil, ErrInvalid
	}
	for _, entry := range snapshot.Entries {
		if entry.Type != "Observation" {
			continue
		}
		resource, _ := object(entry.Raw, "resourceType", "id", "subject")
		subject, err := object(resource["subject"], "reference")
		reference := text(subject["reference"])
		if err != nil || (reference != "Patient/"+snapshot.PatientID && reference != fhirBase+"/Patient/"+snapshot.PatientID) {
			return nil, ErrInvalid
		}
	}
	return snapshot, nil
}

// Reject duplicate keys and case-folded aliases for identity-bearing fields.
// encoding/json's normal last-key/case-insensitive behavior is unsuitable here.
func object(raw []byte, critical ...string) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	first, err := d.Token()
	if err != nil || first != json.Delim('{') {
		return nil, ErrInvalid
	}
	result := map[string]json.RawMessage{}
	for d.More() {
		token, err := d.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return nil, ErrInvalid
		}
		if _, exists := result[key]; exists {
			return nil, ErrInvalid
		}
		for _, canonical := range critical {
			if strings.EqualFold(key, canonical) && key != canonical {
				return nil, ErrInvalid
			}
		}
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return nil, ErrInvalid
		}
		result[key] = value
	}
	if end, err := d.Token(); err != nil || end != json.Delim('}') {
		return nil, ErrInvalid
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, ErrInvalid
	}
	return result, nil
}

func text(raw json.RawMessage) string {
	var value string
	_ = json.Unmarshal(raw, &value)
	return value
}

// SearchPage retains raw resources and the opaque next link. Unsupported search
// modes (including OperationOutcome warnings) must not become partial success.
func SearchPage(raw []byte) ([]Entry, string, error) {
	if len(raw) > MaxBytes || !utf8.Valid(raw) {
		return nil, "", ErrInvalid
	}
	page, err := object(raw, "resourceType", "type", "entry", "link")
	if err != nil || text(page["resourceType"]) != "Bundle" || text(page["type"]) != "searchset" {
		return nil, "", ErrInvalid
	}
	var rows, links []json.RawMessage
	if json.Unmarshal(page["entry"], &rows) != nil || len(rows) > MaxObservations {
		return nil, "", ErrInvalid
	}
	entries := make([]Entry, 0, len(rows))
	for _, row := range rows {
		entry, err := object(row, "resource", "fullUrl", "search")
		if err != nil {
			return nil, "", ErrInvalid
		}
		if search, exists := entry["search"]; exists {
			fields, err := object(search, "mode")
			if err != nil || text(fields["mode"]) != "match" {
				return nil, "", ErrInvalid
			}
		}
		uri := ""
		if full, exists := entry["fullUrl"]; exists {
			uri = text(full)
			if uri == "" {
				return nil, "", ErrInvalid
			}
		}
		entries = append(entries, Entry{FullURL: uri, Raw: entry["resource"]})
	}
	if link, exists := page["link"]; exists && json.Unmarshal(link, &links) != nil {
		return nil, "", ErrInvalid
	}
	next := ""
	for _, rawLink := range links {
		link, err := object(rawLink, "relation", "url")
		if err != nil || text(link["relation"]) == "" || text(link["url"]) == "" {
			return nil, "", ErrInvalid
		}
		if text(link["relation"]) == "next" {
			if next != "" {
				return nil, "", ErrInvalid
			}
			next = text(link["url"])
		}
	}
	return entries, next, nil
}
