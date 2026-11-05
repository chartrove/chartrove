# Bounded Authentik and native-recovery acceptance

This plan implements the source-verification portion of the operator's OIDC
acceptance request. It does not authorize live actions. The authentication model
and upstream attribution remain in [OIDC authentication](oidc-authentication.md).
Registration policy is documented [separately](registration-policy.md).

## Source contracts and prerequisites

Use the exact protected main revision and deployed artifact digest in the receipt.
Source tests are evidence about synthetic behavior, not live Authentik acceptance.
The deployed revision, callback proxy, operator recovery and real provider login
remain **UNVERIFIED** until the approved operational exercise succeeds.

- A native UUID owns records, grants and API tokens. An exact issuer/subject binds
  explicitly to that UUID after local password proof. Email, groups and IdP roles
  do not provision or merge accounts, assign administrators or grant medical access.
- New local, API and OIDC JWTs carry `native_user_id`; `sub` remains the username.
  Middleware checks the current UUID and database lookups retain it as a predicate,
  including if a username is deleted and recreated during a request.
- Password JWTs issued before UUID binding require one fresh password login after
  upgrading. They cannot safely prove immutable account ownership; historical
  client-settable creation times are insufficient evidence. HS256, expiry, native
  passwords, UUIDs, encryption keys and existing accounts are not migrated or rotated.
  Existing API tokens and OIDC JWTs without the added claim retain their independent
  UUID-bound database/revocation checks. New tokens of all three kinds include it.
- Local password logout clears the browser token; it does not revoke a copied JWT
  server-side. Its existing bounded expiry and account deletion remain authoritative.
  OIDC logout revokes the native session in the database. Unlink revokes all sessions
  for that identity, including in-flight/re-link races. Offline logout cannot confirm
  revocation. None of these actions terminates the IdP session.
- Provider-side logout/disablement is checked on the next provider authentication,
  not every Chartrove request. Existing sessions expire no later than the verified
  ID-token expiry or native lifetime. Back-channel logout, introspection, refresh,
  and IdP end-session redirects are **UNSUPPORTED**. Record the accepted revocation
  window; do not claim immediate provider revocation.
- Local login and encrypted database unlock do not depend on provider availability.
  Forgotten-password UI is not password reset. Recovery requires an independently
  held working native administrator password and the encryption-key recovery owner.
  Desktop supports native login/unlock, not web OIDC callbacks.

