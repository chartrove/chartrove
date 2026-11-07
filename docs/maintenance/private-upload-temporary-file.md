# Private temporary files for manual uploads

[Issue #446](https://github.com/karamhouse/chartrove/issues/446) records a concrete upload defect on base `b52c080b29c1eb3fae183835b78521644dbee6b5`. The helper creates a private temporary file and then reopens it through Gin 1.12's `SaveUploadedFile`. That dependency also calls `os.Chmod` on the destination's existing parent directory. The independently owned disposable regression demonstrates private-directory permissions changing from `0700` to `0750` for both memory-backed and disk-backed multipart uploads. A process lacking permission to change its temp directory can also fail before parsing an import.

Copy the multipart stream into the already-open `0600` temporary file, close the source stream and rewind the destination for existing readers. On any open/copy/close/seek failure, close/remove only the owned destination and retain the existing generic error contract. Do not reopen a pathname or change parent-directory permissions. Filename handling, supported import formats, payload bytes, authentication, authorization, encryption and provenance are unchanged. PR #443's independent logging correction is not incorporated here.

Focused tests cover byte-for-byte Unicode/binary content, the returned cursor at offset zero, file permissions, unchanged parent permissions, memory/disk multipart sources and missing-spill-file cleanup with an unrelated synthetic sentinel preserved. Windows exercises content/cursor/cleanup; Unix-specific mode assertions run on Linux/macOS.

```sh
go test ./backend/pkg/web/handler -run '^TestManualUpload' -count=1
```

An initial #139 synthetic native-import investigation exposed an HTTP 500 before record parsing. That generic failure is retained as evidence, without claiming its underlying OS error was captured. The separate disposable-directory regression establishes this source defect. Follow-up native checks passed for three synthetic string Observations (including Unicode and a numeric-looking string), import/readback, indexed string query, unrelated-user denial and zero outbound HTTP. The dedicated temp directory remained `0700`. These checks use independently controlled synthetic users, encryption credentials and a dedicated private temp directory; no further attempts to mutate a host-wide temporary directory are made.

No schema, dependency, configuration, API or frontend change. Required protected Backend, Frontend, Desktop and CodeQL results and exact-head independent review are recorded on the PR; separate owner acceptance is required before merge. Private red/green source and receipts remain under the existing consolidation evidence root in `private-upload-temporary-file-446-20261106`.

Rollback is a normal protected source revert; it must not change any database, patient data, credentials or host directory permissions. Both retention worktrees and all unique historical refs remain held. Existing #438 concurrency/non-atomic transaction limits, #158 unresolved result-reference provenance and #425 operational acceptance remain separate. No deployment, live recovery, Authentik operation, provider connection or real medical-record processing is performed.
