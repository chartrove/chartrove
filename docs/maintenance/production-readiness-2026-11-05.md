# Daily-use readiness at the accepted search checkpoint

Baseline: protected main `a13cd39066a117a745359b84a2a50049f5088e09`
([#432](https://github.com/karamhouse/chartrove/pull/432)). All seven required
PR checks and all six [exact-main jobs](https://github.com/karamhouse/chartrove/actions/runs/42071550539)
passed. These prove source validation, not a deployed release or recovery exercise.
This is a bounded assessment of the existing application, not a new architecture,
provider registry, upstream-reuse assessment or exhaustive security audit.

## Three highest-priority blockers

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
