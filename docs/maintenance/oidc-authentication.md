# Optional OIDC authentication

Chartrove can authenticate existing native accounts through one configured standard
OIDC provider. It remains disabled by default. This is a replacement for historical
[Chartrove #214](https://github.com/karamhouse/chartrove/pull/214), imported from
[Fasten #613](https://github.com/fastenhealth/fasten-onprem/pull/613), proposed by
[TheoAtTechStack](https://github.com/TheoAtTechStack) with commits by Theo
(`Theo@tsa.team`, GitHub `alexszilagyi`). The original head
`bcb60fcb9feca0f683ab7b39fd221b4c274d6883`, original comparison base and discussions
remain preserved. Its discovery/library and configurable provider concepts inform
this implementation; its unsafe email linking, automatic administrator creation,
generated replacement passwords and app-token query redirects are not adopted.
Its dependency and container updates are superseded by current main.

## Security and identity model

- The native user UUID remains the owner of records, permissions and API tokens.
  A verified exact issuer/subject pair binds to exactly one UUID. Matching email,
  username, display name, provider groups or roles never creates or links an account.
  Unknown subjects are denied. There is no just-in-time provisioning or role sync.
- Sign in locally first, open Settings → identity provider sign-in, and re-enter
  the local password to link. The server binds a five-minute transaction to the
  authenticated native UUID and current password hash. Another account cannot
  take an existing issuer/subject binding. Explicit re-linking to its original
  account requires fresh local password proof and IdP authentication.
- The server uses `coreos/go-oidc/v3` and `golang.org/x/oauth2` for discovery,
  authorization-code exchange, S256 PKCE and signed ID-token validation. It checks
  exact issuer, audience, authorized party, expiry, issued-at and nonce. Only
  asymmetric signature algorithms are accepted. State and independent browser
  proof bind the flow to the initiating tab. Both the browser and server consume
  it once. Discovery, token and JWKS endpoints and the callback require HTTPS.
- The browser stores transaction proof in sessionStorage for five minutes.
  The IdP redirects only a code/state to the configured frontend callback. The
  frontend removes the query before completion and submits the code/proof to the
  server; it never processes provider tokens. No native app token, password,
  provider secret or browser proof enters a redirect URL. The document sets
  `Referrer-Policy` through an early `no-referrer` meta element. Reverse proxies
  must redact callback query strings in their access logs.
- Provider tokens are neither persisted nor returned to the browser. The new
  native JWT uses the existing application contract with `token_type: oidc` and
  a random session ID. Its expiry is bounded by the configured native session
  lifetime and verified ID-token expiry. No refresh/offline access is requested.
- Every protected request checks that session's native UUID, active identity,
  expiry and revocation in the database. Deleted or recreated usernames cannot
  inherit a binding. Disabling OIDC or changing its issuer rejects OIDC sessions
  while leaving local authentication intact. Logout revokes the current native
  session. Password-confirmed unlink revokes the binding and all its sessions.
  A pending link started before a later revocation cannot undo it. Fresh
  re-linking never revives old sessions. The latest 200 account-scoped audit
  events retain link/login/logout/unlink identifiers and timestamps, without
  clinical payloads, passwords, subjects or provider tokens in API responses.
- OIDC changes authentication only. Existing owner predicates, per-source
  VIEW/EDIT grants, eligible local-record provenance, expiration, revocation and
  source-access audit remain authoritative. Administrators gain no patient-data
  access. Existing independently issued API tokens retain their existing separate
  management and revocation semantics; unlinking does not revoke local credentials
  or those API tokens.

## Configuration and Authentik compatibility

Configure these existing YAML/environment settings only after operational review:

```yaml
auth:
  oidc:
    enabled: false
    issuer: https://identity.example/application/o/chartrove/
    client_id: chartrove
    redirect_uri: https://chartrove.example/web/auth/oidc/callback
```

Supply `FASTEN_AUTH_OIDC_CLIENT_SECRET` from the deployment secret store for a
confidential client; never commit a real secret. Standard PKCE public clients are
also supported. Configure the exact redirect URI with query response mode in the
provider. It must resolve to this Chartrove frontend in the same browser tab;
include any reverse-proxy subpath. The server neither accepts arbitrary redirect
destinations nor derives one from request headers. Use a stable, provider-specific
subject and issuer. Do not change subject mappings to match an email address.

Authentik provides standard discovery and PKCE:
[provider documentation](https://docs.goauthentik.io/add-secure-apps/providers/oauth2/).
Use that application's discovery issuer as configured; no particular deployment,
Authentik groups or tenant URLs are hardcoded. Only `openid` scope is requested.
Provider outages fail the new sign-in attempt with a generic error. Discovery is
lazy and bounded, so provider availability never gates Chartrove startup, local
password sign-in, database unlock or clinical operations with valid sessions.

This milestone supports the web application. Desktop retains its existing local
authentication and encryption-key recovery. The desktop hash router does not
advertise OIDC. A future native-browser callback requires its own reviewed flow;
it must not reuse web redirects or embed provider login in an untrusted webview.

## Revocation, recovery and rollback

Chartrove logout terminates its native OIDC session; it does not log out the IdP.
Provider-side account disablement/logout is observed on the next provider login
and at native session expiry. Back-channel logout, introspection, provider session
termination and refresh are not implemented. Choose short IdP ID-token and native
session lifetimes for the required revocation window. To revoke immediately in
Chartrove, sign in locally and revoke the identity, or disable OIDC and restart.
An offline browser logout clears the device token but cannot confirm server-side
revocation; unlink from another locally authenticated session or await expiry.

No local password is replaced or removed. Before enabling OIDC in any deployment,
the operator must prove an independent local administrator password login and
encryption-key restore, identify the credential recovery owner, review callback
routing/log redaction and prove a synthetic provider login and revocation. The
existing forgotten-password UI is not a password-reset implementation. This code
change does not provision, rotate or migrate administrator credentials.

Rollback: set `auth.oidc.enabled: false`, restart, and use local authentication.
All OIDC sessions are denied on every request. The additive identity/session/audit
tables can remain for history; do not drop tables or rewrite users/clinical rows.
Reverting the source commit restores the original local authentication paths;
keep OIDC disabled during rollback. Pending transactions are lost on restart,
which fails closed. Transactions are process-local; a multi-replica deployment
needs sticky routing for the five-minute start/complete exchange. A shared
transaction service is outside this single-server milestone.

## Acceptance boundaries

| Check | Evidence/status |
| --- | --- |
| Local login and recovery with provider offline | Synthetic HTTP regression; native passwords unchanged |
| OIDC discovery, code exchange, PKCE, nonce and token checks | Synthetic HTTPS issuer/signing-key regressions |
| Duplicate email, conflicting bindings, roles and immutable UUIDs | Repository and HTTP multi-user regressions |
| Session expiry, logout, unlink and re-link | Database/session middleware regressions |
| Source access and private-record isolation | Existing delegation suite plus actual OIDC HTTP regression |
| Browser state, duplicate/invalid callbacks and settings controls | Focused Angular browser regressions |
| Desktop local authentication/encryption | Existing embedded tests and required Desktop CI |
| Protected Backend, Frontend, Desktop, CodeQL and independent review | Required on replacement PR; do not infer from local tests |
| Live Authentik login, operator recovery and proxy configuration | UNVERIFIED / separate operational approval required |

No production data, live user inventory, Authentik configuration or secrets were
accessed by the synthetic acceptance work. OCR and authorization-aware retrieval
are separate feature assessments after this authentication foundation is accepted.
