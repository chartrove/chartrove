# Atomic clinical snapshot persistence — internal slice

Exclusive Pi ownership continues in the canonical primary checkout on
`task/clinical-snapshot-persistence`, initially stacked on the #466 reader. This is **not** the clinical
import UI or production acceptance. There is no new HTTP route, native bridge method, provider
registration, background job or external connection.

## Contract and provenance

The optional `ClinicalSnapshotRepository` revalidates a bounded raw Patient/Observation collection at
the database boundary. It derives a manual source from the authenticated local owner, exact Patient ID
and trailing-slash-normalized FHIR base, preserving the existing Patient-only source namespace. It never
uses claimed origin metadata as native authorization, merges owners/Patients/endpoints, overwrites a
provider credential, or relabels an ambiguous legacy Patient. An existing source label is retained.

Each accepted new collection commits these changes in one transaction:

1. Reserve the active owner row before transactional reads, then create/reuse the correctly scoped source.
2. Insert/update all resources and derived search projections, clearing removed projection fields.
3. Retain exact decimal lexemes and unknown fields in raw FHIR. Normal resource JSON can undergo the
   existing serializer's whitespace/HTML escaping; it is not the byte-for-byte original.
4. Create only the validated Observation-to-Patient associations. Preserve other raw references without
   inventing target records or cross-source relationships. Do not delete absent Observations.
5. Store the **exact original collection bytes**, SHA-256, local importer, import time and claimed FHIR
   base in an append-only `ClinicalSnapshotRecord`. Acquisition is always `owner-supplied-snapshot`, not
   provider attestation. The hash detects changed originals on export; it is not an authorization secret
   or a signature from a provider.
6. Append a payload-free source audit last. Any final write failure rolls the entire import back. This
   internal repository emits no success events; a future caller must publish only after commit.

Exact historical replay acknowledges a previously recorded collection without changing current clinical
records or adding another original/audit. It must not roll a newer import back. This does not claim that
all previously imported current rows still exist, nor that a newly supplied snapshot is newer according
to provider versions. A future UI must say "previously imported; no changes applied", not "already current".

Original history is **owner-only**, even for administrators and source delegates with EDIT grants.
Metadata queries omit payloads in SQL and JSON. Imported resources do not receive local-edit provenance;
existing source-level sharing can expose their current read-only views without granting history/import
rights. Partial dates remain in raw FHIR; timeline sorting never invents a missing day/time/timezone.
Numeric search projections use the existing FHIRPath engine and are not a precision-preserving analytics
store. Raw values/originals, not those projections, are the fidelity evidence.

## Provenance-only update regression

PR #472's automated source review identified a real gap despite seven green checks on `0c74d4d305`:
entry `fullUrl` can change while the nested resource JSON is identical. Comparing only clinical bytes
incorrectly left current `source_uri` at its old value. The protected unresolved-thread rule blocked the
merge; no bypass or auto-merge was used.

`TestClinicalSnapshotEntryProvenance` failed before correction for both Patient and Observation
(canonical URL to valid UUID URN, expected one update but got zero). The unchanged check now includes
entry provenance. The same tests plus the complete focused persistence suite passed under the race
detector (15.861s), including unchanged resource bytes, exact changed original and historical replay
that cannot roll the new URI back. Fresh final-head protected validation is still required.

## Migration and lifecycle

Additive frozen migration `20261109120000` creates the original-evidence table. It does not rewrite existing
FHIR resources, credentials, key material, encryption format or provider registrations. It uses the
installation's configured store; existing insecure-default/configuration acceptance is still a separate gate.

Source and account deletion now run transactionally and share the owner reservation with this new importer.
They remove the associated originals alongside current records. An injected final source/account-delete
failure must restore both. Concurrent account deletion cannot leave new originals behind a deleted owner.
Other legacy import/write paths do not yet share this reservation; whole-application lifecycle-race
acceptance remains open. Backups and restore copies include the new table without a second patient store.
No clinical history expiry/deletion policy is silently introduced.

Rollback is to a pre-change independently checked encrypted backup and compatible application build,
with writes stopped and keys/configuration preserved separately. Do not drop evidence tables or run a
schema downgrade as a shortcut. No real installation was migrated, restored or deleted during this work.

## Synthetic evidence

Permanent tests exercise real encrypted SQLite, real projections and transactions, and only synthetic
principals/resources. Direct fixture insertion supplies dummy principals; these are storage-authorization
tests, not a claim to have tested password/OIDC authentication.

- Byte-for-byte originals, numeric/extension/date fidelity, source provenance and payload-free SQL traces.
- Initial import, changed values with removed quantity projections, exact replay and old-snapshot replay.
- Distinct local owners, endpoints and Patients; denied foreign/admin/delegate history access.
- Read-only current views despite EDIT grants; ambiguous legacy Patient refusal and Patient-only source reuse.
- Final-audit failure, cancellation, final source/account-delete failure, and full transactional rollback.
- Two independent database connections importing concurrently, and concurrent account deletion/import.
- Encrypted close/reopen and backup into a private new bundle; boot a **copy**, compare restored originals
  and updated records, and prove the immutable backup artifact did not change. Corrupt original hashes fail closed.
- Absence from a later collection does not delete a pre-existing Observation.

Fresh final-metadata focused race suite passed (25.102s). The permanent maximum-collection test passed
(32.018s): 5,000 Observations plus Patient, all 5,001 current resources and exact original preserved within
the two-minute import budget. A separate synthetic whole-import diagnostic measured 31.193s for the same
count, 1,204,010 original bytes; whole `go run` invocation max RSS was 679,870,464 bytes including build
processes, **not** a desktop runtime memory profile. Do not extrapolate this to every payload or machine.

The full local backend suite passed before the final acquisition-label/metadata-only delta (database
247.976s, web 85.288s, handler 78.392s). Focused final-delta race/max tests, `go vet` for changed packages,
and `git diff --check` subsequently passed. Full final-head hosted validation remains required. PostgreSQL
runtime, older native operating systems and the future end-to-end import action are not yet qualified.

The first fixture runs correctly hit the backup private-parent gate and then mistook its bundle directory
for a database file. Fixtures were corrected to private bundle creation and restoring a separate copy;
the backup acceptance gates were not weakened. A named-type compile error and an overly broad test row
selector were also corrected before the passing runs. Exact retained logs are private readiness evidence.
No independent peer/security certificate or deployed acceptance is claimed.
