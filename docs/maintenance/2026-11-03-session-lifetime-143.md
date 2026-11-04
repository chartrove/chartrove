# Installation-configurable session lifetime (#143)

Issue [#143](https://github.com/karamhouse/chartrove/issues/143), imported from
Fasten #451, requests configurable automatic logout and reports ten minutes.
The complete description has no further discussion. Current main issues standard
user tokens with a fixed one-hour expiry. This batch implements configuration of
that absolute lifetime; it does not establish the historical ten-minute cause.

## Behavior and compatibility

`jwt.session.duration` / `FASTEN_JWT_SESSION_DURATION` controls newly issued
password sign-in/sign-up tokens. Missing/empty settings retain one hour. Durations
use Go units, for example 5m or2h, and must be at least one second. Zero, negative,
subsecond, malformed and overflowing settings are rejected. Standalone and
embedded startup validate before storage initialization. Direct sign-up also
validates before account creation; direct sign-in validates after credentials.
Configuration errors return a generic response without tokens or raw settings.

The existing JwtGenerateFastenTokenFromUser signature/default is preserved.
A separate lifetime-aware generator changes only expiry. HS256 signing, issuer,
subject, role, metadata, token ID/type and validation remain unchanged. Existing
signed expiry stays authoritative after a restart or setting change. No key
rotation, record rewrite, schema/module change, browser token-contract change,
per-user preference, idle timer, sliding refresh or revocation mechanism is added.
The shared backend path serves standalone and desktop; no provider/IdP is enabled.
Operator guidance and the example config document the setting and its limits.

## Validation and retained diagnostics

The initial virtual-file Go overlay did not discover the added test file and
reported no tests to run. Its exit0 is retained but excluded from consumer proof.
Ordinary package tests on the owned task branch then execute all seven sign-in
cases: old one-pass/six-fail, corrected seven-pass, with identical spec bytes.
Assertions verify actual signed five-minute/two-hour/default tokens and rejection
of invalid settings, preserving issuer/subject/role/type/ID.

Focused ordinary root tests pass87 test/subtest events across config, auth,
handlers and web. They include file/environment precedence, invalid startup before
storage, no account creation on invalid settings, existing sign-in denial and JWT
wire/signature cases. Actual embedded synthetic encrypted startup/unlock/signup,
close/reopen under a changed lifetime, old-token authenticated account access,
new-token expiry and expired-token denial pass. The encrypted file header remains
non-plaintext and the same synthetic account/password survives. Existing encrypted
wrong-key/unlock/shutdown fixtures also pass. No real records or purchased Fasten
paths are opened. The same focused desktop-graph invocation also passes all87 events and all
seven identical HTTP cases, with encrypted reopen and expired-token denial.

The unchanged120-second embedded limit remains. Private scratch is disposed with
logs/receipts retained. No dependency refresh, cleanup, full local frontend/native
build or50GiB admission override occurs. Cache remains false and both independent
retention owners stay held. A new packaged native UI is not claimed.

Protected Backend/Frontend/Desktop integration and individual issue disposition
remain pending. Raw evidence: dist/validation/2026-11-01-consolidation/
post-dependency-issue-review/issue143/. [Compact validation](session-lifetime-143-validation.json) records the completed
root/desktop consumer checks. The full upstream report and failed overlay remain
preserved alongside ordinary controls.
