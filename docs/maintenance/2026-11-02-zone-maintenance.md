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

Production assets and the macOS package are next built using the established
output-preserving wrapper, followed by strict signature/ZIP verification and ten
fresh synthetic native checks. Those results are pending at this source commit;
no native pass or protected-main integration is claimed yet.

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
moves to permanent evidence. The historical PR and head remain live until safe
protected integration, individual reconciliation and fresh dependency checks.
No real family or purchased Fasten records are accessed by this batch.
