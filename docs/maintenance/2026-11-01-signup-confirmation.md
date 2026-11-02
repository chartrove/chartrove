# Password confirmation - issue 124

The regular signup page and first-administrator wizard lacked a confirmation
input and accepted repeated submissions while a request was pending. Add a
required masked confirmation field to each, disable the button on mismatch,
and guard the submit handler as well as the button. Keep confirmation in the
component, outside the existing API user payload. The wizard previously sent an
unused empty password_confirm property; the backend DTO has no such field.
No backend, database, key-management, dependency or bridge contract changes.

Ten browser component regressions reproduce nine failures on the old behavior
and pass with the repair. They cover both rendered forms, missing/mismatched
confirmation (including form submission), existing payload shape, and duplicate
pending requests. The first test attempt had a union-type inference compile
error, corrected before the meaningful red/green comparison.

Production desktop_prod Angular build exits 0 with existing 9.92 MB bundle and
CommonJS/browser warnings. The first full suite disconnects before executing
any tests (30-second no-message limit); it is a failed run. A serialized diagnostic
rerun keeps all assertions and timeouts, adds DEBUG logging and an explicit
basePath to the original Karma configuration, and exits 0: 400 passing tests,
one existing skip; Chrome exits 0 and cleanup finishes. The cause of the earlier
browser disconnect remains unverified. No timeout or assertion was relaxed.

Raw logs are retained under dist/validation/2026-11-01-consolidation, including
the failing run. GitHub required CI still gates integration. This is browser
component and build coverage, not a claim of a complete native UI walkthrough.
