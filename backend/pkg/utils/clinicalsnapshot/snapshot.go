// Package clinicalsnapshot guards the bounded raw Patient/Observation collection
// shared by native retrieval and the subsequent owner-authorized import layer.
// Validation here never establishes native account/source authorization.
package clinicalsnapshot

import (
	"encoding/json"
	"errors"
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

// Parse is test-first scaffolding. No production caller is enabled yet.
func Parse(raw []byte, fhirBase string) (*Snapshot, error) {
	return nil, ErrInvalid
}
