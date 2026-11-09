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
| Clinical integrity | Lossless import/export, idempotence, updates, attachments and provenance | Not reassessed |
| Identity and access | Native recovery; synthetic Authentik; distinct family identities; negative/revocation tests | Not reassessed |
| Confidentiality | Encryption/key handling, secure defaults, bounded logs and network behavior | [Source request correction](provider-log-privacy.md) integrated; [dynamic response correction](provider-response-privacy.md) in progress |
| Consent and audit | Scoped consent, attribution, transactional audit and live authorization | Not reassessed |
| Recovery | Independently restored encrypted synthetic data, wrong-key failure, restart and rollback | Not reassessed |
| Application acceptance | Browser, embedded backend and native desktop synthetic workflows | Not reassessed |
| Supply chain and release | Dependency findings triaged; immutable source-bound artifacts and reproducible build evidence | Not reassessed |
| Operations | Documented deployment, health, monitoring, retention, recovery objectives and rollback rehearsal | Not reassessed |

## Initial observations

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

## Next implementation slice: atomic clinical persistence

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

## Current execution checkpoint

- Reclamation PR [#469](https://github.com/chartrove/chartrove/pull/469) merged at
  `add1a4ad2cc6c95545c2fb7ead4eff50095df7c4`, after all seven required checks succeeded on
  `5ef82eb0c9f17e0daecb81d5eb358ef862a7524f` (Validation run 42547082306). Exact-main validation is
  separate and must be checked; this is not a product release.
- Draft #466 shared parser tests and the native SMART race suite were freshly executed on its preserved
  head and passed. Its implementation remains unexposed, Patient/Observation-only, bounded and in memory;
  persistence and repeat updates are not implemented. Current-main qualification and review disposition remain.
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
  controls; the complete models race suite passed. Hosted validation and integration remain pending.
- Standalone factory fallbacks still permit unencrypted SQLite and wildcard HTTP listening. Deployment
  configuration and secure-default acceptance require reassessment; the desktop already uses private
  encrypted storage and no application API TCP listener.
