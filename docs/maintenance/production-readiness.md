# Production readiness

Chartrove is not yet certified for production use. This is the active implementation plan and evidence
index for the autonomous Pi ownership assignment of 2026-11-09 (UTC). Historical checkpoint documents
remain evidence, not current acceptance or authorization. Only this repository is in scope.

## Ownership and execution

- Exclusive writer: the current Pi session; do not restart prior Codex work or create other coding agents.
- Canonical repository: `chartrove/chartrove`; default branch: `main`.
- Canonical checkout after reclamation: `/Users/osama/Code/chartrove/chartrove`.
- Use only the primary checkout, switching task branches there. Do not create linked worktrees.
- Deliver small tested PRs and verify exact merged source. Preserve source, failures and private evidence.
- No live deployment, real medical records, clinical authorization, production identity changes or sensitive
  credentials are authorized. Synthetic isolated acceptance is distinct from live operation.

## Ordered implementation plan

1. Reclaim the repository: inspect all worktrees, branches, stashes and prior closeout; establish no competing
   writer; preserve ignored files and unique history; remove obsolete worktrees normally; relocate the primary
   checkout, fix remotes and active references, and verify CLI and terminal identity.
2. Reassess current source, CI, ADRs and open issues/PRs. Recover draft #466 without assuming its tests or
   review are complete. Review #431 as a proposal, preserving original Fasten provenance and reuse direction.
3. Establish a synthetic baseline across backend, frontend and desktop; inspect trust boundaries, storage,
   family authorization, Authentik, ingestion, export, consent, audit and recovery. Prioritize reproduced
   production blockers, not issue count.
4. Implement and validate bounded corrections, including full import/update integrity and direct SMART/PKCE
   where supported. Chartrove remains authoritative; Medplum and external agents stay optional and isolated.
5. Qualify reproducible immutable release artifacts, supported deployment configuration, monitoring,
   rollback and independently restored synthetic backups. Run end-to-end synthetic household acceptance.
6. Report an exact release candidate only when every acceptance gate below has current evidence; otherwise
   continue safe independent work and identify genuinely external authorization blockers precisely.

## Production acceptance gates

| Gate | Evidence required | State |
| --- | --- | --- |
| Repository custody | [Reclamation receipt](repository-reclamation-2026-11-09.md), one checkout, canonical remote | Verified; Pi launch-path anchor pending session restart |
| Clinical integrity | Lossless import/export, idempotence, updates, attachments and provenance | Reader, [atomic store](clinical-snapshot-persistence.md), owner API and [clinical UI](clinical-import-ui.md) integrated; scoped synthetic fidelity/update/recovery evidence exists; broader ingestion/GUI remain |
| Identity and access | Native recovery; synthetic Authentik; distinct family identities; negative/revocation tests | Native bridge UUID/session binding integrated; bounded stable-context repairs accepted; remaining #474 callers and full household/recovery/Authentik remain |
| Confidentiality | Encryption/key handling, secure defaults, bounded logs and network behavior | [Source request](provider-log-privacy.md) and [dynamic response](provider-response-privacy.md) corrections integrated; broader gates remain |
| Consent and audit | Scoped consent, attribution, transactional audit and live authorization | Owner-confirmed API/UI, active JWT receiver checks and bounded post-commit refresh integrated; household/provider consent still open |
| Recovery | Independently restored encrypted synthetic data, wrong-key failure, restart and rollback | SQL/HTTP snapshot-copy restore tested; independent operator/key-loss and whole-household gates remain |
| Application acceptance | Browser, embedded backend and native desktop synthetic workflows | Pending #487: 788 browser passes/one skip and exact-c9 twelve-check hidden native success; later native races/signup and exact-eed runtime fail; visual GUI still permission-gated |
| Supply chain and release | Dependency findings triaged; immutable source-bound artifacts and reproducible build evidence | Source-bound ad-hoc packages/restores retained; latest exact-7f candidate is unaccepted, not current-main/published/notarized |
| Operations | Documented deployment, health, monitoring, retention, recovery objectives and rollback rehearsal | Not reassessed |

