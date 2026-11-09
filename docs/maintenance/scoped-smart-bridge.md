# Local-identity-scoped native SMART bridge

Pi owns `task/scoped-smart-bridge`, initially stacked on the owner-confirmed API. The bridge now requires
both an authorized main window and a backend-authenticated local identity for Patient/clinical reads.
Provider credentials remain in Go. No new provider profile, application TCP listener, external identity
service or clinical-import UI is introduced by this slice.

## Identity is resolved, not asserted

The native host invokes its **existing embedded handler in process** at the fixed authenticated
`/api/secure/account/me` path, supplying the application's current local JWT in a header. The address is
not renderer-configurable and no socket/request URL credential is used. The same JWT, immutable-user,
access-token and OIDC-session checks apply as to normal API requests; there is no second JWT verifier.

Lookup requests have a ten-second context bound, bounded input token and a 64 KiB response recorder.
Only UUID and username leave the resolver; profile fields and diagnostic response bodies are not returned
or logged. Requests/errors do not contain provider tokens. The local application JWT already belongs to
the renderer's API session; it must not be confused with the never-renderer-visible provider credentials.

Each connection is bound to its resolved local UUID in the Go service. Identity is checked before browser
connection and again at completion; a changed/revoked identity releases the new provider connection without
returning it. Patient/clinical retrieval checks the connection owner before I/O and rechecks after I/O.
A known connection ID, reused username or a main-window call alone is insufficient to return data.
Provider/auth windows are rejected before resolver/manager invocation.

`Cancel`/`Disconnect` intentionally remain main-window cleanup operations without a valid local JWT,
so logout/expiry cannot prevent credential release. They cannot open a session or return clinical data.
Disconnect removes the local ownership binding. App shutdown continues to clear the provider manager.

Matched native/bridge signatures now take a local application token for connection and reads. Returned
connection DTOs include only public provider/Patient metadata and a verified local owner, not credentials.
The Angular service checks the expected local UUID and discards late/cancelled/identity-changed previews.
Unsigned browser claim decoding is only a UI stale-context check, never native authorization.

The clinical reader can now be called through this scoped bridge, but the clinical UI action is not yet
implemented. Existing Patient-only operation still works with a current local login. Receiving-profile
confirmation, API wiring and stale-context protection across the legacy import *write* remain the next UI
slice; this bridge milestone is not whole-household consent or import acceptance.

## Active-account regression

While qualifying the shared authority boundary, a focused synthetic test proved that a correctly signed,
UUID-bound token for a soft-deleted account still reached a protected handler (204 instead of expected
401). The unchanged positive control reached 204. `RequireAuth` now rejects `user.DeletedAt != nil` before
setting authenticated context. This is a narrow live-account validity correction, not a new token format,
key rotation, altered OIDC identity linkage or change to stored user records.

## Evidence and limitations

- Real embedded encrypted backend: sign up synthetic principal, resolve real JWT, reject bad signature,
  delete it, recreate the same username, accept the new UUID and reject the prior token. No production
  identity, provider or health record was touched.
- Service tests: foreign/local/window separation, denial before manager data access, successful owner reads,
  session change during connection/read, zero returned payload after revalidation failure and cleanup after expiry.
- Resolver tests: fixed method/path/header, context bounds, malformed/oversized/failed responses and constant
  private-canary-free errors. Real provider protocol/race tests remain in the underlying manager suite.
- Active-account regression failed before the one-predicate correction; complete middleware race suite passed
  afterward. The new bridge/fixtures were developed together, not represented as test-first product regressions.
- All desktop internal packages passed under race (host 25.974s, shell 1.162s, SMART 6.257s, fixture 1.205s).
  Native linking still warns about newer macOS object targets versus 11.0; older-macOS support is not qualified.
- Desktop Node contract suite passed; focused Angular tests passed 15/15. Full Angular run passed 724 with
  one skip and retained existing unknown `fa-icon` test-template console diagnostics. This is not silence-free
  native GUI acceptance or an independent security certificate.
- Initial native test invocation requested module metadata. Only existing uuid v1.6.0 and the root's cached
  testify v1.12.1 were explicitly declared; no versions/catalog were refreshed. A read-only tidy diff suggested
  unrelated pruning, which was not applied. `-mod=readonly` native race tests then passed.

## Final parent integration and request lifetime

The accepted API parent is #473 at `3fa5d8e993d1f4bb0d8d2b212eae4690a78a2e8d`. All seven checks
qualified corrected API head `14eded02f5ebc5518cccbb0963a6c5a37f9926aa` (run 42568249499); protected
matching-head merge succeeded only after both reproduced review findings were resolved. Exact-main
validation 42570752698 also passed. No protection bypass or old-head qualification was used.

The native resolver's request context must actually reach account/authentication SQL. Passing raw
`*gin.Context` loses request cancellation/deadlines under the default adapter. A synthetic application
boundary check reproduced the missing deadline. A trial global `ContextWithFallback` switch was rejected:
the unchanged encrypted native identity race test found SQL's asynchronous Rows cleanup reading a pooled
Gin Request while the engine replaced it on the next call. The global change was removed, not shipped.
Authentication lookups now use the stable request context directly. The authenticated account handler
uses a small adapter which copies only verified UUID/username onto that context, masks inherited values
when identity is absent, and never retains Gin. A bounded request-lifetime/identity snapshot test passes
under race (2.062s), including reuse and missing-identity controls.

Final native internal race passed: host 23.309s, shell 1.182s, SMART 6.430s, fixture 1.223s; native vet
passed. Full non-race web/handler/middleware compatibility passed (68.030s / 81.921s / 0.529s). Desktop
Node contracts and the full Angular run passed (724 tests, one skip), retaining existing warnings.
Attempts at broader web race suites exceeded 300/600-second tool budgets and a 120-second diagnostic
budget; none is counted as a pass. The diagnostic also exposed **legacy** manual-import pooled-context
races. A compiler overlay of the exact accepted main parent independently reproduces them (56.714s),
without the native adapter changes. [Issue #474](https://github.com/chartrove/chartrove/issues/474) owns
that remaining application-wide repair; the scoped account adaptation does not claim to fix it.

These methods authenticate the supplied local JWT through the real backend; they do not introduce
server-side revocation for ordinary logout or attest which browser cookie is currently selected.
Renderer stale-profile checks and backend account validity remain distinct defenses. No native GUI,
whole-household, PostgreSQL or independent operator/key-loss qualification is inferred from these tests.
Final source review and exact-head hosted validation remain required for this bridge slice.