Before any operational test, identify the operator, deployment owner and recovery
owner; exact Chartrove version/digest; a staging origin with matching API/frontend
basepath; exact issuer/discovery URL and HTTPS redirect; public/confidential client
mode; secret-store reference (never its value); and maximum session/revocation window.
Review TLS, trusted routing, callback-query redaction, backups and restart/rollback
access. Multi-replica start/complete requires sticky routing for the pending flow.
Use standard [Authentik OAuth2/OIDC configuration](https://docs.goauthentik.io/add-secure-apps/providers/oauth2/)
with stable subjects and the `openid` scope, without tenant-specific source code.

## Approval boundary and synthetic fixture

**Approval required before execution:** creating/changing Authentik applications,
users, secrets, credentials, provider policies or sessions; enabling/reconfiguring
OIDC in a deployment; logging in to or modifying production accounts; restarting or
rolling back a running service. Obtain a concrete environment/action allowlist and
recovery owner first. No approval is implied by this document or by source/CI passes.

Prefer a disposable staging database and dedicated synthetic IdP application/users.
After approval, create native A (record owner), B (ordinary delegate), C (independent
local recovery administrator), and a second unrelated patient account D. Use dummy
DocumentReference/Observation records, duplicate synthetic emails, distinct UUIDs,
one VIEW grant, one EDIT grant on an eligible local source, an expired grant, an
unshared source and a browser-imported read-only document. Never reuse real patients,
existing credentials or clinical payloads. Record fixtures for bounded cleanup;
cleanup of IdP/database objects is itself limited to the approved synthetic allowlist.

## Ordered acceptance matrix

Stop on unexpected authorization or identity results. Preserve sanitized failure
receipts; do not change production settings or retry against real patient records.

| Step | Expected outcome | Source evidence; operational status |
| --- | --- | --- |
| 1. Prove native recovery before enabling OIDC | C signs in with its native password; independently reopen encrypted staging DB with correct key; wrong key fails; public registration may remain closed | Embedded encryption/recovery and registration HTTP regressions; live UNVERIFIED |
| 2. Explicitly link A | Local session and fresh native password required; provider callback returns to exact origin; issuer/subject binds to A's original UUID; password, role, ownership and grants unchanged | OIDC handler/repository/browser regressions; live UNVERIFIED |
| 3. Provider sign-in | A authenticates through standard discovery/code + S256 PKCE; no provider token is persisted, returned or placed in URLs; current A UUID remains the actor | Synthetic HTTPS issuer/token/HTTP tests; live UNVERIFIED |
| 4. Reject identity and callback conflicts | Unknown subject, same-email unlinked D, an issuer/subject already linked to A, missing/wrong/replayed state or browser proof, invalid nonce/issuer/audience/signature/expiry all fail without account creation or linking | Auth/handler/multi-user tests; live UNVERIFIED; use controlled staging callbacks only |
| 5. Check authorization through both login methods | B sees only active granted sources; VIEW cannot mutate; EDIT applies only to eligible locally managed records; imported/external records remain read-only; C has no implicit clinical access; D cannot read A | Delegation + OIDC HTTP and import regressions; live UNVERIFIED |
| 6. Revoke and expire grants | Reusing an earlier search/detail identifier after expiration/revocation/source deletion denies access; authentication does not retain grant snapshots | Source access/search database and HTTP regressions; live UNVERIFIED |
| 7. Logout, unlink, expiry, username replacement | OIDC logout rejects the prior native session; unlink rejects all old and in-flight sessions; fresh re-link never revives them; expired sessions fail; deleted/recreated native names cannot inherit sessions or grants | OIDC generation/session tests and native replacement HTTP/DB regressions; live UNVERIFIED |
| 8. Provider failure and disablement | Bounded failed login does not block startup, password login or key recovery; unknown/disabled IdP identity cannot start a fresh login; an existing session follows the documented expiry window | Synthetic provider-failure tests; live UNVERIFIED; no claim of back-channel revocation |
| 9. Browser and desktop recovery | Native browser login remains usable after provider failure or OIDC disable/restart; desktop native login/unlock works independently; no desktop OIDC button or embedded IdP login | Browser contracts, embedded tests and required Desktop CI; live UNVERIFIED |
| 10. Approved rollback | Disable OIDC, restart, reject OIDC sessions and recover locally; retain additive audit/identity tables and all clinical data | Configuration/session regressions and documented rollback; live UNVERIFIED |

## Receipt, stop conditions and rollback

For each row retain PASS / FAIL / UNVERIFIED, UTC time, exact revision/artifact,
synthetic actor aliases/native UUIDs, source/grant/session audit IDs, expected and
actual HTTP status and sanitized screenshots if useful. Never retain passwords,
secrets, signed tokens, cookies, authorization codes, raw subjects or medical text.
Do not include callback query strings in access logs or shared reports.

A failed identity/authorization check blocks live acceptance. An unavailable provider
blocks provider acceptance only; native recovery must still pass. Missing deployment
access or an unapproved action leaves that row UNVERIFIED, without altering accounts.
A successful CodeQL/build does not establish live login or recovery acceptance.

Rollback follows the existing OIDC document: disable OIDC and restart only under the
approved operational allowlist, use the verified native account/key, retain tables
and provenance. Do not rotate signing/encryption keys, reset users, drop tables or
rewrite grants as a workaround. Reverting UUID binding restores the username-reuse
session defect; retain the fix during an OIDC-only rollback.
