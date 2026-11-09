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
