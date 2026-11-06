# Explicit collection dates for new manual lab results

This addresses Chartrove #186, preserving the original report by
[SerialVelocity in Fasten #617](https://github.com/fastenhealth/fasten-onprem/issues/617).
Base: `b52c080b29c1eb3fae183835b78521644dbee6b5` (#438).

The inherited add-lab dialog had no date input and passed the LForms export into
resource creation unchanged. Undated Observations therefore had no
`effectiveDateTime` for the Labs table/chart. A new synthetic browser regression
failed on that missing input before the correction; leaving the date blank
already reproduced the existing "Unknown date" display.

The dialog now offers an optional collection **calendar date** for newly entered
results. There is no default. A valid, explicitly entered `YYYY-MM-DD` is copied
to `effectiveDateTime` only on new Observation/DiagnosticReport entries lacking
any effective choice or associated primitive metadata. Invalid calendar days,
ambiguous formats and timestamps cannot be saved through this field. Changing
the date does not reload the selected panel or discard answers.

[FHIR R4 dateTime](https://hl7.org/fhir/R4/datatypes.html#dateTime) supports day
precision without a time or timezone; [Observation.effective[x]](https://hl7.org/fhir/R4/observation-definitions.html#Observation.effective_x_)
represents the clinically relevant time, commonly specimen collection. Users
must supply their known collection date; the application does not infer it from
the encounter, current clock, report issuance, a clinical code or another record.
The raw calendar day is not converted through a JavaScript Date or an instant.

Existing effective dates/periods/timing/instants and primitive extensions are
preserved, as are issuance, values, units, references, notes and other fields.
The exported input bundle is not mutated. With no supplied date, behavior is
unchanged: unknown remains unknown. Historical undated records are not repaired
or migrated, and this field does not add date/time editing to the existing-value
editor or change display behavior for imported alternative date representations.

No schema, credentials, dependencies, authentication, source ownership,
local/delegated edit eligibility or production configuration changes. The existing
List/contained native import route remains authoritative. The encrypted-backend
synthetic contract checks exact day preservation, date indexing, zero/Unicode
values, owner-only reads and restart, with outbound network blocked.

Regressions cover the actual date input through resource assembly and the Labs
table/chart, unknown dates, Gregorian leap boundaries, invalid input, existing
effective choices, panel readiness, and a real published LForms export.
Original failures and validation receipts are retained privately under the
existing consolidation evidence root in `manual-lab-dates-186-20261106`.
Protected CI and independent review are recorded on the scoped PR; source merge
requires separate acceptance. #425 operational acceptance remains separately gated.
