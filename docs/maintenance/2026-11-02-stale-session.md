# Stale-session rejection - issue 103

The earlier signin fix in #323 did not cover a valid session for a deleted
account. The protected-route middleware now verifies the current account before
calling handlers, returning 401 for absent/nil accounts and 500 for genuine account
or access-token database failures. Access tokens must still exist in storage.
Require the Bearer scheme and stop logging malformed authorization headers.
Token format, expiry, signing algorithm and metadata are unchanged.

Ten account/access-token boundary cases cover valid users, missing/wrapped/nil
accounts, revoked/nil access tokens and database failures. Malformed-header tests
also reproduce credential logging and acceptance of a valid JWT with the wrong
scheme before the fix. An end-to-end embedded test creates and deletes only a
synthetic account in a fresh temporary database, then reproduces 500 from the
stale session before repair and verifies 401 for account/source requests afterward.

Targeted regressions pass. Full root web/handler/middleware/auth tests and web vet
pass. Desktop-graph encrypted embedded tests (including the deleted-account case)
and middleware tests pass under unchanged timeouts. This does not change token
identity or provide atomic account-deletion semantics across concurrent requests.

## Preservation checkpoint

Local Karma runs wrote reports into an existing ignored frontend/coverage tree.
The pre-existing primary index birth time is October 26; the verified separate
coffinfish coverage archive is a different snapshot. Existing primary reports
should have been preserved before running Karma. The current mixed tree is now
retained at dist/validation/2026-11-01-consolidation/coverage-after-maintenance,
and further local coverage output is redirected to the private task directory.
No original primary coverage preservation claim is made until recovery completes.

A pre-task local snapshot exists: com.apple.TimeMachine.2026-11-01-183642.local.
macOS denied backup inventory access because Full Disk Access is required.
Recovery is limited to the original frontend/coverage directory; user action or
permission is needed to read that snapshot. The original handoff archive remains
hash-identical. This gap affects generated coverage evidence, not Git history or
application storage; do not claim all acceptance criteria achieved while unresolved.
