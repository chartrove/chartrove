# Optional installation registration and local standalone startup

## Ownership and boundary

Pi owns #498 in the sole checkout, branch `task/optional-platform-startup`, from
accepted `5128397733d2a693f916e6ae59b1678ffa3a4a28`. Validated local product checkpoint:
`7ebfac7493121cd362f5c53713b3557abf514527`. Protected acceptance and separate exact-main
qualification are recorded below.
No existing configuration, profile, registration consent/opt-in policy, installed key,
health record, provider account or deployment was changed.

Existing healthy automatic installation registration remains. This repair makes
optional platform availability nonessential to the local listener; it is not a new
telemetry opt-out policy or proof of a network-silent application.

## Actual findings, not preparation failures

- Actual AppEngine.Start in an owned isolated subprocess, fresh synthetic SQLCipher
  repository and child-only in-memory platform transport: HTTP503 caused panic/exit2
  before local health readiness. Genuine 1.166s red (child0.16s), expected registration
  attempt marker and private log distinguish it from setup/storage failure. No real
  platform/provider request, profile or health record was involved.
- Connected real SaveSystemSettings fault: second-row failure returned nil and left a
  mixed installation-ID/secret pair. Genuine 1.061s red. Fixing only caller handling with
  a fake repository would have missed this persisted false success.
- Added missing-secret-row control exposed a draft gap, not a separate accepted-head
  finding: 0.939s red, nil save and surviving ID overwritten. The final draft validates
  both required credential rows before any write and never rewrites unknown settings.
- Draft helper initially formatted raw local settings errors: two load/save privacy
  assertions failed 0.874s. Fixed messages now retain original causes through Unwrap
  without formatting potentially credential-bearing causes into startup diagnostics.
- PostgreSQL preparation initially failed 1.014s before any container/repository:
  exact reviewed digest absent on the current local Unix daemon. After endpoint/
  linux-arm64 verification, only that pinned fixture was acquired with an empty private
  Docker client config. No context switch, daemon/cache repair or alternate image.

## Implementation contract

- Start continues only for the classified optional remote registration error. Local
  settings load/save errors remain critical, including wrapped storage causes; they
  are never treated as optional merely because a cause resembles a remote error.
- Registration uses a copied caller HTTP client, a ten-second request/client ceiling
  and preserves tighter configured client timeouts. No global HTTP/client/logger change
  or background retry. Rejected statuses are closed without reading their bodies;
  successful replies are bounded to 64KiB+1 and validated before saving.
- Failed network/status/remote-contract diagnostics are payload-free. Success requires
  literal success and a nonblank bounded installation ID; secret stays OPTIONAL per
  the existing remote model, bounded and preserved exactly rather than invented.
  IDs/secrets/status reason/body are not logged. Cached complete credentials avoid HTTP.
- ID and secret save in one context-bound transaction. Actual write errors propagate;
  each existing row must update once. Missing/invalid credential rows fail before any
  surviving row changes. Only those two keys are selected, preserving unknown/future
  setting values; no migration, user-settings refresh or broad schema cleanup.
- A storage-reported error is not success. Remote issuance is not rolled back by local
  failure, and no automatic retry/idempotency or universal rollback is promised.

## Qualification at this checkpoint

- Shared real storage controls cover second-write failure/rollback/error identity,
  ordinary replacement, cancellation, missing row and retained unknown setting on
  SQLCipher and real PostgreSQL. Focused PostgreSQL race passed 8.827s using the existing
  pinned PostgreSQL17 fixture: local Unix socket, generated private password, ephemeral
  loopback publication, bounded resources/tmpfs, no host database volume. Only the owned
  container cleanup succeeded; its public image remains cached. Not production
  PostgreSQL at-rest/TLS/operations acceptance.
- Actual startup subprocess becomes healthy in normal, non-STANDBY mode after the
  synthetic optional503. Owned child completion/cleanup is observed; its no-op event
  bus is a fixture boundary, not a whole live CLI/event-delivery acceptance test.
- Remote success/omitted-secret/unavailable/rejected/malformed/unconfirmed/oversize,
  close/zero rejected reads/no credential logs, cached no-network, local storage
  criticality/cause privacy, no fake saved identity and tighter40ms client controls pass.
  Signing-key/session-startup guards remain. Final focused database/web races
  2.678/2.766s; tagged changed-package vet and diff checks pass.
- Complete serialized ordinary database/web qualification passed161.040/63.327s,
  captured Go-test exit0 with unchanged15-minute package/fixture budgets. All seven
  tested source hashes match; source stayed frozen through actual exit capture,
  documentation edits separate. Owned wrapper62594 subsequently observed absent.
- Full serialized native core races passed cmd1.592/host22.915/shell30.746/SMART6.917/
  fixture1.256/smoke1.184s, with source hashes unchanged. Existing macOS27.2-object /
  11.0-link-target warnings remain, not an older-OS certificate. No complete database/
  web race, new native package/GUI/runtime or actual platform registration is claimed
  by these local checks; hosted qualification below is separate.

## Protected acceptance and independent main qualification

Exact `42219322fb5f756eb28dbfae4e528f024e42d923` passed all seven42720034229:
Backend6m19s, Frontend4m04s, Desktop12m05s, actions50s, Go15m19s, JS2m14s and
aggregate3s. Actual Go tracing, not a retry/infrastructure relabel. Fresh matching
head/CLEAN/no unresolved threads/ready verification preceded protected #499 merge
`390005131e536a32d07d15453a6e7317b43cdb5f` at 2026-11-09T23:40:59Z. Separate
exact-main42722429637 passed; #498 is closed boundedly. Healthy automatic registration
and the optional-secret contract remain, not a new telemetry policy or global offline
certificate. The #500 SDK request-lifetime follow-up is independent.

Normal encrypted-record/provider behavior, TLS/egress, graceful standalone shutdown,
GUI/capture consent, independently operated recovery, #474 global lifetime/privacy,
#479 maximum clinical race and blocked #487 runtime/negative qualification remain
separate. #496 HTTPfailed-body source/main acceptance does not certify this platform
installation path or every SDK error surface.
