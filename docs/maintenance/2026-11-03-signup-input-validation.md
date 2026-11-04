# Malformed sign-up request handling

This bounded defect was found during the retained authentication-feature admission
review. It is independent of historical #217's unanswered reserved-name policy.
UserWizard embeds a pointer to User. An empty object, JSON null or only wizard
fields binds without allocating that user, then the old handler dereferences it
when assigning a role. Malformed JSON returns an internal error; absent/blank
credentials can reach account creation rather than a client-input rejection.

The public sign-up handler now returns HTTP400 with a generic message for invalid
JSON/types or missing/blank credentials, before account lookup or creation. The nil check short-circuits embedded-pointer access. TrimSpace is used
only to identify blank inputs; nonblank username/password bytes are passed through
unchanged. Existing maintainer name admin stays available and roles still follow
actual first/subsequent-account policy, never the submitted role or username.
This does not add a denylist, normalize identifiers, change an existing account,
rotate keys, modify sign-in, change schemas/dependencies or implement #217.
Administrative user creation and broader account-policy work are separate scope.

The same14ordinary HTTP cases run against actual Gin JSON binding/recovery and
recording repository mocks: old two-pass/twelve-fail, corrected14-pass, with
identical spec bytes. Assertions require HTTP400, no panic-recovery output,
zero account count/creation calls and no token on invalid input. Valid admin and
literal spaced credentials retain their exact values and subsequent-user role,
including rejection of a submitted admin role. The raw old source is preserved.
The earlier existing-file overlay discovered no cases and fails its14-case
admission; it is retained but excluded from usable proof. Ordinary source-file
controls provide the evidence.

Focused corrected root checks pass102 test/subtest events across config, auth,
handlers and web, including the prior lifetime, JWT wire, sign-in denial and
actual synthetic encrypted unlock/signup/reopen/expired-token fixtures. The focused desktop graph
also passes all102events and all14identical HTTP cases. The existing120-second
embedded limit remains unchanged. No real medical records, purchased-app data or
production provider is accessed; private fixture scratch is disposed.

No full local frontend/native build, cache/backup cleanup or50GiB admission
bypass occurs. Angular cache remains false and both retention owner branches/
worktrees remain held. Existing protected Backend/Frontend/Desktop checks are
required before integration. No issue is falsely closed for this newly found
request-handling correction. Raw evidence:
dist/validation/2026-11-01-consolidation/post-dependency-issue-review/
signup-input-review/. [Compact validation](signup-input-validation.json) records
the completed root/desktop checks.

## Protected integration

PR #406 merged at3e6e9a73eb50bc43f7e67988d42fbed254c3ed91 with all three
required checks passing in run41750895630:572 browser cases, one existing skip,
16Node checks and6.13MB initial output under unchanged10MB. All complete job logs
and hashes remain. Independent main-push checks also pass. No historical issue
is closed for this correction; #217 remains open with its naming policy pending.
Only the incorporated task branch was retired, its exact tip remains on main,
and both independent retention owners remain held.
[CI](signup-input-ci-validation.json),
[retirement](signup-input-retirement.json),
[fresh checkpoint](2026-11-03-checkpoint-after406.json).
