# Preserve text-only coded concepts (#135)

Imported issue[#135](https://github.com/karamhouse/chartrove/issues/135), from
Fasten#428, reports missing derived code content for text-only CodeableConcepts
and the same omission for allergy manifestations. The original discussion later qualifies its apparent date-extraction observation
because the inspected backup predated earlier fixes. This change addresses text content; it does not claim
all Date/Period indexing work (#69) or missing-code-display work (#80) completed.
The complete original issue and discussion remain preserved, including provenance.

## Correction and compatibility

The current shared Goja token extractor falls through when a CodeableConcept has
text and absent, null or empty coding. Four source lines now retain that literal
text in the existing token text member. It synthesizes neither code nor system.
Nonempty codings keep their existing code/system/display preference and concept
text fallback; no extra text row duplicates valid codings. Empty concepts and
empty text still produce no token. Identifiers and primitive false retain their
existing output. Unicode, quotation marks and literal markup remain data.
The text-only token shape was already possible for an empty coding object with
concept text; no schema, token model, DatabaseRepository or Wails contract changes.
Coded filtering is not expanded into a new full-text-search feature.

The shared script reaches the existing generated AllergyIntolerance, Condition,
Observation and Medication models without regenerating or editing their code.
Fresh inserts persist both allergy code and manifestation descriptions, preserve
the complete original raw JSON and survive encrypted cold reopening. No key,
license, attribution, dependency, toolchain, generated-model layout or production
provider integration changes. No historical row is reprocessed or backfilled.
Existing raw/derived update consistency is a separate admission concern recorded
for delegated access; this is not a claim that old materializations are repaired.

This is a correction to current token derivation, independent of the retired
summary/text_field_extract prototypes. Neither archived commit is replayed, and
no narrative generation, old reprocessing loop or migration is revived. Those
archives, their recoverable history and the maintainer's retirement remain intact.

## Validation

The same15ordinary extractor/generated-model cases have seven old passes and
eight old failures, then15corrected passes. A separate real encrypted repository
insert/reopen case fails on the original script and passes after correction.
Each module graph passes54full model-test events plus that encrypted case.
The synthetic raw payload preserves an unknown large decimal, zero, false and
Unicode byte-identically. All SQL pools close before t.TempDir removes fixtures;
mock settings exclusively identify the disposable database and synthetic key.
No real profile, Keychain, purchased Fasten database or family records are opened.

The initial temporary-map immutability assertion also caught the engine's
existing __path__ annotation. That assertion was corrected to validate persisted
raw bytes; its initial failures remain excluded from proof. Ordinary source test
files are used; no undiscovered overlay tests or zero-test exits count as success.
[Exact source/log hashes and outcomes](text-only-concepts-validation.json).

Protected CI and full issue-content reconciliation remain pending at this source
checkpoint. #135 remains open. No local full frontend/native build is admitted
below the unchanged50GiB reserve; no new packaged native qualification is claimed.
Raw evidence: dist/validation/2026-11-01-consolidation/post-dependency-issue-review/issue135/.

Three additional date controls pass both on the original and corrected script.
The [FHIR R4 search definition](https://hl7.org/fhir/R4/allergyintolerance.html#search)
indexes reaction.onset and separately recordedDate. Current extraction preserves
the reaction offset, keeps top-level onsetDateTime byte-identical in raw JSON,
and leaves the onset index absent when no reaction onset is present. No zero date
is produced in those supported cases. These are baseline controls, not a date fix
or a claim about the author's historical backup. No date implementation or
migration is introduced; broader issue#69 remains open.
