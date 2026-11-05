# Operator-controlled public registration

Chartrove preserves its existing public registration behavior by default. Operators
can disable subsequent public self-registration in the existing configuration:

```yaml
auth:
  registration:
    enabled: true
```

Set `enabled: false`, or `FASTEN_AUTH_REGISTRATION_ENABLED=false`, and restart with
the reviewed configuration to close public signup after initial setup. There is
no browser/API endpoint for changing this operator setting. No live configuration
or existing account is changed by this source feature. An invalid nonempty value
is rejected rather than interpreted as permission to register.

This implements [#417](https://github.com/karamhouse/chartrove/issues/417), preserving
[joestump's original request](https://github.com/fastenhealth/fasten-onprem/issues/47#issuecomment-1845448106)
from the attributed discussion in #16 / upstream Fasten #47. OIDC authentication
and native registration remain separate policies.

## Bootstrap, recovery and authorization

- An empty database can still bootstrap its first local administrator, even when
  subsequent registration is disabled. Protect the initial setup endpoint with
  the deployment's existing network/ingress controls until that setup is complete.
  The empty-database exception is intentional; disabling public registration is
  not a replacement for securing an uninitialized installation.
- The database transaction serializes the user-count, policy and role decision
  before initializing the user, settings and local source. SQLite reserves the
  writer before reading; PostgreSQL uses an application-specific transaction-level
  advisory lock with an explicit READ COMMITTED transaction so the count observes
  the previous lock holder's commit even under a stricter server default.
  Both operate across connections/processes. Concurrent setup can
  create exactly one administrator. With registration enabled, later successful
  requests create ordinary users regardless of client-supplied role; with it
  disabled, later requests fail. Initialization errors roll back the entire account.
  No bootstrap-lock table or schema migration is added.
- Existing local password sign-in, encryption-key restore, administrator-managed
  user creation and existing sessions/API tokens keep their existing contracts.
  The flag does not disable, reset, migrate or revoke existing accounts. An
  administrator can continue creating native users through the authenticated user
  management API; ordinary users cannot. The existing forgotten-password screen
  is not a newly implemented password reset. Operators still need independent
  local administrator credentials and database/encryption-key recovery ownership.
- OIDC linking still requires an existing native account, fresh local password
  proof and provider authentication. Unknown provider subjects are not provisioned;
  email matching and provider roles/groups never create or elevate native users.
- Native administration remains distinct from medical-record access. Owners and
  active explicit per-source VIEW/EDIT grants remain authoritative. Public signup,
  managed creation and OIDC authentication grant no cross-owner clinical access.

## Browser and desktop behavior

`GET /api/auth/registration` exposes only three availability booleans: `enabled`,
`allowed` and `bootstrap`, with `Cache-Control: no-store`. It does not return users,
credentials, clinical data, configuration secrets or a registration capability.
Both signup forms fail closed while that status is unavailable. Existing-account
sign-in navigation stays accessible. The administrator-setup wizard uses the same
policy and endpoint through the existing embedded desktop HTTP transport.

`POST /api/auth/signup` repeats the real decision transactionally; a stale page or
client-supplied availability/role cannot authorize creation. A closed-policy reply
is HTTP 403 with `code: registration_closed`. The browser handles that specific
public-signup denial without logging out an existing native session. Protected
401/403 handling remains intact. A policy change during submission clears entered
passwords and closes the form; an unconfirmed response never installs a token.

Encrypted/locked startup still exposes its existing key-recovery routes. Registration
routes are published only after a successful unlock. Tests exercise bootstrap,
restart, wrong/correct key restore and local login with public registration disabled.
An IdP outage cannot disable these local operations.

## Validation and rollback

Synthetic regressions cover default/env/invalid configuration; real authenticated
HTTP bootstrap, denied public creation, managed creation and private-record isolation;
two-repository encrypted SQLite and disposable PostgreSQL concurrency; initialization
rollback/cancellation; both browser forms and the exact interceptor exception;
encrypted embedded recovery; and existing OIDC linking/revocation with registration
disabled. Required Backend, Frontend, Desktop and CodeQL checks remain unchanged.

Rollback: restore `enabled: true` and restart, or revert this focused source change.
Existing users, passwords, UUIDs, records, grants, keys and audit history are retained.
No production database operation or identity-provider change is part of rollback.
