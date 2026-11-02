# Reviewed slice-helper maintenance (#270)

This replacement adopts github.com/samber/lo 1.53.0 in both Go module graphs.
Only one version field and two new checksums change per graph; existing checksum
bytes and order remain. The selected graphs also drop the old release's
go-funk 0.9.1 benchmark requirement and unused niemeyer/pretty requirement. Neither
module appears in either production import list before or after. All other
selected versions, Go/toolchain settings, database/provider replacements and
application source remain unchanged. The target minimum is Go 1.18, below the
application minimum. No experimental SIMD module is adopted or imported.

Production callers use Uniq and Compact to construct query clauses and Uniq
for summary codes. Their ordering, first-occurrence and zero-value behavior is
retained. Existing synthetic query/summary assertions pass with the old release
and with the candidate in both application module graphs. The excluded model
generator uses ForEach; isolated old/new binaries produce the same 57 files,
also matching the prior qualified generator hashes. No tracked model is
regenerated, and no health-record migration or reindexing occurs.

## Upstream test packaging and validation

The published release's root test command fails to compile: its retained
examples refer to a foo helper removed with unit-test files. This failure remains
explicit; it is not reported as a passing release suite. The
[pinned release workflow](https://github.com/samber/lo/blob/56ef3beaf8adfea1908b094e49b3b639ea604aab/.github/workflows/release.yml)
tests the development checkout, removes unit tests and test dependencies, then
creates the tag. The immediate parent is
56ef3beaf8adfea1908b094e49b3b639ea604aab; the authenticated release points to
cf6fb4f9b08c1d3d6e309581316f106dc30b458e.

All 37 production Go/assembly/header source files belonging to the root module
match byte-for-byte across that parent, the release tag and the authenticated Go
module archive. The separate nested exp/simd module is outside that archive and
application dependency closure. The complete upstream parent root suite passes
949 tests with no skips under its separate development module graph. No test
file or assertion is modified, and no missing helper is invented. Application
compatibility is independently checked under the actual selected app graphs.

Initial diagnostics expecting only one graph delta and comparing the separately
published SIMD module against the root module archive failed before applying
the update. Their actual differences and corrected root-module scope are retained.
This changes neither application validation scope nor test thresholds.

Both module verifications, full vet checks, desktop host/encrypted backend tests
under the unchanged 120-second limit, and the server build pass. Applied module
files and selected graphs equal the prevalidated candidates exactly. Fresh native
packaging, protected integration and individual #270 reconciliation are pending.
[Detailed validation](samber-lo-validation.json).

[Exact current and original intake head recovery](samber-lo-head-recovery.json).
The reviewed historical tip a2752392195b70192961475f81ded01764104fe5 has a
verified checksum-recorded incremental bundle and unique-object proof, alongside
the original intake history. Its branch remains live until protected integration
and fresh dependency/expected-tip checks. Source archives, exact failures, test
commands, module inventories and all generator outputs are retained under
dist/validation/2026-11-01-consolidation/samber-lo-review/ after atomic relocation.
