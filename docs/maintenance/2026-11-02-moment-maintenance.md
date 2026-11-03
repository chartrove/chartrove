# Moment maintenance patch

Paused #249 proposes Moment 2.31.0. Its exact current tip is archived in a verified
bundle; the earlier intake tip and original archive remain linked separately.
The current replacement changes only the existing Moment lock block, using the
verified npm registry URL/integrity. Package declarations, all other lock bytes
and 72 other installed direct versions remain unchanged. Frozen install passes.
[Recovery](moment-head-recovery.json), [validation](moment-validation.json).

The official 2.31.0 tagged source and published runtime bytes match. All original
source-archive files remain unchanged. The unmodified upstream suite passes
3,943 tests and 165,058 assertions; an external reporter confirms zero skips or
todos without filtering tests or altering assertions. A preliminary preload did
not capture count metadata despite a passing suite; that diagnostic is retained.
The actual unchanged Chartrove date/period utility module, TypeScript compiled
and run with each pinned Moment, returns identical outputs for 16 date and five
period fixtures per timezone in UTC and America/New_York. This is compatibility
evidence, not a new partial-date/Period policy or backend reindex migration.

Ten Node checks, the ordinary 487-test browser suite (one existing skip),
production and static Storybook pass. Only yarn.lock differs among the complete
frontend input inventories. Coverage stays private, compiler inputs remain
unchanged and original generated-documentation absence restores. Karma delayed
exit after Chrome shutdown warnings, then exited 0 naturally before the guarded
termination attempt reached the PID. No signal, retry, timeout or assertion
change occurred; the raw warnings and successful exit are retained.

Upstream GHSA-4p3w-j4w9-5jqw/CVE-2026-17495 is patched by 2.31.0 and concerns
server-side non-string moment.locale input. No direct locale setter is used in
Chartrove source; no application exploit path or clean dependency audit is
claimed. No Go/Wails/API/database or application source changes are included.

A fresh synthetic Patient/date/vital fixture was created through actual embedded
signup/import routes with zero outbound HTTP. Native packaging, ten-check smoke,
UI and protected integration remain pending. #249 and its head stay open/live.
Raw evidence: `dist/validation/2026-11-01-consolidation/moment-review/`.
