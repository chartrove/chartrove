# Preserve retained frontend coverage during tests

Ordinary Karma runs wrote directly to frontend/coverage/fastenhealth, so generating
a report replaced any earlier report there. This caused the disclosed original-
primary preservation gap. A warning and session-local wrapper do not protect the
next ordinary ng test invocation after a restart.

Karma now chooses a fresh private OS temporary directory for each invocation and
prints its location. An explicit CHARTROVE_COVERAGE_DIR selects an operator-chosen
path when a retained artifact location is wanted; it is resolved absolutely and
may be replaced during that run. Normal commands leave frontend/coverage alone.
HTML, text-summary and LCOV reporters, instrumentation, tests, thresholds, browser
flags and timeouts are unchanged. No application/runtime dependency is changed.
CONTRIBUTING explains the output and explicit-path behavior.

The regression uses only an exclusively created synthetic prior report after an
atomic absence check. It is clearly marked NOT recovered historical evidence.
Before the fix, the real three-test WebCrypto coverage invocation overwrites that
synthetic index. After the fix, the same invocation writes real HTML/LCOV output
to a fresh private directory (mode 0700) and preserves the prior index bytes,
permissions and mtime. The full suite passes 404 tests with one existing skip,
uses a distinct fresh directory, and still preserves the synthetic prior report.
The explicit output-directory option also passes with real report output and
unchanged prior synthetic contents/metadata.

After verification, only the exactly inventoried synthetic fixture tree was moved
to the task's retained diagnostic location. frontend/coverage is absent again.
No historical or health-record data was used, overwritten or restored in this
regression. In particular, this does NOT recover the original primary coverage;
that still needs the previously described user restore or backup access.
[Validation receipt](coverage-isolation-validation.json). Raw logs and proofs are
retained under dist/validation/2026-11-01-consolidation/coverage-isolation-*.

This test-output-only change leaves the application runtime validated at 4f891ef11
unchanged. [Latest native package](native-4f891ef11-validation.json) includes the
integrated utility graph, passes strict signature/CRC and ten synthetic WebView
checks, and restores original generated outputs with identical contents, stat
metadata and all xattrs. All protected-main checks remain mandatory for this PR.
