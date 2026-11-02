# Validation-only first-party action update review

| Critical | High | Medium | Low | Informational |
| --- | --- | --- | --- | --- |
| 0 | 0 | 0 | 0 | 3 upgrade candidates |

No new trust-boundary vulnerability is identified in this scoped workflow review.
The three maintenance candidates are #284/#285/#286, not retired publishing
workflow changes. Update checkout to 7.0.1, setup-go to 7.0.0 and setup-node to
7.1.0 using full 40-character SHAs. Their proposed SHAs match current official v7
tags and verified GitHub commits; pinned action descriptors, READMEs, package
metadata and credential/cache implementation source are retained with hashes.
[Audit and validation receipt](actions-v7-validation.json).

All three action implementations run on Node 24; the existing hosted Linux/macOS
jobs report runner 2.338.2, above the documented 2.327.1 requirement. The project's
Go 1.27.3 and Node 26.13.0 selections remain explicit and unchanged.

Trust review: ordinary pull_request, main push and manual triggers only; no
privileged PR-target/workflow-run bridge, publishing step, secret or OIDC input,
self-hosted runner, external action or attacker-controlled run interpolation.
Top-level contents:read remains. The workflow/ref interpolation is used only in
the concurrency key. The environment-file value comes from trusted mktemp output.
Checkout keeps persist-credentials:false, confirmed in its pinned input handling
and authentication cleanup. Go keeps cache:false. Setup-node adds explicit
package-manager-cache:false to preserve the previous no-cache behavior against
its new automatic npm-cache default. No registry/mirror credential is configured.

Actionlint exits 0. Parsed YAML comparison proves the only semantic changes are
six pinned uses references and that cache flag. Commands, triggers, permissions,
job names, toolchains, runners, assertions and timeouts match the prior workflow.
Backend/Frontend/Desktop check names and strict main rules are unchanged; fresh
checks using the reviewed actions must pass before integration. Retired upstream
workflows remain disabled and byte-preserved; this does not enable publishing.

The workflow is intentionally excluded by sparse checkout. Its blob is patched
through the Git index; no on-disk workflow directory is materialized. The existing
skip-worktree flag and exact sparse-pattern hash are preserved. Original workflow
bytes remain recoverable from main history, and exact current proposal tips/full
diffs are in a checksum-recorded bundle with native prerequisite and isolated
strict pack/object verification. [Recovery](actions-v7-heads-recovery.json).

No application/dependency/data/Wails contract changes. The current native package
remains applicable. Reconcile the three proposals only after protected integration,
checking exact tips and open dependencies before retiring branches. Original-primary
coverage recovery and pending product/data scope remain unresolved.
