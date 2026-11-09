# Stable legacy repository contexts

Pi owns issue [#474](https://github.com/chartrove/chartrove/issues/474), branch
`task/request-context-isolation`, in the canonical single checkout. This first bounded repair addresses
the observed manual/related import, source-access transaction, account deletion and OIDC SQL producers.
It is not a blanket audit or certification of every remaining legacy caller.

## Failure and boundary

A read-only compiler overlay of accepted main `3fa5d8e993d1f4bb0d8d2b212eae4690a78a2e8d` reproduced
race failures in unchanged encrypted native-source/graph tests (56.714s). Manual SDK writes retained
pooled `*gin.Context` in SQL transactions. `database/sql.Rows.awaitDone` may inspect that context after
the HTTP handler returns; Gin can already be replacing its Request for another call. Disabling Gin's
context fallback does not eliminate the read inside `hasRequestContext`. A global fallback trial was
previously rejected during native bridge qualification, not shipped.

The manual importer now passes an immutable, verified UUID/username job context through identity
checks, source creation, SDK parsing/sync, background-job updates and completion lookup. Related local
resource creation uses the same policy, and its source helper takes a plain `context.Context`. This
preserves the existing tracked-job lifetime: HTTP cancellation does not silently acquire a new job
cancellation/rollback contract. It does not make legacy SDK collection imports atomic or bounded like
the separate clinical-snapshot API. Local JWTs are not copied into jobs.

A full web race diagnostic then found the same retention class in source grants/revocation, related
resource creation and account deletion (607.036s failing run). The relevant request-bound operations
now use the stable request context with copied server-authenticated UUID/username from the native
bridge adapter. Missing identity still becomes nil UUID/empty username, rather than an inherited or
reusable name-only principal. Source-access bodies, pagination, authorization, audit and error policies
are unchanged. OIDC SQL completion/revocation also uses the stable request context; those repository
methods already receive explicit verified user/session identifiers. No issuer, linking, signing key,
token format or production identity was changed.

## Validation and rejected shortcuts

- Unchanged encrypted source/graph regressions pass under race (57.142s); the synthetic import/graph
  process-output privacy control passes (27.848s).
- An explicit job context control cancels and reuses the original HTTP context, then verifies the original
  UUID/username and job ID remain stable, no local JWT is retained, and the old detached lifetime remains.
  Existing delete/recreate-username tests still refuse old-account writes and missing immutable identity.
- Complete web race suite passes (639.873s), including source isolation, exact edits, encrypted backup,
  grants/revocation, deletion, registration and local identity replacement. Earlier 120/300/600-second
  attempts were incomplete, not passes or evidence of semantic deadlock; bcrypt cost-14 race fixtures
  require the longer budget.
- An initial full handler run failed (323.138s): it exposed remaining OIDC transaction contexts, plus old
  direct-handler fixtures that supplied only usernames or no HTTP request. The fixtures now supply the
  actual created account UUID and a real synthetic request, like production middleware; no expected
  assertions or missing-identity denial controls were weakened. Ordinary handler compatibility passed
  (82.491s). Complete handler/middleware race then passed (550.420s / 1.703s).
- All desktop internal race packages pass under the desktop module graph: host 26.150s, shell 1.329s,
  SMART 6.821s, fixture 1.368s. Existing newer-object/older-link-target macOS warnings remain visible.

No timeout suppression, race suppression, external provider, real health data, global Gin fallback,
core-database framework coupling, credential export or dependency/catalog refresh was used. Raw logs,
overlay sources and synthetic runtime evidence remain private. The matched clinical UI inspection
package predates this repair; it is not silently requalified by these source tests. Other raw legacy
web-to-repository contexts remain to be assessed under #474, and production/native GUI/independent
operator qualification remains separate.

## Request-bound follow-up

The accepted first repair is #476 (`53578aeb341c01fef92a580d6eab39062cc1043b`, exact-main validation
42584452848 passed). Pi continues the same issue/branch/checkout with registration/login/status,
source CRUD/read, generic resource/owned-mutation and reviewed-document boundaries. Anonymous auth
operations use the original HTTP context; authenticated repository calls use the verified UUID/username
snapshot. Source-sync completion deliberately retains its existing immutable detached job context rather
than binding already detached work to request cancellation. No password cost or registration budget changed.

