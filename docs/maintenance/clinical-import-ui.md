# Explicit clinical Sources UI: local checkpoint

Pi owns `task/clinical-import-ui`, initially stacked on `task/scoped-smart-bridge` and API #473.
This is not merged or native clinical acceptance.

The Sources card preserves the original native Patient string inside a bounded collection rather than
serializing a parsed numeric Patient. It offers a separate Patient/Observation retrieval action; the
complete raw collection also remains a string. Clinical parsing in the renderer is for display/counts,
not authority or reconstruction. The backend/native parser remains authoritative.

The card shows the backend-verified receiving UUID and username, requires an unchecked-by-default
recording-authority confirmation, and passes expected UUID plus `confirmed: true` to the atomic API.
Each newly retrieved clinical preview resets confirmation. A missing Observation collection is distinct
from a complete zero-Observation response; missing records are not deletions. Provider-origin metadata
remains owner-declared, not provider attestation.

Read, import and cleanup controls exclude overlapping expensive work. Identity is rechecked before read
and write and after asynchronous responses. Stale previews are cleared rather than applied to a different
profile. Import uncertainty retains exactly the same raw preview for safe digest replay; historical
acknowledgment does not claim old data is current. Failed clinical refresh retains only the prior complete
preview, labeled as such, never a partially received replacement. Cancellation and teardown release native
credentials; native cleanup failures ask for application closure rather than claiming release succeeded.

Focused Angular run passed 34/34: real component templates and DOM confirmation, private-canary masking,
Patient/Observation decimal lexical preservation, duplicate/overlapping-action guards, cancellation/late
results, pre/post identity-switch isolation, raw API request contract, retry/historical acknowledgment and
cleanup progress/failure. Tests use controlled service/native promises and HTTP testing transport, not real
provider identities or a running desktop. New tests and implementation were developed together.

The reviewed API notification-backpressure correction is now integrated without bypassing its unresolved
hosted qualification. The native synthetic fixture now grants/searches Observations and preserves exact
Patient decimals. For each of its independent origins, the first two clinical reads are byte-identical;
third onward changes one high-precision decimal, deliberately exercising new/replay/update. This scenario
is synthetic-only and does not contact real providers or alter normal startup profiles. Its tests validate
both origins with the real manager/parser and refuse reads after disconnect.

All native internal race packages passed (host 25.412s, shell 1.217s, SMART 6.904s, fixture 1.324s).
Frontend structural/compiler validation passed; combining that long validation with the first full-browser
run exceeded a 180-second tool budget just after Chrome connected, so that invocation is not a passing
browser suite. An unchanged standalone rerun passed 733 tests with one skip; existing test-template
`fa-icon` console diagnostics and browser-support/deprecated-builder warnings remain, not suppressed.
Native macOS object-target linker warnings remain. No older-platform claim is made.

Next: matching packaged GUI, actual repeated/changed clinical UI import and encrypted
restart/backup-copy recovery. Exact original export and precision-preserving Observation display still
need user-facing operational verification. No live provider, production identity or deployment is authorized
by this checkpoint.
