package smart

import (
	"context"
	"errors"
)

// ReadClinicalSnapshot is test-first scaffolding for a bounded internal reader.
// It is not exposed through the host bridge or a public clinical import path.
func (m *Manager) ReadClinicalSnapshot(parent context.Context, id string) (string, error) {
	return "", errors.New("SMART clinical snapshots are unavailable")
}
