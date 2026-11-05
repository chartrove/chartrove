# Daily-use readiness at the accepted search checkpoint

Baseline: protected main `a13cd39066a117a745359b84a2a50049f5088e09`
([#432](https://github.com/karamhouse/chartrove/pull/432)). All seven required
PR checks and all six [exact-main jobs](https://github.com/karamhouse/chartrove/actions/runs/42071550539)
passed. These prove source validation, not a deployed release or recovery exercise.
This is a bounded assessment of the existing application, not a new architecture,
provider registry, upstream-reuse assessment or exhaustive security audit.

## Current production-first priorities after #434

Current source baseline is `17a70f542f1f5b744584a2f3089388c9ce0dd98b`.
Protected #432/#433/#434 and their exact-main workflows passed all required checks;
the original assessment below is retained as the historical search checkpoint.
The governing product direction uses [#425](https://github.com/karamhouse/chartrove/issues/425),
[#430](https://github.com/karamhouse/chartrove/issues/430) and the separate,
unmerged [roadmap proposal #431](https://github.com/karamhouse/chartrove/pull/431).
This source lane does not adopt/merge that draft or duplicate its provider mapping.

1. **No accepted, recoverable daily-use deployment.** No source-bound deployed
   digest, effective ingress/TLS configuration, all-hop callback-log acceptance or
   independently held native administrator password/key custody is established.
   Encrypted reopen tests do not prove a consistent backup restored without the
   original installation or Authentik. This blocks dependable recovery from device,
   storage or IdP failure. **Operational:** approve the isolated synthetic staging
   artifact, owners/custody, backup method, restore and cleanup allowlist, then run
   the existing #425 native/OIDC/family-permission/recovery matrix. No live action
   is authorized by this assessment.

2. **Clinical ingestion and editing do not yet have complete provenance acceptance.**
   A reproduced source defect makes the legacy owner PATCH use SDK ingestion:
   editing an explicitly local record removes its local-edit marker and disables
   a family's existing delegated EDIT capability. **Integrated in #434:**
   separate owner-only existing-row updates from ingestion, preserving
   local provenance and omitted presentation metadata, reindexing changed fields,
   auditing atomically and checking live native/source ownership at the write.
   SDK replacement still removes local eligibility; imported records are never
   promoted. The complete-resource replacement body remains the owner API contract;
   this is not a field-merging fix for the legacy lab editor or historical #186.
   **Still unverified:** end-to-end FHIR/C-CDA and household-specific provider/payer
   import, storage, retrieval and attribution. Legacy external C-CDA conversion is
   warned and unvalidated; no local replacement is adopted here. The private P0
   household list, current registrations and evidence belong to #430/#431, not
   historical platform-name availability. No real-record/provider test is performed.

3. **Confidentiality and lifecycle recovery remain operationally unproven.**
   #428 removed application callback Referer output and #433 disabled automatic
   ORM traces. Gateway/proxy/IdP/browser logs, artifact/backup access and retention,
   deletion and recovery require acceptance at the chosen runtime. Native logout
   removes the browser token; copied password JWTs retain their bounded expiry.
   OIDC native logout/unlink revoke sessions, while provider-side revocation follows
   its documented expiry window. **Operational:** explicitly accept the session
   window, prove harmless canary redaction at every hop, and verify synthetic grant
   revocation, source/account deletion and restoration without broadening access.
   Source tests and green CodeQL do not establish these deployed controls.

The owner-edit defect is the highest-priority reproduced source correction that
can proceed independently of those approvals. Validation uses synthetic accounts
and encrypted SQLite/disposable PostgreSQL, plus native-authenticated embedded
Desktop HTTP flows. The [owner-update implementation](../../backend/pkg/database/owned_resource_update.go),
[two-engine regressions](../../backend/pkg/database/owned_resource_update_test.go),
[legacy HTTP regression](../../backend/pkg/web/handler/owner_edit_provenance_test.go)
and [embedded native flow](../../backend/pkg/web/embedded_owned_update_test.go)
retain the exact storage identity and current authorization boundaries.
No second database, service, dependency, schema migration,
provider connection or AI data destination is introduced. Chartrove remains the
authoritative record system with Go/Gin, Angular, Wails, SQLCipher and optional
PostgreSQL. Conversational AI remains outside it; future MCP needs separately
approved read-only patient/source scope, live revocation and data-destination rules.

The required Go CodeQL job exposed a separate validation prerequisite: GitHub
cancelled it before setup with an unsupported macOS 14 runner configuration.
[GitHub retired those hosted labels on November 2](https://github.blog/changelog/2026-10-01-github-actions-macos-14-runner-image-retirement/).
Only the Desktop and Go CodeQL image labels move to supported `macos-15`; all
seven required checks, commands, toolchains, permissions and timeouts are retained.
The exact-head replacement must pass full protected validation. Hosted macOS 15
checks do not certify native UI compatibility on older macOS versions.

## Source-qualified owner mutation contract

A separate synthetic two-source reproduction found a data-integrity defect:
legacy PATCH/DELETE selected an arbitrary source when the owner had the same
resource type/ID in multiple sources. Both returned success while modifying or
deleting the other record. Practitioner PUT had the same first-match lookup.
This correction binds client actions to the selected source and rejects ambiguity.

| Existing endpoint | Optional explicit selector |
| --- | --- |
| `PATCH /api/secure/resource/fhir/:resourceType/:resourceId` | JSON `source_id` beside the existing `resource_raw` / sort fields |
| `PUT /api/secure/practitioners/:practitionerId` | JSON `source_id` beside the existing `resource` |
| `DELETE /api/secure/resource/fhir/:resourceType/:resourceId` | One `source_id` query parameter |

Omitting the selector preserves the legacy request/response shape when exactly
one owned live record matches. Two candidates return **409 without a write**;
missing/foreign/deleted targets return **404**. Explicit malformed/null/empty/nil
UUIDs, duplicate DELETE selectors, malformed query escapes and trailing JSON return
**400** rather than silently falling back. The source hint never proves access:
resolution reads at most two metadata rows under native UUID/live-source ownership,
then the update/delete rechecks native/source/record liveness at the write boundary.
Admin roles and active source grants cannot authorize these owner-only operations.

Deletion removes only the selected record and its local marker. Its `delete_owned`
audit is atomic; an unavailable audit rolls back both, and notifications follow
commit. The public legacy repository deletion method also rejects ambiguity.
This retains the existing row-deletion contract; it does not add account-wide
erasure, purge historical audit/provenance or certify a backup-retention policy.
Practitioner PUT uses the existing owner-edit path, preserving local eligibility
and exact raw clinical bytes instead of the SDK ingestion/JSON re-encoding path.
Canonical value casts support legacy PostgreSQL text/UUID columns without a migration.

The UI carries known source IDs through editing, practitioner routes and deletion.
Equal resource IDs in different sources have distinct selection/row keys. Batch
responses remove only actual successful tuples, even when failures arrive last;
route/save responses remain bound to their original source context. Relative
MedicationRequest references are matched within the selected medication's source.
The touched practitioner-loading flow no longer automatically dumps the records
to the console. Other legacy graph/history disambiguation, full-record lab-edit
preservation and comprehensive browser/logging acceptance remain separate work.

Regression evidence: [target resolver/delete](../../backend/pkg/database/owned_resource_mutation_test.go),
[three owner transports](../../backend/pkg/web/handler/source_scoped_mutation_test.go)
and [native encrypted transport](../../backend/pkg/web/embedded_source_scoped_mutation_test.go).
Tests use generated accounts/records, encrypted SQLite and disposable pinned
PostgreSQL. Client diagnostics evaluate the actual TypeScript methods; permanent
browser regressions and complete protected checks remain required. No provider
registry, production data, external converter, model/index service, dependency,
database schema or live configuration changes are part of this correction.

## Original priorities before #433

1. **Automatic database logs can expose clinical values and credentials.**
   Both repository constructors inherited GORM's process-wide logger and enabled
   interpolated SQL at DEBUG. Default error/slow-query traces also include values;
   setting the application to INFO/WARN is insufficient. Existing parameter filters
   protect selected search/import/registration sessions, not all older APIs or
   GORM's early Scan recorder. Synthetic actual-constructor regressions reproduce
   clinical text/credential disclosure at INFO, WARN and DEBUG (a nanosecond
   slow-query threshold deterministically exercises the WARN branch).
   **Source work, unblocked:** explicitly discard automatic ORM output in both
   constructors before opening/migrating storage. Keep application lifecycle
   diagnostics and database errors returned to callers. Do not promise that this
   covers every application, driver, gateway, browser or IdP log. Retain protected
   validation, both database engines and the desktop module graph.

2. **A restorable, independently recoverable backup has not been demonstrated.**
   The desktop deliberately does not retain the encryption key. SQLite uses WAL;
   copying only a live database file is not a proven consistent backup. Startup
   performs migrations. Existing encrypted-reopen/wrong-key tests prove key/unlock
   behavior, not recovery from loss of the original installation. An IPS export is
   not evidence of a complete database/attachments/accounts/grants/identity restore.
   No accepted backup-and-restore receipt, recovery objective or assigned key and
   native-password custodian is recorded in the maintained guides/acceptance plan.
   **Operational acceptance, approval-gated:** choose custody, backup method and
   retention, then restore a synthetic encrypted dataset into an isolated copy
   without relying on Authentik or the original directory. Verify record bytes,
   attachments, native UUIDs, permissions, revocations and provenance, wrong-key
   rejection, restart and rollback. PostgreSQL needs its own approved database and
   storage protection/recovery procedure; SQLite encryption does not cover it.
   This assessment authorizes no backup, restore, migration or key operation on
   existing data. A future backup helper needs its own narrowly reviewed scope.

3. **No specific daily-use artifact and runtime boundary has operational acceptance.**
   The README still advertises a development preview/ad hoc desktop build, not an
   accepted production artifact. Passing CI does not identify an installed digest,
   effective server TLS/ingress configuration or safe logs at every hop. Standalone
   factory network/encryption fallbacks differ from the checked-in example and
   desktop defaults. The proposed staging deployment, Authentik owners, native
   password/key custodians and artifact remain unassigned/unaccepted; do not infer
   them from source tests. **Operational acceptance, approval-gated:** approve one
   exact artifact and isolated synthetic staging scope, verify effective settings,
   native login/unlock independently of the IdP, callback/log confidentiality,
   session/UUID binding and active source authorization across logout, expiration,
   revocation and deletion. Optional OIDC must remain disabled until its separate
   acceptance passes; native-only operation does not require an IdP. Desktop
   embedded HTTP tests do not substitute for native UI/package acceptance.

## Evidence and acceptance boundaries

- Storage/logging: [SQLite constructor](../../backend/pkg/database/sqlite_repository.go),
  [PostgreSQL constructor](../../backend/pkg/database/postgres_repository.go),
  [logger regressions](../../backend/pkg/database/sql_log_privacy_test.go) and
  [PostgreSQL counterpart](../../backend/pkg/database/sql_log_privacy_postgres_integration_test.go).
  The source correction adds no dependency, schema, migration, key rotation,
  authorization change or new service. Reverting it restores unsafe automatic SQL
  output; use synthetic reproduction and retained source for diagnostics instead.
- [Local data/security guide](../LOCAL_DATA_SAFETY.md),
  [desktop storage/recovery](../../desktop/README.md#data-and-startup),
  [encrypted reopen tests](../../backend/pkg/web/session_duration_startup_test.go),
  [wrong-key tests](../../backend/pkg/web/embedded_test.go) and
  [registration/native recovery tests](../../backend/pkg/web/registration_policy_test.go).
- [Bounded Authentik/native acceptance](oidc-operational-acceptance.md) and
  [#425](https://github.com/karamhouse/chartrove/issues/425) retain the exact pending
  owner decisions and private staging brief. No live IdP, provisioning, deployment,
  login, recovery, provider connection or real-record operation is authorized.

The three priorities are ranked by immediate disclosure risk, irrecoverable loss
and unverified runtime exposure. They do not certify clinical correctness or
claim every inherited defect is resolved. Existing manual-record/date limitations
remain in [unresolved acceptance](2026-11-02-unresolved-acceptance.md). External
C-CDA conversion remains explicitly warned and unvalidated; local FHIR imports
are the supported development path. Provider compatibility and original-library /
Medplum reuse belong to the user's separate roadmap work. No inherited provider
issue is closed or reorganized here. Retain Go, Angular, Wails and encrypted
storage; no Rust rewrite, embedded RAG/search/model infrastructure or external
clinical-data processing is part of this fix.
