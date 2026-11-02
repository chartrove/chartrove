# Maintenance consolidation

Owner: sole active maintenance agent, existing primary checkout. Started 2026-11-01.
Authoritative delivery is GitHub; this file is the restart checkpoint. The dated
baseline and backlog review retain historical observations, not current counts.

## Current checkpoint

- Start main: `2d4c2c197fea27c46192bcb320a6c5a652da4020`; clean, one worktree.
- Read merged PRs #323/#324 and both maintenance documents. Do not repeat the
  completed Angular 14–22 migration or restore obsolete upstream implementations.
- Live starting inventory: 196 issues, 101 open PRs; remote heads are recorded
  exactly in [the inventory](maintenance/2026-11-01-inventory.json). Fetch config
  intentionally retrieves only main; local tracking refs undercount remote branches.
- Main ruleset `26337029` is active: PR required, zero approvals, conversations
  resolved, no force push/deletion, no bypass actors. Verified effective rules API.
- First integration branch: `task/protected-maintenance-ci`. Local backend/desktop validation passed; frontend validation running.
- All six inherited workflow IDs were individually disabled by API.
- Actions remain disabled until the inherited workflows are individually disabled
  and this replacement is merged. Required checks follow actual successful runs.

## First integration batch

Retire all six inherited workflows from execution. Preserve their exact source in
`docs/retired-workflows/*.disabled` and main history. Add read-only validation on
GitHub-hosted Linux/macOS, fixed toolchains, SHA-pinned first-party actions, no
persisted checkout token, no secrets, no cloud/container/CDN/release publishing.
Run backend tests/build/vet, Angular tests/build/bridge tests, and macOS host plus
encrypted embedded-backend tests/build. Native UI acceptance remains separate.

Renovate requires dashboard approval, at most two concurrent PRs and one per hour,
with automerge disabled. #222 stays open. Approval is a sustainable dependency
intake policy, not proof that outstanding updates are implemented.

## Preservation

Sparse-checkout hash: `d48de194029f797d33694f8387ee15180f1531af423402379ef199a9f2ca04bc`.
Preservation receipt `dist/validation/coffinfish-c42643e2e-preservation.json` verified:
`f89ee0c67a049bb3e2b1d698ebd89b3b652c4dcac592547aafd9f1cdc91b50af`.
Keep retained ZIPs, evidence, local metadata, tags, exclusions, and settings intact.
No health-record directories are read or used. Tests use private synthetic fixtures.
No historical branches have been deleted in this consolidation yet.

`summary` tip `62a89d799969b1de8ac208543e89a7da85345a91` and
`text_field_extract` tip `38ea0562ea29cbcfaf2460e7fa04f28d0ec0d63e` still exist
remotely and require source/reachability comparison. The historical draft PRs
have base-branch dependencies recorded in the inventory. Locate and verify old
bundles, then archive any additional unique tips with checksums before retirement.

## Remaining execution

1. Complete and merge safe CI batch through protected main; enable only validation,
   prove runs, then enforce reliable checks with no bypass.
2. Reconcile each open PR against current source and original scope; close proven
   superseded dependency targets with evidence. Preserve feature PRs until reviewed.
3. Review each issue body against the dated classification. External provider
   access and product/data decisions are blockers, not implemented features.
4. Review summary/text extraction, verify recoverable archives, then retire only
   branches with documented dispositions and exact-tip leases.
5. Requery all counts, builds/tests, main rules, tracking relationship and branch
   inventory. Target 1 issue (#222), 0 PRs, main only is not yet achieved.

Potential decision boundaries: historical delegated access/OIDC/user permissions,
OCR/Typesense feature scope, provider-owned registrations, and unresolved account
key recovery. Do not merge or falsely close these to meet a count. No Medplum
cutover or real-data migration belongs to this task.

## Validation and environment findings

Full backend suite, both module verifications and vet checks, desktop host tests,
and encrypted embedded-backend integration pass (20.457 seconds; 120-second
limit unchanged). The first frontend attempt failed before tests because local
node_modules held Angular 14/TypeScript 4.6 despite Angular 22/TypeScript 6 in the
committed manifests. The old tree is preserved at
`dist/maintenance-node_modules-pre-angular22-20261101`; a fresh frozen install
passes on Node 26.13.0 without manifest/lockfile changes. Existing LForms and
legacy tooling peer warnings remain. No compatibility requirement was relaxed.

Existing historical bundle independently verified by native Git and SHA-256:
`/Users/osama/Library/Application Support/Pylonary/agent-tasks/legacy-worktree-retirement-20261026/recovery/retirement/historical-chartrove-20261031.bundle`
with checksum `03ffd099a7b71a3fa14ffa16982d423b13d79e5463297709c01392ca84535325`.
It preserves beta/export_summary/manual-input/summary/text_field_extract. Four
prerequisites remain required; do not describe it as a self-contained bundle.
