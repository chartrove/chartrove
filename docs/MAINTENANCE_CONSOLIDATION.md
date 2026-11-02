# Maintenance consolidation checkpoint

Owner: sole active maintenance agent in the existing primary checkout. No new
worktrees or subagents. GitHub and this tracked checkpoint are authoritative.
This is an incomplete consolidation, not a claim that target counts are achieved.

## Verified state

Snapshot: November 2, 2026, after integration #357 and its historical proposal/ref
retirement. [Machine-readable receipt](maintenance/2026-11-02-checkpoint-after357.json).
The receipt was captured on clean main before the next Jasmine tooling branch;
its integration PR/branch is additional until merged and removed. Requery live
counts after every integration.

| Item | Verified result |
| --- | --- |
| Published application main | `6ba97e57c4e0cd7b21848b9368b76c5cb0bec92e` (#357) |
| Open issues | 174, including retained Renovate dashboard #222 |
| Open PRs | 54: 47 paused Renovate proposals and seven historical drafts |
| Origin heads | 64 |
| Default branch | main tracks origin/main with equal OIDs; Jasmine tooling branch is additional |
| Cached legacy refs | 17 nonsymbolic refs plus HEAD, preserved unchanged; not live origin heads |
| Tags / worktrees / stashes | 30 unchanged tags; one existing checkout; no stashes |
| Main protection | Active ruleset 26337029, no bypass actors |
| Required checks | Backend, Frontend, Desktop from Actions app 15368; strict up-to-date |
| PR workflow | PR required, zero external approvals, conversations resolved |
| Force push / deletion | Blocked on main |
| Active Actions registrations | Validation only; inherited Development explicitly disabled again |

The seven remaining drafts are #209 (permissions), #210 (sources-page layout),
#211 (Typesense/RAG), #212 (OCR), #214 (OIDC), #215 (delegated access), and #217
(reserved usernames). Their head/base dependencies remain live. These are not
implemented features or blanket retirement candidates. #284/#285/#286 are now
implemented by #354; all remaining dependency PRs require individual review.

## Latest integrated safeguards and tooling

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

Next reviewed batch: [coherent Jasmine tooling](maintenance/2026-11-02-jasmine-alignment.md).
A real browser reports the existing 4.6.0 engine despite the old separate direct
5.1.2 declaration. Two guards fail before alignment and pass afterward with the
bridge suite (seven total). The active engine is byte-identical; matching types,
reporter, 423 browser tests and production/Storybook pass, and every produced
application file matches the latest native package. Protected CI still gates
integration and the five individual proposal dispositions.

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
have 423 passing tests and one existing skip; production builds retain the 9.92 MB
budget/CommonJS warnings. Local Chrome teardown sometimes required forced
termination after successful assertions; required fresh GitHub checks still passed.
No timeout, test assertion, encryption contract or branch rule was relaxed.

Fresh disposable PostgreSQL checks pass (8.304 seconds), including date-column
startup/persistence/reopen, parameters, credential-free logs and startup cleanup.
Its missing-image preflight and pinned-image pull are preserved. No test container
remained. Callback cancellation fails before repair, passes 200 repetitions after
repair, and passes race checks. Existing encrypted desktop integration also passes.

The latest packaged app source `f30970052` contains the direct OAuth
patch integrated by #357. Build, ZIP CRC, strict extracted signature and all ten real
WebView smoke checks pass. [Artifact, SHA-256 and metadata receipt](maintenance/native-f30970052-validation.json).
Original outputs match contents, stat metadata and all xattrs after restoration.
The prior #353 runtime package remains retained; all required #357 checks passed. Tests use new explicit synthetic storage; direct SMART
is disabled in smoke mode. No purchased-app storage, real family records,
production-provider registration, Medplum cutover or real-data migration was
accessed or introduced.

## Preservation and branch retirement

Sparse settings remain SHA-256
`d48de194029f797d33694f8387ee15180f1531af423402379ef199a9f2ca04bc`;
origin fetch stays main-only. Original Chartrove/FastenCommunity ZIP hashes,
30 tag OIDs and all legacy cached refs match intake. Original handoff receipt
remains `f89ee0c67a049bb3e2b1d698ebd89b3b652c4dcac592547aafd9f1cdc91b50af`.
Existing bundles, editor settings, Finder snapshots, archives and untracked work
are retained. Native builds preserve/restore existing output trees before writing.
Latest restoration matches contents, inodes, permissions, ownership, birth times,
mtimes and all xattrs. The earlier c957 build recorded three OS-managed provenance
attribute changes, with both old/new values preserved; no security-attribute rewrite.

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

Forty-three historical origin refs were retired by this agent after individual review,
archive verification, open-PR dependency checks, atomic dry runs and exact-tip leases.
Nine more (#313, four utility heads, two UUID heads and two declaration heads) were already absent after closure;
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
[one deprecated OAuth wrapper head](maintenance/oauth-patch-retirement-1.json).
Do not delete unknown refs, prune legacy reflogs, or broadly reset/clean.

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

Two product/data questions remain unanswered: archive-only retirement versus a
corrected opt-in legacy narrative repair; and retirement versus retention of
unimplemented upstream feature expansions. Keep live summary
`62a89d799969b1de8ac208543e89a7da85345a91` and text_field_extract
`38ea0562ea29cbcfaf2460e7fa04f28d0ec0d63e` until that disposition is established.
The old repair saves the original temporary model rather than its recalculation;
do not apply it or infer a safe real-data migration from archive verification.
A continuation message is not an answer to those pending decisions.

Native picker/attachment acceptance #101/#168 is still unverified because the
exact-app UI-control bridge failed before observation; the isolated process was
stopped and all synthetic evidence retained. #69/#137 range indexing still needs
a compatible representation/migration decision. [Scope and retained boundaries](maintenance/2026-11-02-unresolved-acceptance.md).

## Resume safely

- Requery GitHub #222, issues/PRs, all origin heads, rules and Actions registrations;
  inspect local status/refs/worktrees before writing. Counts above are a snapshot.
- Keep all 47 remaining Renovate PRs paused with stop-updating. Creation approval,
  two concurrent PRs, one per hour and no automerge remain. Unpause only a selected
  candidate, review current heads and require fresh checks. Earlier strict-main
  automatic rebases caused 374 Validation runs; 66 scoped bot runs were canceled.
- Review remaining bugs and compatible dependency proposals in small batches.
  Open work is not all blocked: do not use pending product decisions to falsely
  retire actionable defects or dependency targets. Provider-owned registrations,
  key recovery, delegated access and data migrations need their specific evidence.
- Local permanent evidence is dist/validation/2026-11-01-consolidation/; disposable
  diagnostics are /tmp/chartrove-consolidation.deb9lg. Do not remove original archives.
- At final acceptance, independently prove exactly issue #222, zero PRs, main only,
  protections, builds/tests and complete preservation. The count targets remain
  incomplete; the original-primary coverage recovery is verified above.
