package clinicalsnapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const base = "https://synthetic-provider.example/fhir"
const patient = `{"resourceType":"Patient","id":"aa001","name":[{"text":"Synthetic López صفر"}]}`

func collection(resources ...string) []byte {
	entries := make([]string, len(resources))
	for i, raw := range resources {
		entries[i] = `{"resource":` + raw + `}`
	}
	return []byte(`{"resourceType":"Bundle","type":"collection","entry":[` + strings.Join(entries, ",") + `]}`)
}

func observation(id, subject string) string {
	return fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"final","subject":{"reference":%q},"code":{"text":"Synthetic measurement"},"valueQuantity":{"value":0.1234567890123456789,"unit":"mg/dL"},"effectiveDateTime":"2020-02","note":[{"text":"Retain café صفر 中"}]}`, id, subject)
}

func TestSnapshotPreservesExactResourceContentAndIdentity(t *testing.T) {
	obs := observation("aa001", "Patient/aa001") // Same ID in different types is valid.
	got, err := Parse(collection(patient, obs), base)
	if err != nil {
		t.Fatal(err)
	}
	if got.PatientID != "aa001" || len(got.Entries) != 2 {
		t.Fatalf("incorrect source Patient/type identity: %#v", got)
	}
	if !bytes.Equal(got.Entries[0].Raw, []byte(patient)) || !bytes.Equal(got.Entries[1].Raw, []byte(obs)) {
		t.Fatal("raw dates, numbers, text or fields were changed")
	}
	if got.Entries[0].Type != "Patient" || got.Entries[1].Type != "Observation" || got.Entries[1].FullURL != base+"/Observation/aa001" {
		t.Fatal("incorrect type/provenance")
	}
}

func TestSnapshotAcceptsExactAbsolutePatientReferenceAndProvenance(t *testing.T) {
	obs := observation("synthetic-obs", base+"/Patient/aa001")
	urn := "urn:uuid:550e8400-e29b-41d4-a716-446655440000"
	input := []byte(`{"resourceType":"Bundle","type":"collection","entry":[{"resource":` + patient + `},{"fullUrl":"` + urn + `","resource":` + obs + `}]}`)
	got, err := Parse(input, base+"/")
	if err != nil {
		t.Fatal(err)
	}
	if got.Entries[1].FullURL != urn || !bytes.Equal(got.Entries[1].Raw, []byte(obs)) {
		t.Fatal("reported resource provenance or payload was lost")
	}
}

func TestSnapshotRejectsUnprovenOrAmbiguousRecords(t *testing.T) {
	good := observation("synthetic-obs", "Patient/aa001")
	valid := collection(patient, good)
	oversized := append(append([]byte{}, valid...), bytes.Repeat([]byte(" "), MaxBytes+1-len(valid))...)
	cases := map[string][]byte{
		"Patient-only is not a clinical collection": collection(patient),
		"foreign relative Patient":                  collection(patient, observation("synthetic-obs", "Patient/001")),
		"foreign provider":                          collection(patient, observation("synthetic-obs", "https://other-provider.example/fhir/Patient/aa001")),
		"unproven identifier reference":             collection(patient, `{"resourceType":"Observation","id":"synthetic-obs","subject":{"identifier":{"value":"aa001"}}}`),
		"duplicate Patient":                         collection(patient, patient, good),
		"duplicate Observation":                     collection(patient, good, good),
		"other resource type":                       collection(patient, `{"resourceType":"Condition","id":"synthetic-obs"}`),
		"missing ID":                                collection(patient, `{"resourceType":"Observation","subject":{"reference":"Patient/aa001"}}`),
		"changed case identity":                     collection(patient, `{"resourceType":"Observation","id":"synthetic-obs","ID":"other","subject":{"reference":"Patient/aa001"}}`),
		"duplicate identity key":                    collection(patient, `{"resourceType":"Observation","id":"synthetic-obs","id":"other","subject":{"reference":"Patient/aa001"}}`),
		"ambiguous subject reference":               collection(patient, `{"resourceType":"Observation","id":"synthetic-obs","subject":{"reference":"Patient/aa001","Reference":"Patient/other"}}`),
		"invalid UTF8":                              bytes.Replace(collection(patient, good), []byte("Synthetic"), []byte{0xff}, 1),
		"oversized valid input":                     oversized,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := Parse(input, base); err == nil || got != nil {
				t.Fatal("accepted an unproven/ambiguous collection")
			}
		})
	}
	for _, origin := range []string{"http://synthetic-provider.example/fhir", base + "?token=synthetic", "https://user@synthetic-provider.example/fhir", base + "#fragment"} {
		if got, err := Parse(collection(patient, good), origin); err == nil || got != nil {
			t.Fatal("accepted unsafe origin metadata")
		}
	}
}

func TestSnapshotCountBoundary(t *testing.T) {
	resources := []string{patient}
	for i := 0; i <= MaxObservations; i++ {
		resources = append(resources, observation(fmt.Sprintf("synthetic-%d", i), "Patient/aa001"))
	}
	if got, err := Parse(collection(resources[:MaxObservations+1]...), base); err != nil || len(got.Entries) != MaxObservations+1 {
		t.Fatal("rejected the supported record-count boundary")
	}
	if got, err := Parse(collection(resources...), base); err == nil || got != nil {
		t.Fatal("accepted a collection beyond its record bound")
	}
	if !json.Valid(collection(patient, observation("valid", "Patient/aa001"))) {
		t.Fatal("invalid synthetic fixture")
	}
}

func TestSnapshotByteBoundary(t *testing.T) {
	valid := collection(patient, observation("synthetic-obs", "Patient/aa001"))
	input := append(append([]byte{}, valid...), bytes.Repeat([]byte(" "), MaxBytes-len(valid))...)
	if got, err := Parse(input, base); err != nil || got == nil {
		t.Fatal("rejected the supported byte boundary")
	}
}