A new four-case real-handler boundary probe deterministically checks source listing, deletion, composition
and registration. Each originally passed `*gin.Context` to the repository. An initial test-only compile
mistake (`require.NotIsType` is unavailable) was fixed before all four genuinely failed; it is not counted
as a behavioral regression. Unchanged controls then pass, including request deadline/cancellation after
simulated context reuse, immutable actor values and absence of local JWT in the repository context.
The combined focused race/old job-authority controls passed (23.069s).

Initial ordinary handler compatibility failed only because two old direct-resource tests provided no HTTP
request; their shared fixture now provides a real synthetic request without changing assertions or the
production adapter. Complete handler/middleware race passes (576.873s / 1.642s), and complete web race
passes (652.820s). Other remaining legacy contexts, such as practitioner/settings/token/summary/diagnostic
callers, are not certified by this bounded follow-up. No global Gin fallback or core-database coupling was
introduced; exact original/provenance/authorization and detached-job controls remain in the normal suites.
After normally integrating accepted clinical UI #477, the complete desktop internal race suites also pass
serialized: host 23.460s, composed encrypted shell flow 27.289s, SMART 5.982s, fixture 1.230s. The ordinary
macOS linker warnings remain visible; this does not qualify a fresh native GUI/package or older OS.

These backend race receipts precede the normal accepted-#477 parent merge; the final desktop graph
receipt follows it. All seven hosted checks qualified exact follow-up head
`f094d12f3fbe230f320306191a5131c710682a84` (42592715042), with no unresolved review threads.
Protected matching-head #478 merged at `dc48128ad8691513ae9c2e8607567482849f6c8f`;
exact-main validation 42594974931 passed. #474 remains open for the unassessed callers above.

## Clinical routes and other request-bound producers

The next actual-callee review found the clinical API itself still passed Gin to its actor/import
repository calls and used Gin as the parent of history/original timers. A timer's different dynamic
type does not make its retained parent safe or connect it to the HTTP deadline/cancellation.

New tests use the actual registered clinical routes, a boundary-only fake repository and synthetic
server-identity middleware (not JWT verification). All three
route classes genuinely failed their original-request deadline controls; a blocked import also
ignored a real HTTP abort (3.111s failing package result). After replacing the Gin object, the tests
model pool reuse and verify the original immutable UUID/username, earlier HTTP deadline, cancellation
and absence of local JWT, including timed history/original children. The owned blocked-handler test
returns the fixed HTTP 408 response on abort and releases its synthetic goroutine on red/green cleanup.
Existing real-JWT/encrypted clinical HTTP tests remain the storage/authority controls, not replaced by
these fakes.

Seven additional real-handler boundary controls genuinely failed for summary, dashboard settings,
access-token listing, practitioner favorites, job listing, synchronous client-failure recording and
admin user listing (1.037s). The original four controls stayed green. The unchanged expanded boundary
and clinical-abort tests then passed under race (2.083s).

Reviewed authenticated practitioner creation/update/favorites/history, settings, summary/IPS, token,
admin-user and SSE actor calls now use the stable authenticated request snapshot. Public health and
global glossary-cache SQL use the original HTTP context; no owner is inferred for that shared cache.
The client-error job route synchronously records an already-failed job; it is not an already detached
worker. Existing detached manual/related/source-sync jobs and their deliberate lifetimes are unchanged.
Remote glossary/dashboard fetching, filesystem work, error/log policy, provider calls and new atomicity
or cancellation contracts for those operations are not qualified by these repository-boundary changes.

