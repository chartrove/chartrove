# Manual Encounter row display

#104 reports empty manual Encounter rows. The real wizard puts the entered
Encounter Service Type in serviceType; the table reads only type[0]. Add serviceType
as a fallback, retaining an existing provider type's priority. The shared Period
formatter also interpreted an omitted endpoint as now (or null/empty as Invalid
date). Show Not provided for absent/null/empty endpoints, preserving valid supplied
formatting and an absent whole Period's empty result. Do not infer ongoing status.

Thirteen targeted cases exercise actual wizard submission, its create conversion,
serialized raw data and component rendering. Nine fail before repair; four controls
pass. All thirteen pass with identical test bytes after repair. A setup-only
capture reset was added before qualified red; earlier diagnostic results remain.
Provider priority, coded service descriptions, raw resource bytes and selection
identity are covered. The full suite passes 500 tests with one existing skip;
production/Storybook pass. Ten unchanged Node checks are reused only after proving
the whole frontend delta from qualified c109657a6 is these three files.
[Validation](encounter-row-validation.json).

This changes presentation only. The existing wizard output (including fields it
omits) is not rewritten, and no FHIR conformance repair is claimed. No reasons,
practitioners, dates or clinical states are invented. Database date/Period bounds,
multiplicity/reindex policy in still-open #69, partial-date precision and handling
of supplied invalid dates remain unchanged. No Go/API/Wails, schema, migration,
dependency or stored-record change is included.

A fresh synthetic fixture imported through actual embedded routes has one Patient
and one manual-shaped Encounter with serviceType/start and no type/end. Outbound
HTTP is blocked and asserts zero attempts. Fresh native packaging/smoke/UI and
protected integration remain pending. #104 stays open.
Raw evidence: `dist/validation/2026-11-01-consolidation/encounter-row-review/`.
