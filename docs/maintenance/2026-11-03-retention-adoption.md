# Artifact retention adoption

The primary integration branch preserves patient changes in 65aa940298 and
merges the original tested retention commit
ed81d1bbede20c0618a01a859a9af887388ca26c at a05da3bdd6. This preserves original
commit reachability, rather than copying the patch or using superseded 0c02f5f2aa.
The separately owned artifact-retention branch/worktree remains untouched.

Independent checks after adoption pass all 14 Node tests and load Karma with
--disable-features=MacAppCodeSignClone enabled before further browser checks.
The protected Frontend workflow already discovers desktop-tests/*.test.mjs,
including the four new storage tests; no workflow edit/enablement was needed.
No full frontend rebuild was run solely for retention. The owner's successful
live DevTools zero-clone and ignoring-SIGTERM descendant probes remain owner
evidence, not independently repeated application acceptance.

The wrapper's 50 GiB full-run reserve remains unchanged. Current storage need not
reach that threshold to finish adoption; existing qualified frontend outputs
will feed native packaging. The three frontend test-tooling changes (Karma and
two new mjs files) do not alter production compiler inputs; qualify that reuse
against the original producer manifest without rewriting it. Native packaging
must restore original outputs and discard its own expanded scratch after use,
retaining ZIP, checksums, logs, input inventories and synthetic records.

Adoption receipts:
dist/validation/2026-11-01-consolidation/resource-card-age-review/retention-adoption/.
Owner cleanup report/patch:
/Users/osama/Documents/Codex/2026-11-03/ho/outputs/cleanup-report.md and
artifact-retention.patch. Reported 47.01 GiB free and 88.94 GiB file footprint
removed reflect APFS sharing, not equivalent physical recovery. Backups, caches,
dependencies and fixtures remain outside this adoption's cleanup scope.

PR #391 review found a real cross-filesystem preservation defect: renaming
documentation from a checkout volume to OS temp fails with EXDEV. The first
merge admission stopped on its unresolved conversation even though all three
checks passed. The finding, blocked admission and red/green receipts remain in
resource-card-age-review/retention-exdev/ and admission-first-blocked.json.

A follow-up on the integration branch keeps the documentation backup inside the
already-owned checkout validation lock. The original ed81d1b commit and owner's
worktree remain untouched. Two identical success/failure regressions fail before
the fix and pass after it on real /dev/shm versus /tmp filesystems in an isolated
existing Linux image (no network, only the test/helper fixture mounted read-only).
All six storage tests pass there; the full macOS Node run passes 14 with those
two filesystem-dependent cases skipped. Linux CI exercises them without skips.
The regression checks original contents, inode, permissions and scratch cleanup.
No browser, production compiler or native application input changed in this
follow-up; the previously qualified native application remains applicable.
