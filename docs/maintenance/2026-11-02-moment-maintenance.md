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
signup/import routes with zero outbound HTTP. Fresh native packaging at c109657a6 passes ZIP CRC, strict signature and ten
synthetic checks. All 22,287 compiled/reused/native frontend files match; the
artifact delta is main bundle/map, index and generated third-party license
notice. The notice diff is retained. Original output content/stat fields and all
xattrs restore exactly, with zero provenance changes.
[Native receipt](native-c109657a6-validation.json).

Native UI signs into the actual synthetic account and renders Moment Synthetic,
26 years, and Heart Rate 72.0 beats/minute dated June 15, 2026. Both raw resources
and original seed files remain unchanged. The app closes and ps proves process
absence, with no Cua call after close. Tab keyboard focus exposed WebView content
at startup and after sign-in; intervening states are recorded without attributing
a cause or adding an application workaround.
[PR #380](https://github.com/karamhouse/chartrove/pull/380) merged after all three
protected checks passed at e5901aac92ff6bc1ab34c7324873ca142f56c85a. #249 is
individually closed as implemented. The archived exact historical head and
integrated task head are retired after fresh dependency/reachability/lease checks.
[Retirement receipt](moment-retirement-1.json).
Raw evidence: `dist/validation/2026-11-01-consolidation/moment-review/`.
