# Patient detail display and consolidation checkpoint

The native synthetic patient review after #390 exposed a third age presentation
path: the Patient resource detail card printed NaN when birthDate was absent.
The card now shows Not provided for non-finite/missing ages. A malformed birth
date no longer causes DatePipe to abort rendering. Valid dates retain mediumDate
formatting; valid zero and adult ages retain their numeric values. PatientModel,
the shared age calculation, raw FHIR records, database and Wails contracts are
unchanged. This does not resolve #5's allergy-date and ethnicity requirements.

Eight focused cases include absent/null/empty/malformed dates, zero, adult and
refresh controls. The exact final spec reproduced five failures before the fix
and passed all eight afterward (SHA-256
68639fbacba89ed39d5b99c1795cdf12139a9a565cf044c16f275a9f95912a77).
The full ordinary browser suite passed 536 tests with one existing skip;
production and Storybook builds exited zero. Frozen install and ten unchanged
Node cases were reused only after proving their inputs unchanged. Browser logs
retain ERR_STREAM_PREMATURE_CLOSE warnings despite successful test/process exits.
No timeout, assertion or budget was relaxed.

Raw evidence is retained under
dist/validation/2026-11-01-consolidation/resource-card-age-review/. It was moved
intact from the private diagnostic directory; relocation.json explains original
absolute paths in unmodified producer receipts. Native packaging at 33c64fbdb6
reuses those qualified production inputs, with the three adopted test-tooling
paths explicitly accounted for. ZIP CRC, strict signature and all ten synthetic
native checks pass. Original ZIP/frontend/assets content, stat metadata and
xattrs are restored exactly. Expanded native build/smoke output and this run's
frontend/Storybook output were disposed after consumers exited, retaining file
inventories, logs, ZIP and synthetic smoke data. Native patient-card UI was not
repeated; its display behavior is covered by the browser DOM regressions.
[Native receipt](native-33c64fbdb-validation.json).

The separately authorized artifact-retention owner holds task/artifact-retention
in /Users/osama/Code/_worktrees/karamhouse/chartrove/artifact-retention. Preserve
that branch/worktree. Final ed81d1bbede20c0618a01a859a9af887388ca26c was merged
after this clean source checkpoint, before further browser checks. Superseded unpublished
0c02f5f2aa is not the adoption target. No full rebuild solely for that tooling fix,
backup/cache/dependency/fixture pruning, or wait for 50 GiB is authorized by the
cleanup handoff. Future expanded output must use disposable scratch.
