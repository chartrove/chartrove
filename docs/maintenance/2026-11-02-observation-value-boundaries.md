# Observation values and reference boundaries

The inherited display model and generic table treated integer zero and boolean
false as absence. A pending Observation without a global value could also throw
while deciding whether to display a chart. Quantity/reference presence checks
lost zero, and the bar chart conflated explicit zero upper limits with unknown
limits. FHIR Range labels used strict signs for inclusive bounds.

Use nullish presence checks, preserve explicit primitive values, omit charts
when there is no chartable value, and keep unknown upper reference limits distinct
from zero before extending them to the chart maximum. Retain the existing empty
reference placeholder and explicit Quantity comparators. The synthetic factories
also preserve zero/false/empty-string inputs; their default integer sample changes
from 4.9 to 4. This sample change is test-only.

The [FHIR R4 Observation contract](https://hl7.org/fhir/R4/observation-definitions.html#Observation.value_x_)
allows an absent value and integer/boolean choices. [Primitive types](https://hl7.org/fhir/R4/datatypes.html#primitive)
include false and integral numbers. [Range bounds](https://hl7.org/fhir/R4/datatypes-definitions.html#Range.low)
are inclusive; Quantity comparators retain their explicit semantics. Empty-string
preservation is defensive display behavior, not a claim that an empty FHIR string
is valid.

## Reproduction and preservation

The initial 34-test run entered every test: 29 failed and five controls passed;
the same test bytes then all passed. Four inherited Range assertions were corrected
to the normative inclusive signs before the production edit. Full-suite validation
also exposed an old assertion explicitly replacing a zero upper bound. Correcting
that assertion retains all cases and adds missing-upper cases with a zero lower
bound. Two additional real-chart tests then reproduced loss of unknown high limits
for lower-only ranges at zero and five; they pass unchanged after the extraction
correction. No timeout, coverage threshold or skip changes are part of this slice.

The first failed full suite and the later intermediate production output are
retained as diagnostics. The latter build exited zero but its compiler-input
identity assertion rejected qualification after the two follow-up tests were
added; that output is not used for native packaging. A fresh final input manifest
must match committed frontend bytes before reuse.

The final ordinary browser suite passes 474 tests with the one existing skip.
All ten Node checks, production and static Storybook builds pass. Compiler inputs
remain unchanged, coverage uses private TMPDIR, and generated documentation is
retained with its original absence restored. Fresh native packaging at c8c6e4a9f passes ZIP/signature and ten synthetic
native checks. All 22,287 compiled, reused and native frontend files match.
Original output content/stat fields and all xattrs restore exactly. Protected
PR #373 merged at `8adb31486e8803135495dba6b5809afb8e74a478` after all three
required checks passed in run 41559562042. The integrated task head was retired
with fresh dependency and exact-tip checks. [Detailed validation](observation-values-validation.json). Raw logs, source-before
snapshots, test hashes and receipts are retained at
`dist/validation/2026-11-01-consolidation/observation-values-review/`.

## Limits

This is a frontend display correction. It changes no Go module, schema, extraction,
provider registration, persisted raw resource, database migration, unit conversion,
clinical interpretation or Wails contract. Existing unsupported Observation value
choices and blood-pressure component policy are outside this change. The broader
DiagnosticReport/lab issue #158 remains open: its complete reported failure has
not been independently reproduced or established resolved by these checks.

## Native UI acceptance and tooling incident

A fresh backend/API-generated fixture contains one patient and five Observations;
raw zero and false survive the import, with all external HTTP blocked. The exact
packaged executable, sole process, explicit data copy and private cache were
verified before UI selection. Its table shows integer 0, boolean false and quantity
0.00 1; the false card retains its Value and the registered pending card has no
invented Value row or chart. All six fixture records and original seed files remain.

After closing this verified synthetic process, calling Cua getAXState unexpectedly
relaunched the same executable without test arguments. The reopened window was
closed immediately; terminal process checks prove both PIDs absent. This post-close
attempt did not remain isolated and is retained as an explicit incident. Source
inspection establishes that no-argument startup enables encryption with an empty
key and returns STANDBY before creating a database repository; no unlock/key was
submitted. OpenStorage may nevertheless have read or created default FastenCommunity
lock/signing metadata. No default-profile bytes were inspected or repaired. Do not
make a blanket no-default-profile-access claim. Future close verification must use
process inspection alone, without any Cua method on the closed binding.
