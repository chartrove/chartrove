# Search runtime dependency maintenance

Replace paused #220 with Goja v0.0.0-20261031103838-cf2ff0ec0b31 in both graphs,
plus its required semver 3.5.0 and regexp2/v2 2.11.0. Go 1.26.0 minimum,
SQLCipher/provider replacements, all application source and database contracts
remain unchanged. All original checksum bytes/order remain, with six new lines
per graph. The exact historical tip and earlier intake are recoverable.
[Recovery](goja-head-recovery.json), [validation](goja-validation.json).

Old/new searches pass 24 test events per graph. A separate program exercises the
actual model factory/extraction against 100 checked-in synthetic R4 fixture
files: 88 supported examples across 32 types match in both graphs. Four Bundle,
three List, four MedicationKnowledge and one ResearchStudy files remain unsupported
as before. Strict canonical JSON preserves array order, values, numeric/boolean
representations after decoding, nulls and all model fields. Raw outputs differ
only in reference-object key order; repeating the old engine also varies. All
five canonical corpus hashes match. No record reindex or migration is performed.

The 57 generated model files and related-version JSON remain identical; generator
source and both JSON inputs match the qualified previous baseline. Applied graph
inventories equal the tested candidates. Production changes are the engine,
required semver and regexp2/v2 replacing the old runtime import. Old regexp2/v1
requirements remain declared but are unimported by both production binaries.
Selected goldmark/pty reductions and dropped chzyer helpers are also outside both
production closures; no production module downgrade is claimed away or hidden.

Upstream Goja has 1,348 pass events and two explicit skips: its always-skipped
nondeterministic profiler example and the separately distributed Test262 corpus.
That corpus is not run here. Semver has 88 passes and no skips. Regexp2 has 3,161
passes and 520 upstream-declared unsupported Rust-corpus skips; complete names and
reasons remain in raw receipts. No test/assertion/timeout is weakened.

Each actual graph passes full model tests (33), full web tests (16), search
isolation races (10), unlock races (four), module verification and full vet.
Desktop host tests pass 28 cases; server build passes. Named test skips are zero
for these application checks. Module scans still report only GO-2026-5932 in
required x/crypto's deprecated OpenPGP code, absent both production import
closures. This is module analysis, not call reachability or a clean whole-app scan.

The entire frontend input manifest matches qualified c109657a6. Reuse its passing
Node/browser/production/Storybook and native UI evidence; hosted Frontend must
still pass on the new PR head. Fresh native packaging, ten-check synthetic smoke
and protected integration are pending. #220 and its head remain open/live.
Raw evidence: `dist/validation/2026-11-01-consolidation/goja-review/`.
