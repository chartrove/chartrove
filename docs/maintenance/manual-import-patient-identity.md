# Manual import Patient identity

The pinned source SDK uses `strings.TrimLeft(id, "Patient/")` when selecting a
Patient. This removes a character cutset rather than a prefix: valid distinct
FHIR IDs such as `aa001` and `001` both become `001`. Source creation keys include
native owner, endpoint and Patient, so this loss can merge manual sources and
their resource-upsert/grant scopes.

## Actual reproduction

Test-first commit `54745bc846805e329d13af7451400f7c7272ebe2`, based on
`54b6050f948a6a2b2c52a897115e1f9679586d8e`, adds an unchanged synthetic embedded
regression using a disposable encrypted database and in-process HTTP.
[Desktop job 128390989623](https://github.com/karamhouse/chartrove/actions/runs/42368309405/job/128390989623)
executed the case and failed six semantic assertions: the first ID changed,
both imports reused a source UUID, the second observation replaced the first,
an existing VIEW delegate saw the replacement and second Patient, and the
replacement persisted after restart. The case completed in 5.24 seconds; the
embedded package completed in 79.079 seconds. Anonymous and revoked-grant access
were denied. This is a demonstrated source-scope collision within one owner,
not demonstrated cross-native-owner access or a claim about real records.

## Correction

Keep the SDK's existing format/version validation and import pipeline. Separately
decode JSON identity fields, preserve the exact canonical FHIR Patient ID, and
require one distinct Patient per upload. Bundle and NDJSON remain supported;
repeated identical Patient IDs are permitted. Multi-patient or invalid-identity
uploads fail before source creation. The additional local parse does not rewrite
or log clinical payloads; the original file is rewound for the SDK. JSON Bundle
decoding still requires memory proportional to the upload, as does the existing
SDK; no constant-memory or large-document performance acceptance is claimed.

Reuse an existing manual source only when its live stored Patient identity
matches. A different platform, missing Patient, multiple Patients or mismatched
Patient produces a generic conflict before credential/resource writes. This
also prevents a new canonical `001` import from appending to a legacy source
whose key is `001` but whose stored Patient is `aa001`.

The two existing assertions that deliberately pinned the SDK's truncated ID now
require the original ID. Their import, source/origin separation, repeat-update,
raw-record and temporary-file checks remain in place. Direct SMART snapshot
validation/dedup/retry is unchanged. SDK, GoFHIR and SQLCipher pins, schemas,
native accounts, grants, encryption and ownership predicates are unchanged.

## Remaining acceptance boundaries

Existing sources, records and grants are not renamed, split, migrated or repaired.
Already mixed records and historical sharing require separately authorized
assessment/reconciliation; this source correction does not certify them safe.
Reimporting an affected original ID can create a separate canonical source, not
silently transfer historical data or grants. An existing source with no proven
Patient is refused, including after an incomplete plain manual import; recovery
of that ambiguous state is an explicit remaining limitation. No automatic retry,
deletion or permission change is performed.

Plain uploads still have no provider namespace beyond the existing endpoint:
equal IDs from different providers cannot be distinguished without explicit
origin/source provenance. Source selection and SDK writes retain their existing
transaction boundaries; concurrent legacy writers are not made safe here.

This is manual-file integrity work. It does not establish Epic clinical import,
repeat synchronization, packaged native UI operation, PostgreSQL operational
recovery, Authentik acceptance (#425), release publication or production approval.
