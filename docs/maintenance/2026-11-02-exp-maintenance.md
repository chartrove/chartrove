# Reviewed experimental collection helper maintenance (#221)

Adopt x/exp v0.0.0-20261014143641-f0251eb4aa3e in both application graphs.
The root previously selected the March 2022 revision; desktop selected April 2026.
Historical #221's companion x/mod 0.42.0 target is already selected in both graphs;
retain it and explicitly record the desktop requirement. x/exp requires the
x/tools 0.52.0 selection; that module is absent from both production import
closures. Two root graph additions are x/tools' deprecated development test helper
modules, also absent from production. The complete graph delta is retained.

Both Go minimums remain 1.26.0, matching all target module requirements. Every
other selected runtime module and all database/provider replacements remain
unchanged. Original checksum bytes/order remain; two root and four desktop lines
are added. No application source, schema, data migration or frontend change occurs.

## Caller, generation and application checks

Chartrove uses maps.Copy and slices.Contains in query, graph and dashboard cache
handling, plus Contains in the excluded FHIR model generator. Seven existing
query/summary assertions pass before and after the update in both module graphs.
All 116 upstream tests for the used maps/slices/constraints packages pass without
skips; this is not a claim about every experimental package in x/exp.

The isolated candidate generator produces the same 57 model files as the earlier
qualified generator, whose source and JSON inputs remain unchanged. Related-version
metadata is also byte-identical across old/new manifest fixtures. No tracked model
or ignored repository metadata is regenerated. Both dashboard-cache boundary
suites pass four events, retaining null-gist rejection and path-escape protection.
Both module verifications, full vet checks, four unlock race tests per graph,
28 desktop tests and the server build pass. Applied selected graphs equal the
prevalidated candidates. [Detailed validation](exp-validation.json).

The desktop baseline import-list command ran after root candidate application.
Its backend replacement then saw the candidate root go.mod and correctly refused
read-only baseline updates. That failed load is retained, not counted as a pass.
Reuse the already qualified Viper snapshot only after proving every selected
baseline version/replacement/Go minimum and all executable source bytes match.
The current candidate import lists are independently captured; x/exp is the only
changed production module. No worktree/source copy or go mod tidy workaround.

Fresh module scans report only the already documented GO-2026-5932 deprecated
OpenPGP finding; that package is absent from both production import closures.
No x/exp-family finding is reported. This is not whole-application security or
call-reachability assurance. All scanner outputs and exact commands are retained.

Fresh native packaging at 48db0aba7 passes ZIP CRC, strict signature and
ten synthetic native checks. All 22,287 qualified frontend files remain identical;
binary metadata confirms the target x/exp revision. Original output content/stat
fields and all xattrs restore exactly, with zero provenance changes.
[Native receipt](native-48db0aba7-validation.json). Protected integration and
individual #221 reconciliation are complete: protected PR #376 merged at
`9ce3494f9e0d0ff20ae6bbe950d1990ff90ac4c3` after all required checks passed
in run 41566730282. Historical #221 is closed as implemented. [Exact current and original intake recovery](exp-head-recovery.json).
The historical tip 88562f34117dbedac29a7df8283ed9f9878f7df5 has a verified
checksum-recorded bundle with all four unique objects independently proved present
in its original archive bytes. Its exact archived head and the integrated task head were manually retired after
fresh dependency, expected-tip, ancestry and bundle checks.
[Retirement receipt](exp-retirement-1.json).
Raw artifacts and private synthetic results remain under
`dist/validation/2026-11-01-consolidation/exp-review/`.