Ordinary complete handler/middleware compatibility passed (83.302s / 0.516s) without fixture changes.
Complete pre-parent web race passed (611.537s). After normally integrating accepted PostgreSQL #480 at
code checkpoint `f5b26a39d5c7268bf99f3f0a72d853faff8beaf0`, fresh complete web/handler/middleware
race suites passed (623.159s / 545.789s / 1.831s). Serialized complete desktop internal race suites
passed: host 23.184s, composed encrypted shell flow 26.608s, SMART 6.020s, fixture 1.223s. Existing
macOS object/link-target warnings remain. Changed-web vet/diff checks passed. No global fallback,
deadline/hash relaxation, new dependency, real provider, production identity or record was used.

The remaining direct Gin-to-SQL/SDK matches in `unsafe.go` are deliberately **not** relabeled as
verified authenticated requests: those opt-in diagnostics select a username from their route without
native authentication. Changing that authority/lifetime requires explicit analysis of the unsafe-mode
contract, not blindly supplying a nil authenticated UUID or blessing an unverified name. #474 remains
open for this producer and further evidence; a grep result or passing suite is not a global audit.

This third repair subsequently qualified exact head `a594cd5e8ab32b126a7e707727efc3672e69c476`
with seven checks (42608933755), no unresolved threads, and protected matching-head #481 merge
`1e8dbc74c72fdac7e51b3a5c68c7b288b0da95b5`. Exact-main 42611382211 attempt 1 could not
acquire its hosted macOS Go-CodeQL runner (ID 0, no steps); application jobs passed. Its infrastructure
annotation and failed-job-only rerun are retained separately. Attempt 2 passed at the same exact
merged SHA; this does not erase the initial infrastructure failure or imply all jobs were rerun.

## Explicit opt-in diagnostic owner scope

The next real-producer inspection found two source handlers still lending Gin to SQL/SDK, while
`UnsafeResourceGraph` delegated its route-selected username to the authenticated adapter without a
UUID. New actual-handler boundary tests genuinely failed both pooled-context assertions and graph
HTTP 200 versus actual 500 (1.273s). These are deliberately boundary fakes, stopping source requests
before SDK/network creation; the graph fake enforces the store's immutable-owner predicate. They are
not real provider/credential-refresh/graph-database acceptance. The unit graph fixture was later
tightened to a valid `MedicalHistory` type and supplemented with the real-store control below.

The flag-off guard remains first and independently refuses all three entry points. With unsafe mode
explicitly enabled, a request-scoped active-user lookup resolves the selected name once. A deliberately
named **diagnostic owner query scope, not caller authentication**, copies its immutable UUID/name,
masks inherited account/token values for lookup, and suppresses token propagation. It never sets Gin
claims, grants a native login, links an identity or treats a route name as authenticated. Later
same-name recreation cannot rewrite the copied UUID. The graph implementation accepts this explicit
scope through a private helper; its normal public handler still uses authenticated middleware values.
Source SQL and SDK constructors receive the stable request context, not Gin or a timed Gin child.

The old deferred source-credential persistence intent matters if a provider rotates a refresh token
before HTTP abort. Its context now deliberately survives request cancellation but has an independent
10-second bound and the same immutable selected owner. This is synchronous finalization, not an
untracked sender/worker or changed tracked-job cancellation policy. New helper controls verify that
bound, own cancellation, identity copying and no local JWT; they do **not** exercise a real provider's
rotation through the handler. Missing/deleted/nil-UUID/already-canceled selections fail closed before
source work. Flag-off, selected-context/reuse, negative-owner and cleanup controls pass under race
(2.010s). An initial post-fix compile mistake used `.Valid` on the model's `*time.Time` deletion field;
it was corrected to a nil check, preserving the failed compile output separately from the genuine red.

At source checkpoint `88a0cb1107f56f6e905bf1c30cf1ceda2cf81164`, fresh complete web/handler/
middleware races passed (632.976s / 549.296s / 1.649s). Complete serialized native internal races
passed: host 22.831s, composed encrypted shell 27.765s, SMART 5.895s, fixture 1.266s. Existing
macOS target warnings remain. Changed-web vet/diff passed. No deadline/hash relaxation, dependency
change, global fallback, production data/identity/provider or enabling of production unsafe mode occurred.

