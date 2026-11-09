package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmokeDirectoryMustBeFreshBeforeStorage(t *testing.T) {
	root := t.TempDir()
	fresh := filepath.Join(root, "fresh")
	require.NoError(t, validateSmokeDirectory(fresh))
	_, err := os.Stat(fresh)
	require.True(t, os.IsNotExist(err), "validation must not create storage")
	require.NoError(t, os.Mkdir(fresh, 0700))
	require.NoError(t, validateSmokeDirectory(fresh))
	preserved := filepath.Join(fresh, "unrelated-existing-evidence")
	require.NoError(t, os.WriteFile(preserved, []byte("synthetic preserved evidence"), 0600))
	require.Error(t, validateSmokeDirectory(fresh), "no database is not enough to prove a fresh directory")
	data, err := os.ReadFile(preserved)
	require.NoError(t, err)
	require.Equal(t, "synthetic preserved evidence", string(data))
	link := filepath.Join(root, "link")
	require.NoError(t, os.Symlink(fresh, link))
	require.Error(t, validateSmokeDirectory(link))
	require.Error(t, validateSmokeDirectory(preserved))
	require.Error(t, validateSmokeDirectory(""))
}

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
