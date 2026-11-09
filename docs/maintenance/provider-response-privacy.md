# Dynamic-client provider response privacy

The readiness review followed provider credential handling beyond the create/reconnect handler corrected
in [PR #470](https://github.com/chartrove/chartrove/pull/470). Both legacy dynamic-client registration
and token refresh read the entire rejected HTTP response body and print it using the standard logger.
Provider errors are untrusted: they may echo authorization details or patient information, and the body
has no length bound. This observation does not assert that a deployed provider actually returned secrets.

## Focused correction

For non-2xx responses, close without reading or logging the body and retain the existing status-based
error. Successful registration, key generation, assertion signing, token decoding and expiry handling are
unchanged. There is no provider catalog, registration, identity, credential or database migration.
This does not qualify success-response limits, transport deadlines, redirects or other SDK/log paths.

## Regression evidence — synthetic only

`TestDynamicClientResponseBodyPrivacy` invokes both real model methods, including embedded-catalog
resolution and synthetic RSA/JWT handling. A replacement `http.DefaultClient` transport intercepts **every**
request without a socket, returning controlled 200/400/401/500 replies. It is restored after each subtest,
as is the captured standard logger. Catalog metadata comes only from the existing pinned SDK; there is
no live discovery, provider account or production authentication.

- The six rejected-response cases fail before correction because the reply marker enters standard logs.
- Unchanged assertions pass after correction: replies are closed, no rejected body is read, markers are
  absent from logs/errors, HTTP status diagnostics remain, and failed operations preserve the existing
  client ID/access token/expiry.
- Positive 200 controls prove real registration and token decoding still update credentials. Token-request
  assertions also prove synthetic signing actually executed, rather than an earlier failure satisfying the test.
- Full `go test -race ./backend/pkg/models -count=1 -timeout=120s` passes (33.883s), including existing
  OAuth refresh, serialized JWK/assertion and model compatibility tests. Dependency downloads were disabled.
- Full `go test -p 1 ./backend/... -count=1 -timeout=15m` passes on source checkpoint
  `f329ceba66` (exit 0). Database 258.521s, web 101.154s, handler 65.263s; all other tested
  packages also passed. This completes the earlier baseline whose tool/test budgets were too short.
- `git diff --check` passes. Full protected CI must qualify the final head before integration.

Exact failing/passing logs are retained in the private readiness evidence, not published as medical data.
This is focused project regression validation, not an independent security-audit certificate. No historic
logs were inspected or deleted, no live credential rotation was attempted, and no production system was changed.
