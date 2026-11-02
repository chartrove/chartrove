# Zone.js maintenance within Angular 22 (#234)

The isolated update selects Zone.js 0.16.3 with manifest range ~0.16.0, accepted
by Angular core 22.2.4's existing ~0.15.0 || ~0.16.0 peer requirement. The
historical angular-monorepo branch now contains only this two-file dependency
patch; no completed framework migration is repeated. All 72 other direct
installed versions and all other lockfile bytes remain unchanged. Frozen pinned
Yarn installation passes without altering the patched manifest/lock.

The published package has unchanged exports, declaration entry point and runtime
dependency fields. Its SHA-512 and SHA-1 match canonical registry metadata and
all 138 published files match installation. The reviewed changelog includes
event-listener options preservation, null-timer guards, Promise/testing changes,
and symbol/event-map hardening. Legacy IE/Edge shims are removed; acceptance
here is the existing Angular 22 and current macOS WebView environment, not a
claim of untested browser support. Original Fasten attribution and licenses,
Wails contracts, database schema and provider behavior remain unchanged.

All 429 browser tests pass with one existing skip, and ten Node checks pass.
The 138 existing NG01354 form warnings match the preceding Fishery run; no
threshold, assertion, timeout, test skip or browser flag was relaxed. Full static
Storybook passes through the pinned Yarn environment. Its generated metadata is
retained and original documentation.json absence restored; frontend/coverage
also remains absent. [Validation receipt](zone-validation.json).

The output-preserving wrapper built production assets and a fresh macOS package
at source `5ad6c590af04e481b23f41850e8622bec2ec5d5e`. ZIP CRC, strict
extracted signature and all ten synthetic native WebView checks pass. All 22,287
produced frontend files match the native asset tree byte-for-byte. Original
output contents, recorded stat fields and every xattr are restored exactly;
there are zero provenance changes. The retained ZIP SHA-256 is
`cd423b6dd1d11ee13aa377a777ec3802ea670082b24b906f12ca2b20d7f7c1b7`.
[Native receipt](native-5ad6c590a-validation.json). Production differences from
the preceding package are retained in the comparison receipt; no cross-browser
or live-provider acceptance is inferred. Protected-main CI passed in run 41519084415; protected PR #368 merged at
`8ed4fda151da2ca5ab70f1378aa470dbc56ca596`. #234 was individually
closed as implemented. Its exact bundled head and merged task head were
manually retired after fresh open-PR dependency checks, dry run and atomic
exact-tip leases. [Retirement](zone-retirement-1.json).

## Preservation

The complete installed 0.15.1 package and root Yarn integrity metadata are in a
checksum-recorded archive, with all archived regular-file bytes independently
compared, original stat fields and a native xattr sidecar retained.
[Installed recovery](zone-installed-preservation.json).
The exact current #234 tip 57fc11c6ea8be48c4c8a738a027cebc650a297f4 is in a
separate verified incremental bundle, with main-reachable prerequisites and all
five unique objects proved present in original pack bytes. The original intake
Angular head also remains in history.bundle. [Head recovery](zone-head-recovery.json).

Two prechecks were corrected without changing application behavior: npm view
returned a one-element list rather than the expected object, and macOS Python
lacks os.listxattr. Exact 0.16.3 registry metadata and native xattr capture resolve
those producer assumptions. The filesystem precheck stopped before changing
source or installing dependencies; the original failed diagnostics remain.

Raw package, install, validation, Storybook, archive and diagnostic evidence is
retained under dist/validation/2026-11-01-consolidation/zone-review/ and
zone-validation/. Execution-time temporary paths remain verbatim after atomic
moves to permanent evidence. The historical proposal and head are now reconciled as recorded above.
No real family or purchased Fasten records are accessed by this batch.
