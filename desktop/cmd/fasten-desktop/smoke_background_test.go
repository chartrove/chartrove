package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackgroundSmokeRequiresExplicitSmokeMode(t *testing.T) {
	for _, mode := range []struct {
		smoke, background bool
		valid             bool
	}{
		{false, false, true},
		{true, false, true},
		{true, true, true},
		{false, true, false},
	} {
		err := validateSmokeBackground(mode.smoke, mode.background)
		if mode.valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
