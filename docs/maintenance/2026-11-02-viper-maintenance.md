# Reviewed configuration-library maintenance (#272)

Adopt Viper 1.21.0 in both current application module graphs with its required
fsnotify/afero/gotenv, mapstructure v2, locafero and conc family. The root graph
also advances cast and pflag to versions already selected by the desktop graph.
The desktop records the already-selected go.yaml.in/yaml/v3 3.0.5 dependency.
Both Go 1.26.0 minimums and all database/provider replacements remain unchanged.
All original checksum bytes/order remain; 18 root and 14 desktop lines are added.

The complete selected graph changes are retained, including removal of the old
Viper remote backend's unused cloud/etcd/Consul graph. Six unused transitive module versions
select lower after that graph disappears; none occurs in either production import
closure before or after. No production module is downgraded. Chartrove explicitly
reads YAML and does not expose Viper remote configuration or file-watch APIs.
Its resolved configuration is preserved; no health-record migration, provider
cutover, schema change or frontend update is part of this replacement.

## Configuration and application validation

Four private old/new probes cover defaults, YAML and nested decoding, quoted
strings, zero/false, environment/explicit precedence, JSON-detached settings and
key validation. Canonical effective settings match before/after in both graphs.
The initial probe exposed the old unlock copy's precedence loss; that defect was
separately reproduced and fixed by protected PR #374 before adopting this update.
The candidate configuration/web suites now pass 17 tests in each actual graph,
including both unlock boundary regressions and encrypted desktop checks under the
unchanged 120-second timeout. Executable source matches the applied baseline.

The authenticated upstream module copy passes 139 test events. Its existing
Linux-specific symlink-replacement test skips on macOS; Chartrove does not call
the file-watch API. No test source changes occur. Both module verifications and
full vet checks, four unlock race tests in each graph, 28 desktop tests and the
server build pass. Applied module files and selected graphs equal the prevalidated
candidates. [Detailed validation](viper-validation.json).

Fresh module-level scans use pinned govulncheck 1.9.0 and the advisory database
updated November 2 at 22:01:43Z. Each reports only the previously documented
GO-2026-5932 deprecated x/crypto/OpenPGP finding. That package is absent from both
production import closures. No Viper-family finding is reported; a module scan
is not whole-application security or call-reachability assurance. Two initial
scanner argument/current-directory errors are retained as unqualified diagnostics.
Returned advisory inventory records are not counted as affected findings.

Fresh native packaging at b36c00e74 passes ZIP CRC, strict signature and
ten synthetic native checks. All 22,287 qualified c8c6e4a9f frontend files remain
byte-identical. Binary metadata confirms Viper 1.21.0 and mapstructure v2.4.0.
Original output content/stat fields and all xattrs restore exactly, with zero
provenance changes. [Native receipt](native-b36c00e74-validation.json).
Protected PR #375 merged at `0c55e3d2e349a9679b4bc05abd6e50ab0ca2b0cc`
after all three required checks passed in run 41563444499.
Historical #272 is individually reconciled as implemented.
[Exact current and original intake recovery](viper-head-recovery.json). The current
historical tip 98feaa2d66fe94808c57affbed6dccc3ad4101c6 is preserved in a verified
checksum-recorded bundle with all four unique objects independently proved present
in its original archive bytes. Its exact archived branch and the integrated task head were manually retired
after fresh dependency, expected-tip, ancestry and bundle checks.
[Retirement receipt](viper-retirement-1.json).
Raw artifacts, exact commands/errors, module snapshots, import lists and upstream
sources remain under `dist/validation/2026-11-01-consolidation/viper-review/`.
