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

Next: reviewed API notification-backpressure correction, full browser/static/native builds, synthetic
native provider fixture with clinical search, actual repeated/changed clinical UI import and encrypted
restart/backup-copy recovery. Exact original export and precision-preserving Observation display still
need user-facing operational verification. No live provider, production identity or deployment is authorized
by this checkpoint.
