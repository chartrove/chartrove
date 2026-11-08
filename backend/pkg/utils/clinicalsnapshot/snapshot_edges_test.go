package clinicalsnapshot

import (
	"fmt"
	"strings"
	"testing"
)

func TestSnapshotRejectsConflictingProvenanceAndEnvelopes(t *testing.T) {
	good := observation("synthetic", "Patient/aa001")
	valid := string(collection(patient, good))
	for name, input := range map[string]string{
		"duplicate envelope": strings.Replace(valid, `"type":"collection"`, `"type":"collection","type":"collection"`, 1),
		"case envelope":      strings.Replace(valid, `"entry":`, `"Entry":`, 1),
		"missing Patient":    string(collection(good, observation("two", "Patient/aa001"))),
		"invalid ID":         string(collection(patient, observation("bad/id", "Patient/aa001"))),
		"foreign fullUrl":    `{"resourceType":"Bundle","type":"collection","entry":[{"resource":` + patient + `},{"fullUrl":"https://other.example/Observation/synthetic","resource":` + good + `}]}`,
		"trailing JSON":      valid + `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			if result, err := Parse([]byte(input), base); err == nil || result != nil {
				t.Fatal("accepted conflicting collection")
			}
		})
	}
}

func TestSearchPageRejectsAmbiguousFraming(t *testing.T) {
	good := observation("synthetic", "Patient/aa001")
	valid := `{"resourceType":"Bundle","type":"searchset","entry":[{"resource":` + good + `}]}`
	if entries, next, err := SearchPage([]byte(valid)); err != nil || len(entries) != 1 || next != "" || string(entries[0].Raw) != good {
		t.Fatal("valid raw page rejected")
	}
	for _, input := range []string{
		strings.Replace(valid, `"entry":`, `"Entry":`, 1),
		strings.Replace(valid, `"entry":`, `"entry":[],"entry":`, 1),
		strings.Replace(valid, `"resource":`, `"search":{"mode":"outcome"},"resource":`, 1),
		strings.Replace(valid, `"resource":`, `"fullUrl":42,"resource":`, 1),
		strings.TrimSuffix(valid, "}") + `,"link":[{"relation":"next","url":"?a=1"},{"relation":"next","url":"?a=2"}]}`,
		strings.TrimSuffix(valid, "}") + `,"link":[{"relation":"next","URL":"?a=1"}]}`,
	} {
		if entries, next, err := SearchPage([]byte(input)); err == nil || entries != nil || next != "" {
			t.Fatal(fmt.Sprintf("ambiguous page accepted: %s", input))
		}
	}
}