## Accepted source and current work

- Clinical UI #477 qualified corrected head `6d161d4c42c5b6aeec8c6d68e33b4f53bbca5f54`
  (42588079214), then merged at `8629b689e41915b09e95cfba6d9cedc24976bd50`.
  Exact-main validation 42590642336 passed. The coupled stale-identity/cleanup-error regressions,
  exact source-specific display and composed encrypted flow are documented in the UI receipt.
- Stable legacy context repairs #476 and #478 are accepted. The follow-up qualified exact head
  `f094d12f3fbe230f320306191a5131c710682a84` (42592715042), merged with no unresolved threads
  at `dc48128ad8691513ae9c2e8607567482849f6c8f`; exact-main 42594974931 passed.
  [#474 remains open](request-context-isolation.md) for actual remaining callers, not hidden by CI.
- PostgreSQL #480 qualified exact head `e0c02909e447882f1deea92f200ab3eefbd3ec7b`
  (42601840418), merged at `3c1acc588600fe4d5ef8b072696cc6302c2a2da3`; exact-main
  42605419029 passed. Actual pinned disposable PostgreSQL runtime exposed and corrected an
  identical-Patient update misclassification. Exact originals, precision/scale, ownership,
  rollback/reopen and concurrent deletion have real runtime controls. The
  [PostgreSQL receipt](clinical-snapshot-postgres.md) retains #479's instrumented maximum-load
  limit; this is not production PostgreSQL security/operations or independent recovery acceptance.
- Clinical/remaining request-bound #481 qualified exact head
  `a594cd5e8ab32b126a7e707727efc3672e69c476` (42608933755), then protected matching-head
  merged at `1e8dbc74c72fdac7e51b3a5c68c7b288b0da95b5`. Exact-main 42611382211 attempt 1
  failed hosted macOS runner acquisition before any CodeQL Go step (runner ID 0); application jobs
  passed. Failed-job-only attempt 2 subsequently passed at the same exact merged SHA; the initial
  infrastructure failure remains recorded.
- Diagnostic #482 qualified exact head `ca1672ac68aebb4c1f81748b4db0bffbb1438eab`
  (42618587011), then protected matching-head merged at
  `92c31890c5ef34f462241d2961186d3b09a45730`; exact-main 42622419523 passed. The
  explicitly opt-in, unauthenticated query scope pins the selected owner's UUID without Gin/JWT
  authentication claims. Actual encrypted graph/boundary and complete race controls pass.
  Unsafe mode remains default-off/not production-approved; #474 inventory/review remains open.
- Native #483/#484 adds a private create-once bounded smoke receipt and separately gated hidden
  native test windows, corrects the stale SMART smoke argument/identity false-positive control,
  and refuses nonfresh synthetic directories before storage. Clean exact-68e checkpoint
  `68e3c9732e6438c8dbf8ba4dbeda276c3207144a` includes accepted #482: full packaging,
  strict extracted ad-hoc signature, actual packaged encrypted WAL restore (3.82s / 4.19s),
  pre-GUI invalid-flag refusal and all ten actual hidden native WebView/service checks pass.
  A separate private invalid-token overlay at the earlier unchanged smoke-page checkpoint correctly
  fails that control instead of counting any exception as success. See the
  [native runtime receipt](native-smoke-runtime.md). Exact head
  `2a03fea8a4932f5d57ac42e06a5a8c6246eea732` passed seven checks (42624220876), protected
  matching-head merged at `0252c677ffd3c56900eacf60d7c214401a7effe3`; exact-main
  42627751541 passed. This is hidden native runtime, not interactive visual acceptance.
- Accepted #485/#486 confines automatic application JWT insertion/local auth-error handling to the
  configured API's scheme/host/path namespace, not external Lighthouse or prefix siblings. Synthetic
  opaque-native URL and registered-provider controls correct invalid null-origin bases without
  trusting equal null origins. Genuine five accepted-classifier/four native-base reds and a newly
  introduced DI-seam wiring omission are separately recorded. All 41 coupled controls, full frontend
  (772 passes/one skip, Node/production/Storybook), full native race, clean exact-b916 package,
  strict signature, actual packaged encrypted restore and ten hidden native checks pass. The custom
  smoke page does not exercise the normal Angular interceptor. See the
  [application token receipt](application-token-boundary.md). Exact head
  `323aaf278892b740263628735658eaf8c18311eb` passed seven checks at 42630052659 attempt 2;
  initial Desktop failed hosted runner acquisition before steps, not product execution. Protected
  matching-head merge `1b7a2b243fa78734638cc79d80089255cbb7d9cd` and separate exact-main
  42638457680 passed. Failed-job-only rerun did not re-execute all earlier jobs.
- OIDC account-actor follow-up #489 replaces four remaining aliased Gin lookups across five secure
  routes. Genuine helper/registered-route reds, focused lifetime races and ordinary owned-TLS local
  recovery pass; the instrumented bcrypt timeout stays incomplete evidence. Exact head
  `0f54002f7408f0d7eac7fd6ebaa33daf6e8cca67` passed seven hosted checks (42663352354), then
  protected matching-head merged at `6054e7bb50cf7b972bf21d8440410190378c4bea`.
  Separate exact-main validation is pending; #474 final inventory is not closed.
- Current Pi #487 pins legacy file receiving UUID across confirmation/conversion/import, enforces
  optional unambiguous server receiver assertions before SDK/persistent source/job work, and preserves
  absent-field clients/detached jobs. Its actual native gate exposed raw FormData failure missed by
  Go/browser tests. A shared bounded native-only byte adapter has 788 browser passes/one skip and
  exact-c9 actual twelve-check native success. Later full local native races fail signup fixtures;
  exact-eed native attempts time out after three and six checks; the latter substage is unknown. Fixed-label progress
  now preserves partial timeout evidence without extending budgets. A server-guard-bypass negative
  timed out and is not yet calibrated. See the [manual receiver receipt](manual-import-receiver.md).
  This slice is not accepted or overall ready; hosted/current runtime qualification remains.
- The latest ZIP is exact `7f44f3839131ce03aaf110cc95c13d1c2e9f4805`, not a later docs commit,
  accepted native runtime, published/notarized/current-main release, older-macOS or independently
  operated key-custody qualification. Earlier c9 (twelve checks), b916/68e (ten checks), a594/e442
  and negative/diagnostic archives remain distinct and privately retained; positive ZIPs were
  restored/byte-verified after overlay builds.
- Earlier, the normal Cua route launched the exact-a594 clinical inspection app with a fresh synthetic,
  explicitly **unencrypted** SMART-UI-test directory; no record/profile/consent input occurred.
  Exact-window capture instead refuses `direct_capture_consent_required`; TCC booleans do not
  prove direct-capture consent. Two bounded AX reads showed only window/menu chrome. No screenshot,
  blind/foreground input, permission bypass or update occurred. The named lifecycle ended; the
  task-owned synthetic app remains running. Legitimate owner approval via
  `cua-driver permissions grant` is needed for visual QA, not an inferred grant or overall blocker
  while safe independent work remains.

The following slice plans are historical implementation intent. Current acceptance is above and in
exact-source receipts, not the earlier future-tense wording.

## Initial observations (historical)


- Initial primary `main` was clean at `91509eb7e3a23f5683f562d58b5dad1879cd1dbb`, five commits behind
  `origin/main` (`b3144a99a84e7e876bd4e964fc8201d11d5dfa4d`). No stash was listed.
- Three linked worktrees exist: Atlantic blue branding, theme and a detached exact-main build. Their heads
  are ancestors of remote main, and each has no tracked modifications or nonignored untracked files.
  Ignored build products and dependency directories must be preserved before removal.
- No repository `AGENTS.md` exists in the tracked primary tree. Shared owner instructions apply.
- Draft #466 preserves `f5cb683a1315fe932045812aff0e5f9596bf0d7f`. Its closeout explicitly leaves clinical
  persistence and repeat updates unfinished. The original local branch had been removed; a branch at that
  exact object has now been restored. GitHub reports all seven checks successful for that head, but this
  does not supply independent review, current-main compatibility or clinical end-to-end acceptance.
- Draft #431 remains unmerged at `e106e80f199c524c1d6a2461863a3971736051ef`.
- Repository visibility is public. Ignored preservation material must never be staged or uploaded.
- The legacy branch-protection endpoint returned 404, but active ruleset 26337029 enforces PRs, resolved
  threads, strict up-to-date seven-check validation, no deletion and no non-fast-forward. No bypass is needed.
- Reclamation and relocation are complete; see the [source/file preservation receipt](repository-reclamation-2026-11-09.md).
  The only checkout is canonical. Rex reports the new path. Pi's captured launch-path limitation is explicitly
  documented there, with a labelled nonrepository anchor rather than a duplicate clone or symlink.

## Historical plan: atomic clinical persistence

Exclusive Pi ownership continues in the primary checkout, task branch `task/clinical-snapshot-persistence`,
initially stacked on the qualified #466 reader and accepted #470 source-log correction. This is synthetic,
internal-only development; host/API/UI exposure comes after repository acceptance.

1. Revalidate the bounded raw collection at the database boundary, derive source identity from the exact
   FHIR base and Patient ID, and require the authenticated local owner. Never promote imported records to
   editable local provenance or infer grants/native authorization from origin metadata.
2. Commit source creation/reuse, all Patient/Observation upserts, ownership checks, associations, immutable
   original snapshot and payload-free audit together. Retain raw lexical values and unknown extensions;
   indexed projections are not a substitute for originals. No pre-commit success events.
3. Exact historical replay is a no-new-clinical-write acknowledgement, not permission to roll newer records back.
   Changed snapshots update current records while preserving originals and avoiding deletion of absent
   Observations (a bounded snapshot is not evidence of provider-side deletion).
4. Exercise encrypted reopen, distinct owners/endpoints/Patients, raw fidelity, repeated/changed import,
   injected final-write failure and rollback, concurrent import, cancellation, source/account deletion and
   backup/restore. Existing Fasten/manual import and grants must remain unchanged.
5. Require accepted migrations, source review and protected validation before integration. Then add the
   explicit owner-confirmed native bridge/API/UI action and synthetic end-to-end flow as its own slice.

## Historical plan: explicit owner-confirmed import API

Pi owns `task/clinical-snapshot-api`, stacked on the atomic store in #472. Add only authenticated owner
routes: explicit confirmation plus the expected receiving UUID, a bounded raw **string** collection that
never round-trips numeric values through JavaScript numbers, payload-free errors, owner-only metadata
and an attachment-only exact-original download. Publish refresh events only after a successful new commit.
Exercise real encrypted HTTP persistence/restore boundaries and negative/stale-recipient/parser cases.

Native/UI exposure follows separately. In particular, review the current main-window SMART bridge's
local-identity binding before adding clinical retrieval; main-window authorization alone is not evidence
of which local identity owns a provider connection. The UI must confirm the receiving local identity,
clear stale previews on identity change and never label owner-declared origins as provider attestation.
Legacy account/import writer coordination, numeric display precision and narrative/network egress also
remain application-wide acceptance items, not claims supplied by the internal store.

## Historical plan: scope the native SMART bridge

Pi owns `task/scoped-smart-bridge`, initially stacked on the locally tested API. Before exposing clinical
retrieval, bind each new native provider connection to a real authenticated local UUID resolved through
the embedded backend, not browser-decoded claims or a caller-supplied identity. Recheck active identity
before and after Patient/clinical reads and connection completion. Keep the existing main-window boundary,
provider credentials in Go and constant errors; do not add arbitrary URL fetching or an external auth service.
Cleanup must still release in-memory provider credentials after local logout/expiry. Match native/asset
signatures and preserve Patient-only operation, with cross-identity/window/session-change negative tests.
The clinical import UI and its explicit receiving-profile confirmation remain a subsequent slice.

## Historical plan: explicit clinical import in the Sources UI

Pi owns `task/clinical-import-ui`, stacked on the scoped native bridge. Preserve raw Patient text and
collection text rather than reserializing parsed numeric FHIR. Keep Patient-only import usable, add an
explicit bounded Patient/Observation retrieval/preview, name the receiving backend-verified local profile
and require authority confirmation before the owner-bound API write. Guard read/import/disconnect work,
discard stale identity results, keep transport-uncertain outcomes honest and do not infer provider-side
deletions or newer data from a previously seen digest. Exercise real handlers in controlled asynchronous
browser tests, plus synthetic native end-to-end repeated/changed import and encrypted restart/restore.
This plan does not authorize real providers, production identities, deployment or external record uploads.

## Implementation receipt chronology

- Reclamation PR [#469](https://github.com/chartrove/chartrove/pull/469) merged at
  `add1a4ad2cc6c95545c2fb7ead4eff50095df7c4`, after all seven required checks succeeded on
  `5ef82eb0c9f17e0daecb81d5eb358ef862a7524f` (Validation run 42547082306). Exact-main validation is
  separate and must be checked; this is not a product release.
- Resumed #466 merged at `dea9101b4b8959ba5f359fc9ce5d5bb57bdd71ca` after all seven checks passed
  exact head `469f06c16febdc50c501862f21d07a1f518ab9b0` (run 42556131799). Its implementation remains
  an unexposed in-memory reader, not a clinical-import action or end-to-end persistence acceptance.
- The [atomic persistence slice](clinical-snapshot-persistence.md), source checkpoint `626c6db23c`, is
  locally qualified and merged with accepted parent/main history without rewriting it. Exact originals,
  5,000-Observation acceptance, distinct scopes, replay/updates, encrypted restore, rollback/cancellation
  and deletion races have concrete synthetic evidence. The final-metadata focused race suite passed
  (25.102s) and maximum collection passed (32.018s). #472's initial seven green checks were insufficient:
  source review reproduced an entry-URL-only stale provenance defect; tests failed before correction,
  then both resource types and the full focused store suite passed under race (15.861s). The review thread
  was resolved with that evidence. All seven checks then qualified `b5a097f37d` (run 42560019280);
  #472 merged at `af644a006a8b5bc1bc857be0b8aad743b30f7b15`. Exact-main run 42561759380 passed.
- The [explicit owner-confirmed API](clinical-snapshot-api.md) is locally implemented on its own task
  branch. Real JWT middleware, encrypted HTTP originals/history, stale-recipient refusal, rollback/event
  ordering, strict/bounded intake and immutable backup-copy restore passed focused race tests (11.707s).
  Complete web/handler suites passed (63.519s / 51.416s). Native local-identity binding and the clinical UI
  are not implemented by this API slice.
- The [scoped native bridge](scoped-smart-bridge.md) now resolves local identity via the actual embedded
  JWT/account stack and guards connection/Patient/clinical reads before and after I/O. Real encrypted
  signup/delete/username-recreation tests and all native internal race suites passed. A separate
  soft-deleted-account middleware regression failed before correction (204 vs expected 401); the unchanged
  active-account positive control and complete middleware race suite passed afterward. Full Angular
  acceptance passed 724 tests with one skip and retained existing test-template console diagnostics.
  Native package vet/build and frontend structural/compiler validation passed; matching packaged GUI
  and clinical import/receiving-profile confirmation are not yet qualified. The bounded bridge merged
  via #475 after all seven checks qualified `6f6443607002d49ddb83dade6f5beb03c84e34bd` (42575202030),
  with no unresolved review threads. Protected matching-head merge is `5e44373a8496f475edc364552d2070a840d4a3fe`;
  exact-main validation 42578148508 passed. This is not independent security or GUI acceptance.
- #473 source review found synchronous refresh delivery could hold the sole import slot indefinitely.
  A real-bus, unconsumed-client HTTP regression reproduced the committed-response stall. The bus now offers
  optional nonblocking best-effort publication; the API never falls back to blocking delivery or creates
  per-import sender goroutines. The unchanged stalled-client test and a second changed import pass with
  both exact originals retained, while legacy-only notification is skipped with a fixed warning. Fresh
  focused HTTP race (12.949s), event-bus race (1.195s), full web/handler (62.819s / 52.351s), desktop graph
  controls (1.386s) and changed-package vet passed. All seven checks qualified `1b633bd28f` (42565808793),
  but protected merging was refused by a second unresolved source-review thread, not bypassed.
- #473's second review found both original-download slots could be retained by connected non-reading TCP
  clients. A real-socket test with two stalled 8 MiB originals reproduced failure to recover admission
  within 35 seconds. A separate 30-second response write deadline fixes it without altering global server
  timeouts; a subsequent caller retrieves exact original bytes while the first clients stay connected.
  Focused HTTP race passed (47.132s), full web/handler passed (62.264s / 82.430s), and the same real-socket
  test under the desktop graph passed (32.028s). All seven corrected-head checks passed on
  `14eded02f5ebc5518cccbb0963a6c5a37f9926aa` (42568249499), with both review threads resolved.
  Matching-head protected merge succeeded at `3fa5d8e993d1f4bb0d8d2b212eae4690a78a2e8d`; exact-main
  validation 42570752698 passed. This qualifies the scoped API slice, not overall production readiness.
- Initial whole-backend local execution exceeded a 240-second tool timeout during the database package.
  A diagnostic database run with an overly short 120-second test timeout ended during bcrypt cost-14
  fixture creation. No earlier semantic assertion failure was reported. These are incomplete baselines,
  not passing runs or proof of a product deadlock. The full local backend suite subsequently passed on
  `f329ceba66` with a 15-minute per-package budget; database 258.521s, web 101.154s, handler 65.263s.
  This completes the incomplete local baseline; it is not a substitute for later-head hosted validation.
- Source credential payload logging and decoder-input echo were reproduced with synthetic canaries.
  The [focused regression/fix](provider-log-privacy.md) merged via #470 at
  `e4ae20acec3c77be7c09616ce7c7cc05a65dc543`, after all seven checks passed its exact PR head.
- Legacy dynamic registration/token response-body logging was also reproduced using an intercepting
  in-memory HTTP transport. The [focused correction](provider-response-privacy.md) closes rejected replies
  without reading their potentially unbounded bodies. Both real success paths retain positive regression
  controls; the complete models race suite passed. PR #471 merged at
  `0c175f744099c5b10479f27571eec5c8347ce9bb` after all seven checks qualified its exact head;
  exact-main run 42553276768 also succeeded.
- Final bridge integration found that raw Gin contexts do not provide request lifetime by default,
  while globally enabling fallback is unsafe when SQL retains pooled request objects. That global trial
  was rejected after a native encrypted identity test reported a race. Native authentication/account calls
  now use stable request contexts and a copied server-authenticated UUID/username, preserving cancellation
  without pooled Gin lifetime. Native race/vet, snapshot-context tests, full non-race web/handler/middleware,
  desktop Node and full Angular tests passed. Broader race attempts were incomplete and found legacy manual
  import pool races; an exact-main compiler overlay independently reproduces them (56.714s). Pi owns
  [#474](https://github.com/chartrove/chartrove/issues/474) for that next bounded repair; it remains a
  production-readiness gap rather than being hidden behind green non-race CI. Pi owns its local
  `task/request-context-isolation` branch in the canonical single checkout. The first bounded
  [correction and evidence](request-context-isolation.md) now gives manual/related imports immutable
  verified-owner job contexts, preserving the prior detached job lifetime. The unchanged source/graph
  races pass (57.142s), privacy control passes (27.848s), and the complete web race suite passes (639.873s).
  Full-suite findings also led to stable source-access/deletion and OIDC SQL boundaries; full handler and
  middleware race pass after fixing fixtures to supply actual verified UUIDs/requests (550.420s / 1.703s).
  Native desktop graph race also passes. Other legacy web-to-SQL callers still require assessment; passing
  these suites is not global context, GUI, operational identity or independent recovery certification.
- Standalone factory fallbacks still permit unencrypted SQLite and wildcard HTTP listening. Deployment
  configuration and secure-default acceptance require reassessment; the desktop already uses private
  encrypted storage and no application API TCP listener.
