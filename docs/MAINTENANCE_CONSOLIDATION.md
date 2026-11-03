# Maintenance consolidation checkpoint

Owner: maintenance agent in the existing primary checkout. A separately
authorized artifact-retention owner holds its existing worktree and branch;
preserve both. No additional worktrees or subagents were created by maintenance.
GitHub and this tracked checkpoint are authoritative.
This is an incomplete consolidation, not a claim that target counts are achieved.

## Verified state

Snapshot: November 3, 2026, after #395 and individual #267 disposition.
[Machine-readable receipt](maintenance/2026-11-03-checkpoint-after395.json).
The gofhir-compatible-models task branch is additional to this historical snapshot.
The separately owned artifact-retention branch/worktree remains preserved.
Requery live counts after every integration.

| Item | Verified result |
| --- | --- |
| Published application main | `0b6455253299b2f57804cdf6023317e624f30b28` (#395) |
| Open issues | 163, including retained Renovate dashboard #222 |
| Open PRs | 14: eight paused Renovate proposals and six historical drafts |
| Origin heads | 21 |
| Default branch | main tracks origin/main; current FHIR task is additional |
| Cached legacy refs | 17 nonsymbolic refs plus HEAD, preserved unchanged; not live origin heads |
| Tags / worktrees / stashes | 30 unchanged tags; two authorized checkouts now; no stashes |
| Main protection | Active ruleset 26337029, no bypass actors |
| Required checks | Backend, Frontend, Desktop from Actions app 15368; strict up-to-date |
| PR workflow | PR required, zero external approvals, conversations resolved |
| Force push / deletion | Blocked on main |
| Active Actions registrations | Validation only; inherited Development explicitly disabled again |

The six remaining drafts are #209 (permissions),
#211 (Typesense/RAG), #212 (OCR), #214 (OIDC), #215 (delegated access), and #217
(reserved usernames). Their head/base dependencies remain live. These are not
implemented features or blanket retirement candidates. The maintainer explicitly
retained all six capabilities on November 3; corrected implementations and their
individual data/authentication policies remain required. #284/#285/#286 are now
implemented by #354; all remaining dependency PRs require individual review.

## Latest integrated safeguards and tooling

Merged #392 updates Chart.js to 4.6.1 and its color leaf to 0.3.4 with two erased
widget type-parameter corrections. Identical real canvas consumers pass old/new;
frozen install and published integrity checks pass. Protected Backend/Frontend/
Desktop checks pass. The first Desktop job failed before tests on a Go proxy
download timeout; its retained log qualifies the successful isolated retry.
#238 is individually reconciled and its exact archived tip/task branch retired.
No fresh local full production/Storybook/native package was built below the
50 GiB reserve; CI provides full browser/production and native Go checks. Last
locally packaged native source remains 33c64fbdb, before this runtime-library
update, and must not be presented as a new Chart.js package.
[Scope and limits](maintenance/2026-11-03-chartjs-4-maintenance.md).

Merged #393 updates GridStack to 14.0.0 through the attributed custom adapter.
Old/new real-browser geometry, configuration, edit-state and responsive cases pass.
The first production check exceeded the unchanged 10 MB budget by 3.57 kB;
loading the existing guarded dashboard lazily corrects this. Actual router controls
prove authenticated loading and signed-out denial. Final protected checks pass:
541 browser tests, one existing skip, 16 Node tests, and 9.86 MB initial output.
#302 is implemented; #243 is superseded by validated v14, without claiming v8.4
remains installed. Both archived historical tips and the task branch are retired.
Local native chunk loading and pointer dragging remain unverified; no new local
package is claimed. [Scope and retained failures](maintenance/2026-11-03-gridstack-14.md).

Merged #394 integrates #232/#278/#279/#280/#281 as one compatible family in
both Go graphs. The uncorrected version-qualified SQLCipher replacement failed
an actual encrypted-header test. The correction retains the exact cipher source
for every selected SQLite version. A separate cross-table pagination regression
is old-green/new-red/corrected-green. Full backend, desktop, embedded and all
protected checks pass; synthetic encrypted old/new/old and desktop cold opens
preserve records and credentials. Expanded PostgreSQL old/root/desktop fixtures
each pass 14 events with no failures/skips, and no owned containers remain.
All five original PRs are individually closed as implemented; their exact archived
heads and merged task ref are retired. Historical manual retirements now total
82, with 12 previously absent heads still attributed to unknown actors.
[Scope and retained failures](maintenance/2026-11-03-gorm-persistence-family.md),
[guarded retirement](maintenance/gorm-family-retirement.json).

Merged #395 implements #267's Maroto 2.4.4 with its required Go 1.26.1
patch-level minimum, retaining actual Go/CI 1.27.3. Thirty old/root/desktop
structural cases and independent one/seven-page synthetic PDF consumers pass;
all applied and protected checks pass. The original PR is individually closed as
implemented. Its archived exact tip was already absent at retirement, actor
unknown; the merged task ref is retired. Historical manual retirements remain
82 and already-absent heads with unknown actors now total 13. An outer receipt
postprocessing error followed the successful guarded core operation; independent
owner/worktree/clean-main checks qualify the final receipt without repeating it.
[Scope](maintenance/2026-11-03-maroto-pdf.md),
[qualified retirement](maintenance/maroto-retirement.json).

The next FHIR candidate advances the compatible library from 0.0.7 to 0.0.9.
#262's 0.1.1 compiles and accepts shorthand imports but fails typed resourceType/
known-field serialization on stable Go 1.27.3; it is not adopted. A scoped
retirement of that incompatible version follows protected integration of 0.0.9,
without an implementation claim for 0.1.1 or the still-open shorthand issue #53.
All 98 supported canonical outputs among 100 synthetic fixtures match, with two
unchanged unsupported inputs. Raw unknown-number/zero/false/date/Unicode and
SDK interfaces, encrypted old/new/old and desktop cold opens pass. Other selected
versions and all replacements stay unchanged. No real records or native package
are used as validation. [Scope](maintenance/2026-11-03-gofhir-compatibility.md).

Each historical feature draft now has a source-admission review, with concrete
unmerged auth/data/service gaps and preserved capabilities. A strict verified
bundle covers all 2,586 unique objects; exact base OIDs recover from preserved
main even where they are not head ancestors. All six drafts remain open;
no feature is falsely implemented or blanket-retired.
[Individual review](maintenance/2026-11-03-historical-feature-admission.md),
[recovery proof](maintenance/remaining-features-recovery.json).

Merged #391 preserves patient detail missing-date/age display, the original
artifact-retention commit and a tested cross-filesystem documentation preservation
correction. Backend/Frontend/Desktop checks pass, including all 16 Node cases in
Linux CI; signed native ten-check smoke passes with qualified production-input
reuse. Primary main is clean; the patient task branch was retired at its exact
integrated tip. The retention owner's branch/worktree remains untouched. Storage
cleanup/adoption is complete and must not be repeated.
[Patient evidence](maintenance/2026-11-03-patient-detail-age.md),
[retention adoption](maintenance/2026-11-03-retention-adoption.md).

The separate ng2-charts #307 major migration remains pending; #392 does not
implement its new wrapper APIs. Local focused checks do not bypass full-run
storage admission, and no storage cleanup or retention adoption is repeated.

Merged #390 migrates ngx-highlightjs to the Angular 22 compatible API. Browser
consumer cases, 529 full browser tests (one existing skip), production/Storybook,
signed native ten-check smoke and protected checks pass. Isolated native light
theme JSON tokens are observed; native XML and dark contrast remain unverified.
#308 is reconciled and its exact archived tip retired. The resulting synthetic
patient-card NaN finding is corrected by #391.
[Patient detail evidence](maintenance/2026-11-03-patient-detail-age.md).

The maintainer chose to retain provider-addition requests as future product work.
Those unresolved issues remain open; the one-issue goal is not achieved by
retiring them. [Decision and individual-review boundaries](maintenance/2026-11-03-provider-future-scope.md).

Merged #389 changes only the devcontainer Node feature key from :1 to :2, with
Node26.13.0/Go1.27.3/Chrome and all other configuration retained. Published OCI
integrity/source checks, JSONC semantics, isolated linux/arm64 actual Makefile
Corepack/Yarn validation and all three protected checks pass. Full Go/Chrome
project devcontainer and amd64 installation remain unverified. The initial global
Corepack harness failure is retained. Sparse patterns and the original S flag
remain unchanged; the excluded file was never materialized. #318 is individually
reconciled; exact historical/task heads retired after fresh gates.
[Scope](maintenance/2026-11-03-devcontainer-node-feature.md),
[retirement](maintenance/node-feature-retirement-1.json).

Merged #388 removes the unused direct idb declaration and sole leaf lock block.
All 72 retained direct versions and remaining lock bytes stay unchanged. Frozen
install, ten Node checks, 523 browser tests (one existing skip), production/
Storybook, a fresh signed native ten-check smoke and all protected checks pass.
Original installed metadata and the exact historical tip are archived. #303 is
explicitly retired; idb 8, offline storage #19 and draft persistence #29 are not
implemented. A first disposition guard used the React intake path and failed
before mutation; the corrected full-content/head check passed before closure.
[Scope](maintenance/2026-11-03-unused-idb-retirement.md),
[retirement](maintenance/idb-retirement-1.json),
[harness correction](maintenance/idb-disposition-harness-correction.json).

Merged #386 updates ng-select to 24.2.1 and corrects the wizard lazy-loading
boundary within the unchanged 10 MB production budget. Selection cases pass on
old/new versions; 523 browser tests, production/Storybook, signed native smoke
and all protected checks pass. Synthetic native UI exercises both entry points
and disabled empty attachment pickers; native multiple-selection remains
unverified (covered by browser DOM cases). #295 is individually reconciled.
[Scope](maintenance/2026-11-03-ng-select-attachments.md),
[validated historical and approved narrative retirement](maintenance/ngselect-and-narrative-retirement.json).

Merged #387 adopts React 19 development tooling with ten Node checks, 523
browser tests (one existing skip), production/Storybook builds, a fresh signed
native ten-check smoke and all three protected checks passing. Interactive
Storybook is unverified because the connector and staged browser signature
checks failed; diagnostics remain intact. #326 is individually reconciled;
its archived exact tip and merged task branch were retired after fresh gates.
[Scope and limits](maintenance/2026-11-03-react19-tooling.md),
[retirement](maintenance/react19-retirement-1.json).

Merged #385 adopts jose 6.2.15 for the two existing decodeJwt consumers. Thirteen
real AuthService cases pass on old/new versions; the full suite passes 513 tests
with one existing skip. Node checks, production/Storybook, signed native ten-check
smoke and all protected checks pass. Native sign-in renders the synthetic user
and Administrator; sign-out returns to the empty form. Both raw synthetic records
and seed database bytes remain unchanged. Its read-only verifier created an empty
WAL and SHM metadata in that synthetic seed; both sidecars and original receipts are
retained, with [qualified correction](maintenance/synthetic-seed-sidecar-correction.json).
#305 is closed as implemented; #246 is superseded
by v6, without claiming that v4 remains installed. Both historical branches were
already absent at retirement checks, actor unknown; exact archived tips remain
recoverable. The integrated task branch was deleted after guarded checks.
[Evidence](maintenance/2026-11-03-jose-runtime.md),
[retirement](maintenance/jose-retirement-1.json).

Merged #384 updates tslib's historical direct selector to 2.9.1. Other lock
bytes, package.json, separate selectors and 72 other direct versions remain
unchanged. All 38 upstream Node tests (no skips/todos), ten existing Node checks,
500 browser tests (one existing skip), production/Storybook, fresh signed native
ten-check smoke and all protected checks pass. All 22,287 frontend files match;
original output content/stat/xattrs restore exactly. #252 is individually closed
as implemented; its verified archived head and integrated task branch are retired.
[Evidence](maintenance/2026-11-03-tslib-maintenance.md),
[retirement](maintenance/tslib-retirement-1.json).

#56 is separately retired as an accepted upstream support explanation whose
remaining task was a Fasten FAQ update. Already published fork documentation
records the Lighthouse/direct-SMART boundary. No self-hosted Lighthouse or
individual provider fix is claimed. Original authorship/discussion and before/
after evidence remain intact. [Disposition](maintenance/lighthouse-question-disposition.json).

Merged #383 fixes manual Encounter descriptions by falling back to serviceType
while preserving provider type priority. Missing Period endpoints show Not
provided; raw records and wizard conversion remain unchanged. Nine targeted
failures become thirteen passing cases with identical test bytes; the full suite
passes 500 tests with one existing skip. Production/Storybook, ten fresh native
checks and all protected checks pass. Native UI shows the synthetic row and opens
the same record, with both raw records unchanged. GitHub closed #104 on integration;
the exact task head is integrated and retired. #69 remains open. Source-profile
dark-theme title contrast remains a separate observation.
[Evidence](maintenance/2026-11-03-manual-encounter-rows.md).

Merged #382 supplies current local data/security guidance, an empty required-key
example and corrected desktop/toolchain/PostgreSQL prose. #88 is closed as the
requested scoped guidance; #193 is explicitly retired as an attributed upstream
deployment reference, without an integration/custom-certificate API claim.
Eight corrected config observations and all protected checks pass. Its initial
harness briefly entered locked standby and listened on *:8080; the owned process
was stopped, no database files were open in the captured snapshot, and both
initial runs remain failed diagnostics. Corrected positive controls never start
a server. All incident/source/process evidence and limits remain documented.
[Evidence](maintenance/2026-11-03-local-data-guidance.md).

Merged #381 updates Goja and required semver/regexp2/v2 in both graphs, retaining
Go minimum, original checksums and database/provider replacements. Actual graph
checks, isolated searches/races, full models/web, upstream suites with explicit
skips, generator invariance, signed native ten-check smoke and protected checks
pass. Strict canonical outputs match 88 supported synthetic examples across 32
types; raw key-order variation also occurs on repeated old-engine runs. The same
12 fixture files remain unsupported. #220 is closed as implemented; its verified
historical tip and integrated task head are retired.
[Evidence](maintenance/2026-11-02-goja-runtime.md),
[retirement](maintenance/goja-retirement-1.json).

Merged #380 implements Moment 2.31.0, preserving unrelated lock bytes and 72
other direct versions. Frozen install, ten Node checks, 487 browser tests (one
existing skip), production/Storybook, 3,943 upstream tests/165,058 assertions
(no skips/todos), actual formatter comparisons in two timezones and fresh native
ten-check smoke pass. Native UI renders a valid age and dated vital; both raw
synthetic records remain unchanged. All protected checks pass; #249 is closed as
implemented and its archived exact head and integrated task head are retired.
[Evidence](maintenance/2026-11-02-moment-maintenance.md),
[retirement](maintenance/moment-retirement-1.json).

Merged #378 corrects the source-profile missing-age display. Three DOM regressions
fail before repair and pass unchanged afterward; valid zero/adult controls remain.
The full frontend suite passes 479 tests with one existing skip, production and
Storybook build, and a fresh native package passes all ten synthetic checks.
Native UI shows Not provided while retaining all three synthetic records. All
three protected checks pass; the integrated task head is retired.
[Scope and native evidence](maintenance/2026-11-02-patient-age-display.md).
Merged #379 separately fixes the dashboard empty-array/invalid-date and stale-age
paths. Six regressions fail before repair and all ten targeted cases pass with
identical test bytes after it. The full suite passes 487 tests with one existing
skip; production/Storybook, fresh ten-check native smoke and all protected checks
pass. Native UI shows Age not provided; all three synthetic raw resources remain
byte-identical. The integrated task head is retired. No broader calendar-date
policy or persisted Patient change is implied.
[Dashboard evidence](maintenance/2026-11-02-dashboard-age-display.md).

Merged #356 aligns the two excluded Go generator tools. All 57 model outputs,
related-version JSON and executable package inputs match; application verification,
builds, desktop/encrypted-backend tests and all three protected checks pass.
#258/#276 are individually reconciled; both archived heads were manually retired.
The unchanged upstream Jennifer example-vet failure remains explicit.
[Scope and evidence](maintenance/2026-11-02-generator-tools.md).

Merged #357 adopts direct exact OAuth 1.2.1 in place of the deprecated forwarding
wrapper. All 73 other direct versions are retained; 423 browser tests, bridge,
production/Storybook, compiled-source comparisons, ten native checks and all three
protected checks pass. #224 is individually superseded; its incidental 1.4.1
delegate is explicitly not adopted. The archived head was manually retired.
[Scope](maintenance/2026-11-02-direct-oauth-patch.md),
[retirement](maintenance/oauth-patch-retirement-1.json).

Merged #358 aligns the actual existing Jasmine 4.6 engine, declarations and HTML
reporter without forced overrides. Two guards fail before alignment and pass
with the bridge suite (seven total); 423 browser tests, HTML diagnostic,
production/Storybook and all three protected checks pass. Every produced app file
matches the current native frontend. #237 is replaced by matching declarations,
#247 is implemented and #245/#297/#304 are explicitly retired without claiming
5.13/7 adoption; all five archived heads were manually retired.
[Scope](maintenance/2026-11-02-jasmine-alignment.md),
[retirement](maintenance/jasmine-retirement-5.json).

Merged #359 adopts the bounded React/DOM 18.3.1 tooling patch and required
scheduler, with 72 other direct versions retained and every produced app file
identical. All required checks pass; #282 is reconciled and its archived head
manually retired. [Scope](maintenance/2026-11-02-react18-tooling.md),
[retirement](maintenance/react18-retirement-1.json).

Merged #360 adopts exact Compodoc 2 with a documentation-only correction for
primitive Storybook controls. All entity/public member inventories remain; actual
reader probes retain 328 public descriptors and 254 boolean/number controls.
Ten Node checks, 423 browser tests, production/Storybook and all protected checks
pass; every produced application file stays identical. Internal type-rendering
changes and initial diagnostic failures remain explicit. #235 is superseded,
#293's target implemented through this corrected replacement; both archived
heads were manually retired. [Scope](maintenance/2026-11-02-compodoc2.md),
[retirement](maintenance/compodoc-retirement-2.json).

Merged #361 removes only the unused humanize-duration declaration/leaf lock
block, with 73 retained direct versions and application output unchanged. #244
is explicitly retired without its 3.35.0 upgrade. #210 is superseded by inherited
hover-width removal, without adopting its 280px policy or closing #167/#99.
All protected checks pass; all three archived refs were manually retired.
[Scope](maintenance/2026-11-02-unused-duration-and-layout-review.md),
[retirement](maintenance/duration-layout-retirement-3.json).

Next defect correction: [wizard invalid-section feedback (#102)](maintenance/2026-11-02-wizard-validation-feedback.md).
Three DOM regressions fail before repair and pass unchanged afterward. Eight
targeted wizard tests, 426 full browser tests, ten Node checks, production/Storybook
and ten fresh native checks pass. Save guards and submission code are unchanged;
all three protected checks still gate integration and issue closure.

Merged #354 adopts verified SHA-pinned v7 actions, unchanged project
toolchains/permissions/commands/check names, and disabled credentials/caches.
Actionlint, parsed semantic comparison and all three required hosted checks pass.
#284/#285/#286 are closed as implemented. Their exact archived heads were retired
with fresh dependency checks and expected-tip leases.
[Review](maintenance/2026-11-02-validation-actions-v7.md),
[retirement receipt](maintenance/actions-v7-retirement-3.json).

Merged #353 removes the token-response payload from an existing
identity callback console log while preserving exchange/storage behavior. A
real-SDK synthetic transport regression fails on leaked markers before repair
and passes afterward; full ordinary browser command exits 0 (423 passes, one
skip), bridge, ten freshly packaged native checks and all protected CI pass. [Scope and evidence](maintenance/2026-11-02-idp-callback-log-privacy.md).

Merged #352 adopts @types/fhir 0.0.44 and @types/qrcode 1.5.6.
Five compile errors exposed by the new FHIR declarations are corrected through
erased annotations/imports. Production, ordinary 422-test browser suite and
Storybook pass; executable assets match the current native package, with only
two type edits embedded in one source map. All required CI passed; #225/#227
are reconciled as implemented. Their exact archived heads were already absent
after closure; no deletion actor is inferred. [Scope and evidence](maintenance/2026-11-02-fhir-qrcode-declarations.md).

Merged #351 removes the unused direct UUID v9 dependency while
retaining the existing Web Crypto helper and UUID v8 tooling consumers. All 74
retained direct versions and all generated application assets stay identical;
frozen install, bridge, production, Storybook and protected checks pass.
#230/#314 are explicitly retired, not upgraded. Their exact heads were already
absent after closure; no deletion actor is inferred. #46 is retired as an
informational reference after a fresh body/comment check. #193 remains open for
its unresolved custom-TLS request.
[Scope and recovery](maintenance/2026-11-02-unused-uuid-retirement.md).

Merged #350 addresses manual-upload progress (#113). Seven new
regressions fail before repair and pass afterward; the full ordinary browser
command exits 0 with 422 passes/one existing skip, five bridge tests pass, and
ten newly packaged native checks and all three required protected checks pass. [Scope, data boundaries and evidence](maintenance/2026-11-02-manual-upload-progress.md).

Merged #349 adopts Karma 6.4.4 through three lock lines.
Its executable distribution is identical to 6.4.3; all 74 other direct versions
stay unchanged. The ordinary full browser command exits 0 (415 passes, one skip).
#228 is reconciled as implemented by this replacement; the full wider refresh is
explicitly superseded, with the exact original head archived and safely retired. [Scope and recovery](maintenance/2026-11-02-minimal-karma-patch.md).

Merged #348 removes the unused upstream Chromatic publishing CLI. Runtime selections and all 75 retained direct versions stay unchanged;
frozen install, production/Storybook, bridge and all protected checks pass.
#239/#299 are explicitly retired, not upgraded; their two exact archived heads
were retired after fresh dependency checks and atomic expected-tip leases. [Scope and recovery](maintenance/2026-11-02-chromatic-retirement.md).

The optional address/contact edit-dialog crash (#105) is repaired by merged #347.
Eleven new regressions and the full browser suite (415 passes, one existing skip,
exit 0), all three protected checks, and ten current native checks pass. [Scope and evidence](maintenance/2026-11-02-optional-contact-dialogs.md).

#343 integrated the shared Go utilities after both graph checks, generator
invariance, fresh PostgreSQL, protected CI and ten current native checks passed.
#233/#265/#266/#269 are reconciled as implemented. Their four heads were already
absent after closure; exact tips remain in the verified recovery bundle, and no
deletion actor is inferred. [Scope](maintenance/2026-11-02-shared-go-utilities.md),
[recovery](maintenance/utility-heads-recovery.json), [retirement observations](maintenance/utility-retirement-4.json).

#344 integrated safe default Karma output after protected checks. Real red/green
synthetic-report, full suite and explicit-directory runs pass without changing
thresholds, tests or application code. [Isolation proof](maintenance/2026-11-02-coverage-isolation.md). Original reports
are now separately [recovered and verified](maintenance/2026-11-02-primary-coverage-recovery.md).

#345 retired unsupported TSLint/Protractor starter targets while preserving every
source byte and the installed dependency tree. #226/#251/#300/#312 are explicitly
retired, not upgraded. Runtime selections and 76 retained direct package versions
stay identical; unit/production/Storybook, all required CI and ten current native
checks pass. Their four exact archived heads were retired with fresh dependency
checks and atomic expected-tip leases. [Scope](maintenance/2026-11-02-frontend-tooling-retirement.md),
[retirement receipt](maintenance/legacy-tooling-retirement-4.json). A maintained
frontend linter/full browser journey suite remains unconfigured; retired commands
are not falsely reported as passing or as equivalent to native smoke coverage.

## Integrated and resolved work

Every integration below passed the required checks through main protection.
Task-created branches were removed only after merge ancestry and exact-tip checks.

| PR | Result |
| --- | --- |
| #332 | Retire inherited publishing workflows; validation-only CI; bounded Renovate intake |
| #333 | Individually evidenced upstream administrative and superseded-PR dispositions |
| #334 | Required matching signup password confirmation; issue #124 |
| #335 | Reject shared/default server JWT signing keys; remove config/token dumps; issue #52 |
| #336 | Synchronously release owned callback listener on cancellation; reproduced race fixed |
| #337 | Guard invalid/duplicate saves and show pending state; remove clinical payload log; issue #35 |
| #338 | Reject deleted/revoked sessions and wrong auth schemes; remove raw-header logs; issue #103 |
| #339 | Current-graph word-wrap/semver/follow-redirects replacements for #198/#201/#206 |
| #340 | CryptoJS tooling compatibility for #199; supported TypeScript bound; retire standalone #313 |
| #341 | Minimal footer replacement for #197; equivalent Compose cleanup for #216 |
| #342 | Concise restart checkpoint and current native/ref receipts |
| #343 | Shared Go utility alignment for #233/#265/#266/#269; unchanged generated models/replacements |
| #344 | Fresh private coverage output by default; real synthetic-report preservation regression |
| #345 | Retire unsupported starter tooling; retain exact source/dependencies and supported validation |
| #346 | Current runtime, retirement and restart evidence checkpoint |
| #347 | Optional address/contact edit-dialog crash repaired; issue #105 |
| #348 | Remove unused upstream Chromatic publishing CLI; explicitly retire #239/#299 |
| #349 | Minimal Karma 6.4.4 target; wider #228 graph refresh explicitly superseded |
| #350 | Accessible manual-upload progress, duplicate guard and outcome cleanup; issue #113 |
| #351 | Remove unused direct UUID v9; retire #230/#314 and informational reference #46 |
| #352 | FHIR/QR declaration alignment and erased bundle annotations for #225/#227 |
| #353 | Prevent existing frontend identity callback from logging token payloads |
| #354 | Adopt reviewed v7 actions in validation only; reconcile #284/#285/#286 |
| #355 | Publish verified original-primary coverage recovery and metadata limits |
| #356 | Align excluded model/version generators; reconcile #258/#276 |
| #357 | Direct exact OAuth patch; supersede deprecated wrapper #224 |
| #358 | Align actual Jasmine tooling and CI guards; individual five-proposal disposition |
| #359 | Paired React 18 patch with required scheduler; reconcile #282 |
| #360 | Compodoc 2 with preserved primitive documentation controls; reconcile #235/#293 |
| #361 | Retire unused duration package and superseded source-filter layout proposal |
| #362 | Identify invalid wizard sections without weakening Save guards; issue #102 |
| #363 | Correct native related-resource upload and prove actual #101/#168 chooser/import/save acceptance |
| #364 | Align name parser and remove practitioner payload logs; reconcile #229 with native name persistence |

Closed issues #73/#109/#162 have current implementation/test evidence. #90/#115
are explicitly retired upstream marketing/status-page requests. Earlier ten
upstream administrative/release closures and other reviewed PR dispositions remain
in [the first disposition receipt](maintenance/2026-11-01-dispositions.json),
[additional PR receipt](maintenance/2026-11-02-additional-dispositions.json), and
[individual issue evidence](maintenance/2026-11-02-verified-issue-dispositions.json).
No unresolved issue was closed merely to reduce counts.

For detailed source rationale and limits, read the dated per-change documents in
`docs/maintenance/`. The earlier chronological journal remains recoverable at
[main #341's ledger](https://github.com/karamhouse/chartrove/blob/40040ac30862326ed07bbf71255a8b2d9beb8283/docs/MAINTENANCE_CONSOLIDATION.md).
Read merged #323/#324 and both original maintenance documents before resuming;
do not repeat Angular 14–22 modernization or restore obsolete implementations.

## Validation and data boundaries

All merged batches passed Backend/Frontend/Desktop. Frontend full suites currently
have 429 passing tests and one existing skip in the practitioner correction.
Production builds retain the 9.93 MB budget/CommonJS warnings. Local Chrome teardown sometimes required forced
termination after successful assertions; required fresh GitHub checks still passed.
No timeout, test assertion, encryption contract or branch rule was relaxed.

Fresh disposable PostgreSQL checks pass (8.304 seconds), including date-column
startup/persistence/reopen, parameters, credential-free logs and startup cleanup.
Its missing-image preflight and pinned-image pull are preserved. No test container
remained. Callback cancellation fails before repair, passes 200 repetitions after
repair, and passes race checks. Existing encrypted desktop integration also passes.

Merged #363 has a fresh packaged app from source
`340d02293`. Production build, ZIP CRC, strict extracted signature and all ten
native smoke checks pass. [Artifact and preservation receipt](maintenance/native-340d02293-validation.json).
Actual native UI acceptance now proves the manual-upload chooser/import (#168)
and, after correcting generated-file multipart serialization, attachment
popup/selection/save/persistence (#101). Read-only verification of the stopped
synthetic database confirms the exact 78 attachment bytes and preserves the
original three synthetic resources. An encounter navigation reload was needed
for the new report link; automatic in-place graph refresh remains unverified.
[Acceptance and limits](maintenance/2026-11-02-native-attachment-upload.md).
Both issues were individually closed after all three required checks passed and
#363 merged. Merged #364 updates the name parser with 72
other direct versions retained, and removes clinical payload/response/error
console logging with red/green synthetic regressions. Fresh package `9add19dd1`
passes production build, CRC, strict signature and ten native checks. Actual wizard creation/save and read-only persisted-name checks
pass; original output contents/stat fields and all xattrs match after restoration.
[Scope and acceptance](maintenance/2026-11-02-name-parser-patch.md).

During #363, original contents/stat fields matched while three OS-managed
provenance xattrs changed, with exact old/new values retained. The latest
practitioner build also restores all xattrs exactly. No provenance
attribute was rewritten. Earlier ZIPs and original coverage recovery remain.
Tests and UI acceptance use explicit synthetic storage with production providers
unconfigured. No real family records or purchased Fasten storage were accessed.

Merged #365 retains all four factory sources and 72 other direct
versions. All 22,287 produced app files match the validated 9add19dd1 native
frontend, and Go/Wails sources/modules are unchanged. Storybook now runs under
the existing pinned Yarn environment so Compodoc works on Node 26. Full browser,
Node, production and Storybook checks pass; generated metadata is retained and
the original documentation.json absence restored. All protected CI checks pass;
#242 is individually reconciled and its archived head manually retired.
[Scope and evidence](maintenance/2026-11-02-fishery-tooling.md).

Merged #366 adopts maintained v2.27.7 and its required helpers, with
server commands validated and the native production import list byte-identical.
#331 adds an unused v3 requirement without migrating imports; that addition is
explicitly retired by the integrated policy, not described as implemented v3.
All three protected checks pass. Both historical heads are individually
reconciled and manually retired with fresh dependency checks and exact-tip leases.
[Scope and validation](maintenance/2026-11-02-cli-v2-maintenance.md).

Merged #367 for #259 has one selected module change, passing
before/after repository graph assertions, upstream tests, vet and builds. Desktop
production imports are byte-identical; all three protected checks pass and its
archived head is manually retired. #182 is closed as a duplicate of still-open
#53, whose synthetic M/m rejection is reproduced in both Go graphs, not fixed.
#194 and #217 remain unresolved after read-only scope review.
[Scope and receipts](maintenance/2026-11-02-graph-alignment.md).

Merged #368 for Zone.js #234 changes only its supported manifest range and
lock entry. All 72 other direct versions stay fixed; frozen install, package
integrity, 429 browser tests (one existing skip), ten Node checks and static
Storybook pass. Original metadata/coverage states are preserved. Fresh production
build, ZIP/signature and ten synthetic native WebView checks pass at 5ad6c590a;
original output contents/stat fields and all xattrs are restored exactly.
All protected checks pass; #234 is individually reconciled and its exact
archived head manually retired.
[Scope and preservation](maintenance/2026-11-02-zone-maintenance.md).

Merged #369 implements the corrected JWT v5 #329 replacement, preserving issued-at rejection
with explicit validation after a reproduced default-parser regression in both
graphs. Independent legacy HMAC/wire checks pass both directions, as do auth,
model, desktop, encrypted-backend, upstream, vet and race checks. Auth scans
show no JWT or called findings; the unrelated required-module OpenPGP advisory
remains. Fresh native Go packaging at 7cca0f39f reuses the byte-identical qualified
frontend; ZIP/signature and all ten synthetic WebView checks pass. Original
output contents/stat fields and all xattrs restore exactly. All three protected
checks pass; #329 is individually reconciled and its archived head retired.
[Scope and receipts](maintenance/2026-11-02-jwt-v5-migration.md).

Merged #370 implements the GitHub Gist client replacement for #330 with reproduced JSON
null and cache-ID boundary corrections; local handler, desktop, vet and upstream
checks pass. Fresh native packaging at f22b03676 retains all 22,287 qualified
frontend files; ZIP/signature and ten synthetic native checks pass. Original
outputs/stat fields and all xattrs restore exactly. All three protected checks
pass; #330 is individually reconciled and its exact archived head retired.
[Scope and recovery](maintenance/2026-11-02-github-client-migration.md).

Merged #371 implements the slice-helper #270 replacement with passing old/new query/summary
assertions in both Go graphs and 57 byte-identical isolated generator outputs.
The published examples fail to compile after upstream release test removal; the
matching runtime's complete upstream parent suite passes 949 tests without skips.
Applied module/vet, desktop/encrypted-backend and server build checks pass.
Fresh native packaging at 93aa68702 retains all 22,287 qualified frontend files;
ZIP/signature and ten synthetic native checks pass. Original outputs/stat fields
and all xattrs restore exactly. All three protected checks pass; #270 is
individually reconciled and its exact archived head retired.
[Scope and preserved diagnostics](maintenance/2026-11-02-lo-maintenance.md).

Merged #372 implements the #106 correction: shared/fallback loading, error/retry and request
lifecycle handling. Fourteen actual-component/HTTP regressions fail before repair
and pass unchanged afterward; the full browser suite passes 443 specs with one
existing skip, ten Node checks, production and static Storybook builds pass.
Fresh native packaging at 2f0fd5bfc, ZIP/signature and ten synthetic checks pass.
Actual native Encounter rendering/row navigation pass with the three synthetic
records retained. Original outputs/stat fields and all xattrs restore exactly.
All three required checks pass; #106 is individually closed with its evidence.
The task head was retired after exact-tip, dependency and merge checks.
[Scope and diagnostics](maintenance/2026-11-02-datatable-loading-feedback.md).

Merged #373 preserves zero/false/absent results and inclusive
reference limits. The full browser suite passes 474 tests with one existing skip;
Node, production/Storybook, native ZIP/signature and ten synthetic checks pass.
Native synthetic table/false/pending-card acceptance passes; an unintended
post-close Cua relaunch is explicitly documented with its standby boundary.
[Scope and incident](maintenance/2026-11-02-observation-value-boundaries.md).
All three required checks passed; its exact task head is integrated and retired.
Newly observed local `.kilo/` metadata
has unknown provenance. This task never changed it; a subsequent read-only
observation found it absent, with no actor established. Its original hash receipt
is retained; no restoration or unrelated cleanup was attempted.
[Metadata receipt](maintenance/observation-held-metadata.json).

Merged #374 preserves explicit database, loopback and
unsafe-route settings in a separate immutable snapshot. Two actual regressions
fail in each Go graph and pass unchanged after repair. Full web suites (16 each),
unlock race suites (four each), both vet checks, desktop tests and server build
pass. Dependency and frontend bytes remain unchanged. Fresh native packaging at
6368cdcc3 passes ZIP/signature and ten synthetic checks; all 22,287 frontend
files and original output metadata are preserved. All three required checks pass;
its exact task head is integrated and retired.
[Scope](maintenance/2026-11-02-unlock-configuration.md).

Merged #375 implements the Viper #272 replacement using required library-family updates while
retaining Go minimums, database/provider replacements and all original checksum
bytes. Four old/new configuration probes, 17 config/web tests per graph, 139
upstream events (one Linux-only skip), both vet checks and unlock race suites,
desktop tests and server build pass. Full selected/import graph deltas are retained;
no production module is downgraded. Today's module scans retain only the known
unimported deprecated OpenPGP finding. Fresh native packaging at b36c00e74
passes ZIP/signature and ten synthetic checks; all qualified frontend files and
original outputs/metadata remain identical. All required checks pass; #272 is
individually reconciled and both exact reviewed heads are retired.
[Scope and limits](maintenance/2026-11-02-viper-maintenance.md).

Merged #376 implements the x/exp #221 replacement, retaining the Go 1.26 minimum and database/provider
replacements. Seven query/summary checks pass before/after in both graphs; 116
upstream used-package tests pass. All 57 isolated generator files and related
metadata are identical, and both dashboard cache boundary suites pass. Module,
vet, race, desktop and server checks pass; only x/exp changes in either production
import closure. Fresh native packaging at 48db0aba7 passes ZIP/signature
and ten synthetic checks, with all frontend files and original output metadata
preserved. All three required checks pass; #221 is individually reconciled and
both exact reviewed heads are retired.
[Scope and provenance](maintenance/2026-11-02-exp-maintenance.md).

Merged #377 corrects stale README toolchain/CI statements,
documents the protected PR workflow and post-close native automation guard, and
separates current provider qualification from dated sandbox protocol evidence.
All 23 relative links/anchors pass; application/dependency source is unchanged.
Historical #191 is individually retired as an upstream-site request, with maintained
fork support alternatives and explicit production limitations. All three protected
checks pass; its exact task head is integrated and retired.
[Scope and disposition boundary](maintenance/2026-11-02-support-documentation.md).

The source-profile and dashboard missing-age corrections are integrated as
#378 and #379, respectively, with validation and preservation evidence above.

## Preservation and branch retirement

Sparse settings remain SHA-256
`d48de194029f797d33694f8387ee15180f1531af423402379ef199a9f2ca04bc`;
origin fetch stays main-only. Original Chartrove/FastenCommunity ZIP hashes,
30 tag OIDs and all legacy cached refs match intake. Original handoff receipt
remains `f89ee0c67a049bb3e2b1d698ebd89b3b652c4dcac592547aafd9f1cdc91b50af`.
Existing bundles, editor settings, Finder snapshots, archives and untracked work
are retained. Native builds preserve/restore existing output trees before writing.
Latest restoration (3b9e3a55b) matches contents, inodes, permissions, ownership,
birth times, mtimes, flags and all xattrs. The preceding 340d02293 build and earlier
c957 build each recorded three OS-managed provenance attribute changes, with
exact old/new values preserved; no security-attribute rewrite.

The 17 cached legacy tracking refs plus HEAD now also have a private complete
standalone bundle and local config/ref/reflog archive, with strict isolated object
verification and no main prerequisites. They remain unchanged, with unique old
prototype work held for reconciliation; none is falsely counted as retired.
[Recovery and remaining review](maintenance/2026-11-02-legacy-tracking-preservation.md).

Recovery references (incremental bundles need retained main prerequisites):

- [Original intake bundle and receipt](maintenance/2026-11-01-history-receipt.json),
  SHA-256 `9f0781f4c5858f7f3f9e9ea2d33f8eb377cd1f8331b0242f7c802b3b408966b1`.
  Strict isolated pack verification includes all 6,357 required unique objects;
  [object proof](maintenance/history-object-proof.json).
- [Rebased replacement heads](maintenance/replacement-heads-receipt.json),
  SHA-256 `fdfb8768a50254af6c01201fa7d73d28f76a0d9ac7737342acbd1f91be4b9141`.
- [Current TypeScript #313 tip](maintenance/typescript-pr313-recovery.json),
  SHA-256 `3d505089b8728c4e91056169ff2c792b964aac96384877c89dd01c9f321a8c78`.
- Prior owner archives remain under Library/Application Support/Pylonary/
  agent-tasks/legacy-worktree-retirement-20261026/recovery/retirement/:
  historical bundle `03ffd099a7b71a3fa14ffa16982d423b13d79e5463297709c01392ca84535325`,
  pass-two bundle `63e228d76a6148278ec38b948bc7fc5c23a1b7e2b83b018ae9cd268fcacdbd27`.
  Native verification and prerequisite ancestry are retained in local evidence.

Seventy-three historical origin refs were retired by this agent after individual review,
archive verification, open-PR dependency checks, atomic dry runs and exact-tip leases.
Twelve more (#313, four utility heads, two UUID heads, two declaration heads,
the parser head and two jose heads) were already absent after closure;
deletion actors not established.
Receipts: [12 refs](maintenance/historical-retirement-12.json),
[six more](https://github.com/karamhouse/chartrove/pull/333#issuecomment-6570768946),
[six security refs](maintenance/security-retirement-6.json),
[two crypto refs plus one externally absent](maintenance/crypto-toolchain-retirement-3.json),
[four minimal replacements](maintenance/minimal-retirement-4.json),
[four utility heads observed absent](maintenance/utility-retirement-4.json),
[four retired tooling heads](maintenance/legacy-tooling-retirement-4.json),
[two Chromatic heads](maintenance/chromatic-retirement-2.json),
[one Karma head](maintenance/karma-retirement-1.json),
[two UUID heads observed absent](maintenance/uuid-retirement-2.json),
[two declaration heads observed absent](maintenance/type-package-retirement-2.json),
[three validation-action heads](maintenance/actions-v7-retirement-3.json),
[two generator-tool heads](maintenance/generator-tool-retirement-2.json),
[one deprecated OAuth wrapper head](maintenance/oauth-patch-retirement-1.json),
[five Jasmine proposal heads](maintenance/jasmine-retirement-5.json),
[one paired React patch head](maintenance/react18-retirement-1.json),
[two Compodoc proposal heads](maintenance/compodoc-retirement-2.json),
[three duration/layout refs](maintenance/duration-layout-retirement-3.json),
[one parser head observed absent](maintenance/name-parser-retirement-1.json),
[one Fishery head](maintenance/fishery-retirement-1.json),
[two CLI heads](maintenance/cli-retirement-2.json),
[one graph head](maintenance/graph-retirement-1.json),
[one Zone head](maintenance/zone-retirement-1.json),
[one JWT v5 head](maintenance/jwt-v5-retirement-1.json),
[one GitHub SDK head](maintenance/github-client-retirement-1.json),
[one slice-helper head](maintenance/samber-lo-retirement-1.json),
[one Viper head](maintenance/viper-retirement-1.json),
[one experimental-helper head](maintenance/exp-retirement-1.json),
[one Moment head](maintenance/moment-retirement-1.json),
[one Goja head](maintenance/goja-retirement-1.json),
[one tslib head](maintenance/tslib-retirement-1.json),
[two jose heads observed absent](maintenance/jose-retirement-1.json),
[ng-select plus two approved narrative heads](maintenance/ngselect-and-narrative-retirement.json),
[React 19 tooling head](maintenance/react19-retirement-1.json),
[unused idb head](maintenance/idb-retirement-1.json),
[Node devcontainer feature head](maintenance/node-feature-retirement-1.json).
Do not delete unknown refs, prune legacy reflogs, or broadly reset/clean.

Local artifact storage now has an ignored Go module boundary at
dist/validation/go.mod, so full Go checks discover real application packages
instead of non-buildable archived snapshots. All retained snapshot bytes are
unchanged; full root package inventory still covers every tracked backend/scripts
package and desktop is checked separately.
[Local boundary metadata](maintenance/local-evidence-module-boundary.json).

## Recovered coverage and remaining decisions

**Original primary coverage is recovered and independently verified** from the
pre-task local snapshot after the user granted Full Disk Access. All 349 files
match byte-for-byte, with original times, permissions, ownership, flags and ACLs.
Only macOS-managed provenance differs; both original and new values are retained
without rewriting security metadata. The source archive, manifests, complete
operation logs and checksum receipt are retained beside the recovered reports.
The mixed reports and separate coffinfish evidence also remain. frontend/coverage
stays absent, protected from ordinary Karma writes by #344.
[Recovery method and limits](maintenance/2026-11-02-primary-coverage-recovery.md),
[compact receipt](maintenance/primary-coverage-recovery.json).

The maintainer authorized retirement of the legacy narrative repair prototypes
while preserving archives and records. A verified bundle covers both exact tips,
two unique commits and all 77 unique objects. After #386 published the decision,
fresh dependency and exact-tip gates passed and both origin heads were retired:
summary `62a89d799969b1de8ac208543e89a7da85345a91` and text_field_extract
`38ea0562ea29cbcfaf2460e7fa04f28d0ec0d63e`. Cached legacy aliases remain
held unchanged for separate review. No repair or migration was run.
The old repair saves the original temporary model rather than its recalculation;
do not apply it or infer a safe real-data migration from archive verification.
[Decision and recovery](maintenance/2026-11-03-legacy-narrative-retirement.md).
The maintainer also selected retention of selected upstream expansions; the
specific LDAP/OIDC, delegated-permissions, OCR and Typesense/RAG selections are
pending clarification. Preserve every prototype while that choice is pending.

Native #101/#168 acceptance now has observed chooser and synthetic persistence
evidence above; earlier bridge failures remain retained as historical evidence.
#137 is now a duplicate of still-open #69; the canonical task retains date/Period
bounds and multiplicity, and still needs a compatible representation/migration
decision. [Consolidation evidence](maintenance/date-period-duplicate-137.json). #186 lab-date review found that the edit endpoint replaces raw
Observation payloads, so preservation must be addressed before adding date
editing. No lab correction or migration has been implemented. Optional missing/unparseable patient birth dates now show Not provided in Explore
after #378, with the original numeric getter/calculation unchanged. A separate
dashboard age path is corrected by #379, including refreshes. [Retained boundaries](maintenance/2026-11-02-unresolved-acceptance.md).

## Resume safely

- Requery GitHub #222, issues/PRs, all origin heads, rules and Actions registrations;
  inspect local status/refs/worktrees before writing. Counts above are a snapshot.
- Keep all 16 remaining Renovate PRs paused with stop-updating. Creation approval,
  two concurrent PRs, one per hour and no automerge remain. Unpause only a selected
  candidate, review current heads and require fresh checks. Earlier strict-main
  automatic rebases caused 374 Validation runs; 66 scoped bot runs were canceled.
- Review remaining bugs and compatible dependency proposals in small batches.
  Open work is not all blocked: do not use pending product decisions to falsely
  retire actionable defects or dependency targets. Provider-owned registrations,
  key recovery, delegated access and data migrations need their specific evidence.
- Local permanent evidence is dist/validation/2026-11-01-consolidation/; disposable
  diagnostics are /var/folders/3x/74t3wk415llg_1pxfgfq6tpr0000gn/T/chartrove-resume-_5nn98gn.
  The old /tmp/chartrove-consolidation.deb9lg disappeared at the November 2 restart;
  do not recreate it with different evidence. The restart pointer and reverified
  durable artifacts are retained. Do not remove original archives.
- At final acceptance, independently prove exactly issue #222, zero PRs, main only,
  protections, builds/tests and complete preservation. The count targets remain
  incomplete; the original-primary coverage recovery is verified above.
