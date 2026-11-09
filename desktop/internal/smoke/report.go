package smoke

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

const ReportFilename = "desktop-smoke-report.json"
const maxReportBytes = 8192

// WriteReport persists one bounded native-completion receipt inside the selected
// private smoke directory. Never overwrite an earlier receipt or follow a report
// symlink. Publish only after write/sync/close, so a reader cannot accept partial
// JSON. Failed private temporary evidence is retained; post-publication temporary
// unlink errors are logged without invalidating the completed receipt. This is
// not used by normal application startup or a power-loss durability guarantee.
func WriteReport(directory string, report Report) error {
	if !filepath.IsAbs(directory) {
		return errors.New("native smoke receipt requires an absolute private directory")
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return errors.New("native smoke receipt requires a private nonsymlink directory")
	}
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if len(data) > maxReportBytes {
		return errors.New("native smoke receipt exceeds its size bound")
	}
	file, err := os.CreateTemp(directory, ".desktop-smoke-report-")
	if err != nil {
		return err
	}
	n, writeErr := file.Write(data)
	if writeErr == nil && n != len(data) {
		writeErr = io.ErrShortWrite
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	// Hard-link publication is atomic and refuses any existing target, including
	// a symlink; rename would overwrite prior evidence on Unix.
	if err := os.Link(file.Name(), filepath.Join(directory, ReportFilename)); err != nil {
		return err
	}
	if err := os.Remove(file.Name()); err != nil {
		log.Print("native smoke temporary receipt retained after publication: ", err)
	}
	return nil
}
