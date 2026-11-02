# Manual bundle upload progress and recovery

Issue #113 remains reproducible on the maintained manual-upload screen: the
selected-file preview has no pending status; another call can submit a duplicate;
HTTP failure bypasses the completion callback, retaining the preview; and an empty
accepted-file event throws before processing. Sandbox-import progress is separate.

Add an accessible live status/spinner and busy/disabled file input while a bundle
is pending. Reject duplicate or empty selections, await conversion/import responses
and clear pending state and preview in one finally block. Report conversion/import
errors in the existing alert, and remove raw error-object console dumps. There
is no invented completion percentage or new backend/provider capability.

Seven new real-component regressions exercise pending/success, failure/retry,
duplicate calls, empty selection, conversion decline, conversion failure, and
conversion followed by import. The original code fails all seven, exits 1. The
corrected focused suite passes 12 tests, exits 0. The full ordinary browser command
passes 422 tests with one existing skip, exits 0; five bridge tests pass. Added
assertions verify that the real file input disables and re-enables. No diagnostic
wrapper, forced termination, relaxed timeout/threshold or removed assertion.
[Validation receipt](upload-progress-validation.json).

All Bundle/XML files are synthetic. HTTP testing intercepts every request,
including the conversion URL; no external service or health-record database is
contacted. The existing external-conversion consent modal, MIME behavior, local
manual-import endpoint, multipart format, persistence contracts and Wails bindings
are unchanged. This does not establish external converter availability or fix the
separate native file-picker gesture reports #101/#168.

Full production/macOS package build, ZIP CRC, strict extracted signature and
all ten native smoke checks pass for fd5194396. Original output contents, inodes,
permissions, ownership, timestamps and all xattrs are restored exactly.
[Native receipt](native-fd5194396-validation.json). Protected Backend/Frontend/
Desktop checks must still pass before integration. Original-primary generated coverage recovery remains unresolved;
these runs use new private output directories.
