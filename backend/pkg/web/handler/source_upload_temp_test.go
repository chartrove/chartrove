package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func syntheticUploadContext(t *testing.T, raw []byte, spill bool) *gin.Context {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("file", "synthetic-record.bin")
	require.NoError(t, err)
	_, err = file.Write(raw)
	require.NoError(t, err)
	require.NoError(t, form.Close())
	request := httptest.NewRequest(http.MethodPost, "/synthetic-upload", &body)
	request.Header.Set("Content-Type", form.FormDataContentType())
	if spill {
		require.NoError(t, request.ParseMultipartForm(0))
	}
	t.Cleanup(func() {
		if request.MultipartForm != nil {
			require.NoError(t, request.MultipartForm.RemoveAll())
		}
	})
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = request
	return context
}

func TestManualUploadPreservesPrivateTempDirectory(t *testing.T) {
	for _, spill := range []bool{false, true} {
		name := "memory"
		if spill {
			name = "disk-backed"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.Chmod(dir, 0o700))
			t.Setenv("TMPDIR", dir)
			t.Setenv("TMP", dir)
			t.Setenv("TEMP", dir)
			before, err := os.Stat(dir)
			require.NoError(t, err)
			raw := []byte("Synthetic café صفر 血液\x00\r\nno clinical data")
			file, err := storeFileLocally(syntheticUploadContext(t, raw, spill))
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, file.Close()); require.NoError(t, os.Remove(file.Name())) })
			require.Equal(t, dir, filepath.Dir(file.Name()))
			position, err := file.Seek(0, io.SeekCurrent)
			require.NoError(t, err)
			require.Zero(t, position, "existing import consumers read from the beginning")
			copied, err := io.ReadAll(file)
			require.NoError(t, err)
			require.Equal(t, raw, copied)
			info, err := file.Stat()
			require.NoError(t, err)
			after, err := os.Stat(dir)
			require.NoError(t, err)
			if runtime.GOOS != "windows" {
				require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
				require.Equal(t, before.Mode().Perm(), after.Mode().Perm(), "upload must not chmod an existing parent directory")
			}
		})
	}
}

func TestManualUploadFailureRemovesOnlyItsOwnTempFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	t.Setenv("TMP", dir)
	t.Setenv("TEMP", dir)
	context := syntheticUploadContext(t, []byte("Synthetic unavailable upload"), true)
	require.NotEmpty(t, context.Request.MultipartForm.File["file"])
	// Remove only the synthetic multipart spill file before opening its header.
	// This deterministically exercises the source-open failure and target cleanup.
	require.NoError(t, context.Request.MultipartForm.RemoveAll())
	sentinel := filepath.Join(dir, "synthetic-retained-sentinel")
	require.NoError(t, os.WriteFile(sentinel, []byte("preserve"), 0o600))
	file, err := storeFileLocally(context)
	require.Error(t, err)
	require.Nil(t, file)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, filepath.Base(sentinel), entries[0].Name())
	content, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	require.Equal(t, "preserve", string(content))
}
