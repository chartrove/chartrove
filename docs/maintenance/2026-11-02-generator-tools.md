# Generator-only Go tools

This current-graph replacement reviews #258 (Jennifer 1.7.1) and #276 (x/mod
0.42.0). Their complete historical diffs contain only root go.mod/go.sum changes.
[Exact tips and verified recovery bundle](generator-tool-heads-recovery.json).
The proposals remain open until this replacement passes protected CI and merges.

Jennifer is imported only by the build-excluded database model generator; x/mod
is imported only by scripts/related_versions. Both executable dependency graphs
exclude both tools. Fresh before/after package inventories match all runtime
package/module identities and selected source/embed filenames: 662 server and
699 desktop packages. The only two selected module-version changes in either
graph are Jennifer 1.6.1 to 1.7.1 and x/mod 0.41.0 to 0.42.0. Desktop needs only
the Jennifer requirement/checksums; x/mod follows the local backend requirement.
No unused explicit desktop x/mod requirement is added.

The [complete Jennifer comparison](https://github.com/dave/jennifer/compare/v1.6.1...v1.7.1)
adds clear/min/max renderers, updates reserved names and standard-library import
hints, replaces deprecated ioutil.WriteFile with os.WriteFile, and fixes a comment.
The [complete x/mod comparison](https://github.com/golang/mod/compare/v0.41.0...v0.42.0)
changes module metadata and certificate-transparency test data/tests; no runtime
Go implementation file changes. Existing Go directives and all provider/backend/
SQLCipher replacements remain byte-identical.

Unchanged generators were compiled against both root graphs and run in separate
new directories using only public FHIR metadata and disposable module input.
All 57 generated model files and the related-version JSON are byte-identical.
Tracked models and the existing related_versions.json were not overwritten.
This does not run a data migration. Root and desktop module verification, normal
application vet, executable builds, desktop host tests and encrypted embedded
backend checks pass with the unchanged 120-second limit.

Upstream x/mod modfile/module/semver tests pass. Ordinary upstream Jennifer tests
fail vet on eleven invalid example identifiers in both old and proposed versions,
with identical diagnostics under Go 1.27.3. Those failures and exact exit codes
are retained; no vet-disabled run or relaxed application check is substituted.
The actual Chartrove generators compile and output-equivalence checks pass.
This upstream limitation remains explicit rather than being reported as a test pass.

[Validation and evidence hashes](generator-tools-validation.json). Public-source
comparisons, full runtime/module inventories, old/new generator output hashes,
original module inputs and complete operation logs are retained under
`dist/validation/2026-11-01-consolidation/generator-tools-review`.
Application Go/Angular source, frontend selections, database contracts and Wails
bindings are unchanged. Original application build outputs were not touched.
The current packaged application evidence remains the #353 runtime build; no new
native smoke run is claimed for these excluded development tools. Required fresh
Backend, Frontend and Desktop CI still gate this integration.
