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

Fresh native packaging, synthetic smoke/UI and protected integration are pending.
Raw evidence is retained under
`dist/validation/2026-11-01-consolidation/dashboard-age-review/`.
