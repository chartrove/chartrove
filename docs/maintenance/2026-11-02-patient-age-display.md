# Missing patient birth-date display

Native table QA previously observed NaN Years Old in a patient header whose
synthetic Patient had no birthDate. The existing numeric getter intentionally
returns NaN for missing/unparseable values; the template directly appended its
age suffix. Format that numeric result for display: non-finite values become
Not provided, and finite values retain the same age/suffix, including zero.
The original getter body and approximate 365-day calculation remain byte-identical.
This does not establish precise calendar age or change partial-date policy.

Three new real-component rendering cases fail for absent/null/unparseable dates
before repair. Two valid zero/adult controls and the four existing tests pass.
All nine pass with identical test bytes after the display correction; raw synthetic
resources remain unchanged. The ordinary full suite passes 479 tests with one
existing skip. The unchanged ten Node checks are reused only after proving the
entire frontend input delta is the three profile component/template/spec files.
Production and static Storybook builds pass, compiler inputs remain identical,
coverage uses private TMPDIR and original generated-documentation absence restores.
[Detailed validation](patient-age-validation.json).

Fresh native packaging, synthetic profile UI and protected integration are pending.
The native fixture is a copy of the previously verified three-record synthetic
picker dataset, with no Patient birthDate. No real or purchased records are used.
This changes no Go/API/Wails, database, migration, persisted raw resource, dependency,
other age formatter or clinical interpretation contract. The display correction is
limited to the observed source-profile header.
Raw before/red/green/build receipts remain under
`dist/validation/2026-11-01-consolidation/patient-age-review/`.
