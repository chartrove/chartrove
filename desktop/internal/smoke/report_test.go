package smoke

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteReportPrivateCreateOnce(t *testing.T) {
	for _, report := range []Report{
		{OK: true, Checks: []string{"synthetic native binding"}},
		{OK: false, Checks: []string{"synthetic earlier check"}, Error: "synthetic native failure"},
	} {
		dir := t.TempDir()
		require.NoError(t, os.Chmod(dir, 0700))
		require.NoError(t, WriteReport(dir, report))
		path := filepath.Join(dir, ReportFilename)
		info, err := os.Lstat(path)
		require.NoError(t, err)
		require.True(t, info.Mode().IsRegular())
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
		original, err := os.ReadFile(path)
		require.NoError(t, err)
		var decoded Report
		require.NoError(t, json.Unmarshal(original, &decoded))
		require.Equal(t, report, decoded)
		require.Error(t, WriteReport(dir, Report{OK: true, Checks: []string{"must not replace evidence"}}))
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, original, after)
	}
}

func TestWriteReportRefusesSymlinksAndUnsafeDirectory(t *testing.T) {
	root := t.TempDir()
	private := filepath.Join(root, "private")
	require.NoError(t, os.Mkdir(private, 0700))
	outside := filepath.Join(root, "outside")
	require.NoError(t, os.WriteFile(outside, []byte("preserved synthetic evidence"), 0600))
	require.NoError(t, os.Symlink(outside, filepath.Join(private, ReportFilename)))
	require.Error(t, WriteReport(private, Report{OK: true}))
	data, err := os.ReadFile(outside)
	require.NoError(t, err)
	require.Equal(t, "preserved synthetic evidence", string(data))
	link := filepath.Join(root, "directory-link")
	require.NoError(t, os.Symlink(private, link))
	require.Error(t, WriteReport(link, Report{OK: true}))
	public := filepath.Join(root, "public")
	require.NoError(t, os.Mkdir(public, 0755))
	require.NoError(t, os.Chmod(public, 0755))
	require.Error(t, WriteReport(public, Report{OK: true}))
	_, err = os.Stat(filepath.Join(public, ReportFilename))
	require.True(t, os.IsNotExist(err))
	require.Error(t, WriteReport("relative-directory", Report{OK: true}))
}

func TestWriteReportBoundsBeforeCreatingFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0700))
	require.Error(t, WriteReport(dir, Report{Error: strings.Repeat("synthetic", maxReportBytes)}))
	_, err := os.Stat(filepath.Join(dir, ReportFilename))
	require.True(t, os.IsNotExist(err))
}
