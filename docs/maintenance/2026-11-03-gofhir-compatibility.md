# FHIR model compatibility and rejected JSON v2 target

The corrected maintenance candidate advances both modules from gofhir-models
0.0.7 to 0.0.9. It explicitly rejects #262's 0.1.1 target for the current stack;
it does **not** implement that target. The published version inventory has no
newer 0.0.x release. Root/desktop select only this module-version change; all
other selected versions, provider mirrors, exact SQLCipher replacement, Go
1.26.1 minimum and actual Go 1.27.3 toolchain remain unchanged.

## Why 0.1.1 is unsuitable

Compile-only admission passed, which was insufficient. The complete candidate
runtime consumer accepts shorthand gender codes, but its known-field typed
round-trip fails on both R4 and R4B. On Go 1.27.3 a synthetic Patient serializes
as an object containing `"Unknown":{"resourceType":"Patient"}` without a
top-level resourceType; unknown input members also disappear. That cannot be
used as a correct FHIR export. The library's experimental `json:",unknown"`
tag is obsolete under stable JSON v2. The [Go 1.27 release notes](https://go.dev/doc/go1.27)
and [migration guide](https://go.dev/doc/jsonv2-migration) describe the change.
The upstream default-branch tip is still the 0.1.1 release commit
259d7a862268b348a92da03a86721b6b2c4274fa at this review.

No experimental flags, global JSON changes, dependency-cache patches, provider
SDK replacement or private compatibility fork are introduced. No candidate
0.1.1 code was applied to the application manifests or real records. Its entire
isolated graph, desired shorthand regression, output, original issue #53 intake
and discussion, published source metadata and failed controls remain retained.
An initial test probe failed to compile because of an unused import; that result
is disqualified, and the corrected runtime probes are recorded separately.

Issue #53 remains open. The current compatible library still rejects M/m; a
synthetic 0.1.1 import accepting them does not establish an acceptable application
upgrade when its exports are broken. No Cambridge/Epic production connection
is tested, and the clinical shorthand capability is not declared implemented.
Rejecting this particular dependency target does not retire that defect.

## Compatible release evidence

The public API comparison covers 8,812 old and 8,818 new R4/R4B symbols.
Exactly 258 existing structs wrap optional/required bool fields in the new
Boolean type; field names, tags and every other type remain unchanged. The six
new symbols are that type and its two JSON methods in each version package.
Existing callable type signatures match after excluding parameter names.
The initial comparison counted renamed parameters as changes; a later guard
also needed to account for required Boolean fields, as well as pointers/slices.
Both raw comparisons and these qualifications remain preserved.

Actual compile-time assertions check the provider SDK's ResourceInterface for
R4/R4B Patient and Observation, including contained-resource return types.
Runtime cases check standard genders and rejection of invented/invalid codes,
resource identity, false booleans, zero, Unicode, partial birth dates and offset
timestamps. Real encrypted embedded imports of the four standard codes preserve
raw data through close/unlock/sign-in/reopen, including an unknown integer above
2^53, unknown arrays/nulls and false/zero values. Their comparison uses Number
decoding to avoid concealing precision loss in the verifier. Wrong keys are
rejected, encrypted headers remain, HTTP is blocked and uploads leave no residue.

The same 100 repository synthetic fixtures are compared through the actual
typed R4 parser/marshaler: 98 supported canonical outputs match exactly, with the
same two unsupported inputs (bundle/example3.json and diagnosticReport/example2.json).
All original fixture hashes remain unchanged. This is not a claim that every
possible FHIR payload or unsupported input is covered.

Separate copies of the immutable old encrypted synthetic seed cold-open under
the new root and desktop graphs; the old graph also reopens a copy after the
new graph. All retain both records, role, password sign-in, wrong-key rejection
and encrypted header with zero network requests. The original seed's recorded
bytes, size, mode and modification times remain unchanged. No database schema,
migration, generated database model or Wails bridge source changes.

The compatible parser also handles a valid JSON-escaped standard gender code,
`ma\u006ce`, as male; this is separate from the unresolved M/m shorthand.
Its identical final consumer is checked against old and new graphs. The complete
backend/desktop/embedded runs precede that additional focused assertion; final
protected CI must run every test on the published head. No timeout, assertion,
check, production budget or storage admission is weakened.

Applied results are captured in the [compact validation receipt](gofhir-validation.json). Full local
frontend/Storybook/native packaging is not repeated for this Go-only batch.
No new packaged native acceptance or real-record migration is claimed. The last
local qualified native source predates later dependency updates; the unchanged
50 GiB full-build reserve still applies. No storage cleanup is performed.

## Recoverability and disposition

The original #262 head 8fefb92c640012524c0f86aab34078f2afb66d0a has a strictly
verified 1,427-byte bundle, SHA-256
a0e485586db360c300ad305cee55cd855156759979bf917ee7f2a51fe87c8539.
All four unique objects are proved present in the original pack bytes and its
prerequisites remain reachable from main. Original authorship/intake/patch and
the desired but currently unavailable shorthand behavior remain recoverable.
[Original-head recovery](gofhir-head-recovery.json).

Raw evidence and helper sources are retained under
dist/validation/2026-11-01-consolidation/gofhir-review/.
After protected integration, #262 can be individually retired as an incompatible
version proposal, with an explicit explanation that only 0.0.9 maintenance is
adopted and #53 remains unresolved. No original PR or branch is retired before
fresh content/head, consumer, archive, integration and ownership checks.
