# Repository reclamation — 2026-11-09

This records repository custody, not product or deployment acceptance. The new autonomous Pi assignment
supersedes the previous Chartrove stop only; no prior agent or queued task was resumed.

## Initial state and preservation

The historical primary checkout was `/Users/osama/Code/karamhouse/chartrove`, clean on `main`
`91509eb7e3a23f5683f562d58b5dad1879cd1dbb`. It was five commits behind remote main
`b3144a99a84e7e876bd4e964fc8201d11d5dfa4d`. No stash existed. No tracked repository `AGENTS.md`
was present. Shared owner instructions apply. The sparse checkout initially hid `.github`; the maintained
workflow/editor directories were materialized before reviewing them.

All three registered linked worktrees were inspected, including tracked, staged, nonignored untracked,
ignored files, current commit ancestry and remote branch state:

| Linked worktree under the historical shared root | Head | Disposition |
| --- | --- | --- |
| `atlantic-blue-branding` | `d29e5d7e1fd7f3863fd4fd6a79504aba4773f501` | Integrated in main; branch retained |
| `atlantic-blue-theme` | `9c60d5878205cf18639ea87c8bc4c374dee3609d` | Integrated in main; branch retained |
| `atlantic-blue-main-build` | `b3144a99a84e7e876bd4e964fc8201d11d5dfa4d` | Detached exact-main build; reachable through main |

Each was tracked-clean with no nonignored untracked files and no commits unique against `origin/main`.
Ignored data was not assumed disposable. Every ignored path, including dependencies, packaged builds,
frontend assets and generated metadata, was moved without deletion into ignored private
`dist/reclamation-2026-11-09/` inside the primary checkout. Inventories verified **157143**, **157140**
and **157138** entries respectively, including SHA-256 for regular files, symlink targets, modes, sizes,
mtimes and inodes. Finder rewrote/recreated `.DS_Store` metadata during this operation; those disposable
metadata changes are excepted from equality, not treated as lost application work.

Full `lsof` and working-directory/process observations showed no prior coding agent, build, test or app
using the linked worktrees and no competing repository writer. Existing Codex app servers elsewhere
were left untouched. The primary's active processes were this Pi session and its helpers; the SonarQube
mount was read-only. Docker held read-only directory handles, not writable source files.

Normal `git worktree remove` removed all three registrations without force. Two removals returned
`Directory not empty` after unregistering their worktrees; remaining files were preserved by moving those
unregistered residue directories into the same ignored preservation area. No hard reset, clean, force
removal or history rewrite was used. Branches were not deleted.

For additional abandoned-history recovery, `git fsck --connectivity-only --no-reflogs --unreachable`
completed with exit 0. All 39,090 reported unreachable objects were packed privately and the pack was
indexed and verified. All 1,895 unreachable commits additionally received local
`refs/reclamation/20261109/<object-id>` refs. These refs/objects are **not published**. Existing archive
refs remain intact. Original object inventory, pack, hashes, refs and surviving reflogs are retained in
that ignored private directory. This conservative preservation includes historical objects, not a claim
that every object belonged to a removed worktree.

## Outstanding source recovered

- Draft [#466](https://github.com/chartrove/chartrove/pull/466) head
  `f5cb683a1315fe932045812aff0e5f9596bf0d7f`: restored local
  `task/direct-smart-observation-reader` at the exact published head. Prior closeout records unfinished
  clinical persistence/repeat updates and incomplete review. All seven GitHub checks eventually passed
  that exact head; it still requires current-source review and acceptance. No result is transferred to a
  later head. The original test-first and guard commits remain reachable.
- Draft [#431](https://github.com/chartrove/chartrove/pull/431) head
  `e106e80f199c524c1d6a2461863a3971736051ef`: fetched into its existing-named local documentation branch.
  It remains a proposal, not adopted architecture or authorization for prior agents.

## Canonical checkout verification

The primary was fast-forwarded while clean, then moved without copying to
`/Users/osama/Code/chartrove/chartrove`. `git worktree repair` and the resulting porcelain inventory show
**one checkout only**, on `task/repository-reclamation`. Origin fetch/push is
`https://github.com/chartrove/chartrove.git`; `gh repo set-default --view` returns `chartrove/chartrove`.
Remote HEAD and `refs/heads/main` both resolve to `b3144a99a84e7e876bd4e964fc8201d11d5dfa4d`.
The separate `legacy-fork` remote remains historical Fasten evidence, not a Chartrove publication target.

The active GitHub ruleset **26337029**, `Protected main - PR workflow`, remains enforced: PRs, resolved
threads, strict up-to-date seven-check validation, no deletion or non-fast-forward. The legacy branch
protection endpoint's 404 did not mean protections were absent. Repository visibility is public; private
preservation artifacts must never be force-added, uploaded or included in build contexts.

Active README, contribution/security/branding/provider guides and application support links now target
the canonical repository. Dated receipts and attribution retain original names/URLs. The ignored local
Serena project name was repaired from an unrelated historical label to `chartrove`; no dependencies or
global harness configuration were changed.

Rex's supported OSC 7 directory report was rebound to the canonical path and read back as
`file://localhost/Users/osama/Code/chartrove/chartrove`. Pi's already-created native bash tool still captures
the launch path and initially refused commands after relocation. A clearly labelled nonrepository
`RELOCATED.txt` anchor at the historical path permits this session's tools to start; every subsequent
command explicitly changes to the canonical checkout. It is not a duplicate clone or symlink. Remove
that anchor once this Pi session can resume from the canonical directory; do not mistake it for an active
checkout. No process was restarted and no old agent was reactivated.

## Validation boundary

Local validation: `git diff --check` passed; `make prepare-backend` passed; Node desktop checks reported
29 tests, 27 pass, 0 fail and 2 explicitly skipped cross-filesystem cases (no alternate filesystem fixture).
`sh scripts/test-container-context.sh` passed both original/current synthetic privacy controls. Raw local
logs are retained privately beside reclamation evidence; full hosted source validation is required before
merge.

This change updates repository identity, custody evidence and the active readiness plan only. Clinical
behavior, Authentik and deployment acceptance are unchanged. See the [active plan](production-readiness.md)
for remaining production gates; green validation on this change cannot establish them.
