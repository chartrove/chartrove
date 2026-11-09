# Access-token issuance requires persistence

Pi owns [#490](https://github.com/chartrove/chartrove/issues/490) on
`task/access-token-persistence`, from accepted main
`6054e7bb50cf7b972bf21d8440410190378c4bea`, in the canonical single checkout.
Implementation checkpoint: `3f07cea49cb185d04f24ceb750ab80a001390e42`.

## Observed defect and bounded change

`CreateAccessToken` generated an access JWT, attempted to store its UUID-bound metadata, then
explicitly returned HTTP 200, success and the credential even when storage returned an error.
It also logged that raw error and provided no issuance cache restriction. This is a false-success
and unmanaged/unusable-credential issuance defect, **not a demonstrated authentication bypass**:
`RequireAuth` independently requires the stored account-bound access-token row.

The handler now returns fixed HTTP 500 / `Failed to store access token`, emits only a fixed
payload-free handler log, and does not publish credential data on a reported storage failure.
Issuance responses carry `Cache-Control: no-store`, including errors. Successful JWT generation,
metadata, expiration behavior, immutable request actor, ownership and revocation are unchanged.
No other token endpoint, signing implementation, SQL logger or expiration/input policy is audited
or certified by this change.

## Test-first evidence

- The unchanged baseline genuinely fails the new actual-handler privacy and real-store controls:
  ordinary package 1.992s, actual 200 versus expected 500, success/credential still returned,
  absent no-store, and a private synthetic error canary in the handler log.
- A boundary-only repository fake isolates the handler's raw-error disclosure. It checks the
  fixed response/log, no credential field, stable non-Gin actor context and no retained JWT.
  This is not SQL logger, password or identity-provider acceptance.
- A targeted GORM create fault exercises real SQLCipher storage and real signed-JWT middleware.
  The injected pre-create failure leaves no row; a correctly signed synthetic credential for
  that absent row receives 401. Removing the fault preserves normal metadata/owner/expiry/JWT
  behavior and successful authorization. Another account cannot revoke it; owner revocation
  causes subsequent authorization to receive 401. Fixture accounts bypass password hashing;
  this is not local password/IdP onboarding or production credential operation.
- Unchanged focused ordinary/race packages pass (4.398s / 5.588s). Complete ordinary handler/
  middleware compatibility passes (352.973s / 1.611s). Changed-handler vet and diff checks pass.
  These are distinct scopes; no complete local handler/web race or native GUI acceptance is claimed.
- Earlier 420-second race tool preparation and 300-second ordinary preparation were interrupted
  before behavioral evidence. The private `-x` trace shows a full dependency/standard-library build
  ending at link. They are not behavioral reds, test passes or grounds for a toolchain/cache repair.

## Limits and delivery

A storage error can be ambiguous about commit outcome; this contract suppresses credential delivery
on error, not a universal rollback/idempotency guarantee. The deterministic test fault is before
creation. Existing record-level authorization remains required even for correctly signed tokens.
Raw logs, synthetic JWTs, keys and databases remain private. No real provider, health record,
production identity, dependency refresh, hash/deadline relaxation, priority or OS setting change
was involved. Source/hosted/exact-main acceptance and the separate blocked native manual-import
runtime in #487 must not be conflated.

Exact head `8c098702c1d16e09854e831d04c159d871932c29` qualified all seven checks42676427229.
After ready-for-review, fresh matching-head/CLEAN/no-unresolved-thread verification preceded
protected #491 merge `8eebe540e72e846b345f3a3d427139ed2ad7cc99` at 2026-11-09T20:03:10Z.
Separate exact-main42689213196 completed success; #490 closed only for this bounded contract.
The consumer privacy/state follow-up [#492](companion-token-privacy.md) remains separately qualified.
