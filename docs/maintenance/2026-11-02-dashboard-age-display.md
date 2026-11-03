# Missing dashboard age

Native QA for #378 showed 0 years on a dashboard whose synthetic Patient has no
birthDate. The real dashboard FHIRPath projection returns an empty array for the
absent date. That truthy array becomes undefined after unwrapping, and
Moment(undefined) represents now, inventing zero. Empty/unparseable dates can
instead render NaN years. The widget also retains an earlier age across refreshes.

Unwrap before checking presence, require a valid Moment date, and clear age when
a complete new query result arrives. The template shows Age not provided when no
age can be derived; the valid string 0 years remains visible. The existing
calendar-year diff and patient ordering stay unchanged. No raw Patient, database,
API, dependency, migration or Wails contract changes. Name/gender/vital-sign
refresh behavior and partial-date policy remain outside this repair.

Tests use the real FHIRPath projection and component rendering. Six regressions
fail before repair; valid zero/adult and two existing cases pass. All ten pass
with identical test bytes afterward. The initial inherited mock returned
Observation resources for the Patient query and pre-seeded zero; that diagnostic
is retained but excluded from qualified counts. The corrected mock returns an
empty Patient result before each explicit test projection.

The full ordinary suite passes 487 tests with one existing skip. Ten unchanged
Node checks are reused only after proving the entire frontend input delta from
qualified 3b9e3a55b is these three widget files. Production and static Storybook
builds pass, compiler inputs remain identical, original generated-documentation
absence restores, and coverage stays private. The adversarial unparseable-date
fixture's Moment warning remains visible; no warning/assertion is suppressed.
[Validation receipt](dashboard-age-validation.json).

Fresh native packaging at 9518224f1 passes ZIP CRC, strict signature and all ten
synthetic checks. All 22,287 compiled/reused/native frontend assets match; only
main bundle/map filenames and index differ from qualified 3b9e3a55b. Original
output content/stat fields and all xattrs restore exactly, with zero provenance
changes. [Native receipt](native-9518224f1-validation.json).

Native dashboard UI shows Age not provided and the known synthetic Encounter.
All three raw resources remain byte-identical to the seed; Patient birthDate is
still absent. The existing synthetic session restored without a new manual
sign-in; fresh sign-in is separately covered by native smoke. After closing the
window, ps proves process absence, with no Cua calls after close.

The initial background window was blank. Title click/Raise did not immediately
render, and a body click returned AXError.notImplemented; Tab keyboard focus
then exposed the expected dashboard. The targeted sample and sandbox warning
are retained without claiming a cause or adding an application workaround.
Protected integration is pending.
Raw evidence is retained under
`dist/validation/2026-11-01-consolidation/dashboard-age-review/`.