An additional real encrypted SQLite graph test imports the strict synthetic Patient/Observation via
the real-JWT clinical HTTP route, then registers the explicitly opted-in diagnostic graph route with
no JWT. A compiler overlay of accepted `1e8dbc74c7`'s `unsafe.go` genuinely returns HTTP 500 instead
of 200 (5.224s package); the corrected unchanged real graph and expanded boundary/helper selection
pass under race (6.802s). The owner's Patient is present, another selected owner's response does not
disclose it, and disabling the flag restores 503. Fixture user insertion bypasses password hashing,
not a password/OIDC acceptance test. No external provider/refresh is involved. After adding this
stronger control, fresh complete handler/middleware races passed (558.990s / 1.556s); production
implementation remains the same as the earlier full-web/native checkpoint.

This preserves explicitly opt-in diagnostic capability while constraining owner/lifetime reuse; it
is not a security endorsement of unauthenticated provider proxying/name sync. Existing unsafe
provider error/body/log policy, credential failure recovery, request-path authority and production
configuration remain separate. #474 stays open pending final inventory/review; zero direct grep
matches or a passing suite alone are not a complete caller/security audit.

This fourth repair qualified exact head `ca1672ac68aebb4c1f81748b4db0bffbb1438eab`
with seven checks (42618587011), no unresolved threads, and protected matching-head #482 merge
`92c31890c5ef34f462241d2961186d3b09a45730`. Separate exact-main 42622419523 attempt 1
completed success; local/PR success was not substituted for it.

## OIDC account-actor follow-up

Pi continues #474 on `task/oidc-password-context` from accepted
`1b7a2b243fa78734638cc79d80089255cbb7d9cd`. A further actual-callee inventory found four
aliased `r.GetCurrentUser(c)` calls still passing Gin: password proof shared by secure link/revoke,
identity listing, audit listing and logout. Earlier changes to explicit-ID OIDC repository methods
had not removed these actor lookups. This is not an anonymous username-login path.

All four now use the existing authenticated request adapter: original HTTP lifetime, copied verified
UUID/name, no local JWT, no pooled Gin or timed Gin parent. The downstream explicit user/session
operations keep their original request context and explicit identifiers. Password, local recovery,
OIDC issuer/linking/session/role rules and detached jobs are unchanged; no global fallback is enabled.

The existing real-handler boundary table gains a password-helper case that genuinely failed on
accepted source (3.119s package). Five actual registered secure routes also genuinely fail their
pooled-Gin boundary controls (2.109s): link, identities, audit, revoke and logout. Synthetic identity
middleware and a boundary fake stop before password/session/external work; these tests are **not**
JWT/password/Authentik acceptance. They retain the repository context, then deterministically cancel
and replace the original request/actor, asserting its original deadline/cancellation, immutable
UUID/name and absence of JWT. Both unchanged boundary groups pass under race (3.394s).

The existing real SQLite/owned-TLS OIDC HTTP local-recovery/session/authority test passes ordinarily
(55.920s). An instrumented combined attempt timed out at eight minutes (483.255s package), with the
preserved stack runnable in cost-14 bcrypt password comparison inside the real link path, not a
blocked SQL waiter. This is incomplete race evidence, not a pass or proof of database deadlock.
Substantial unrelated host load had also affected other native signup qualifications; causation is
not claimed from that observation alone. No hash, assertion, registration/native/clinical budget,
priority, OS setting, unrelated process, dependency or provider was changed. Changed-handler vet and
diff checks pass. Fresh hosted/full-suite acceptance and remaining caller inventory are still open;
ordinary flow or these focused boundary races do not certify the entire OIDC/native/production scope.

Exact head `0f54002f7408f0d7eac7fd6ebaa33daf6e8cca67` subsequently passed seven hosted checks
42663352354. After ready-for-review, fresh matching-head/CLEAN/no-unresolved-thread verification
preceded protected #489 merge `6054e7bb50cf7b972bf21d8440410190378c4bea`.
Separate exact-main 42668872750 completed success. This does not close #474 or accept native #487.
