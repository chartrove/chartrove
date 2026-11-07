# Observation dates retain their source meanings (#440)

Reproduced on accepted main `b52c080b29c1eb3fae183835b78521644dbee6b5`:
the inherited Observation card labeled `effectiveDateTime` as "Issued on" and
ignored the actual `issued` value. With an effective date of `2020-02-01` and an
issued timestamp of `2020-02-03T10:30:00+05:30`, it showed February 1 as issuance.
With only issuance present, that date was hidden entirely.

The model now retains the source `issued` field, following the existing
DiagnosticReport model convention. The card shows `effectiveDateTime` as
"Observed on" and `issued` as "Issued on", each only when its own value exists.
This reflects the distinction in [FHIR R4 Observation definitions](https://hl7.org/fhir/R4/observation-definitions.html)
between the clinically relevant time and the time the result version became
available. Neither field substitutes for the other.

Source strings, precision and timezone offsets remain unchanged; a source year
or month does not acquire an invented day. No Date conversion, default clock,
record mutation, ingestion, schema, API ownership, source permissions or
credential change. Existing `effective_date` consumers retain their contract.

Synthetic rendered-card regressions cover distinct dates, either date alone,
neither date, year/month precision, Unicode values and raw-resource immutability.
Five of six failed before correction; the unknown-date case already passed.
Exact source and red/green receipts are retained under the existing private
consolidation evidence root in `observation-date-labels-440-20261106`.
Required CI and independent review are recorded on the scoped PR.

This does not add effectivePeriod/Timing rendering or change the Labs date
formatter. It is independent of #186 / PR #439's new-result collection-date
input. #425 operational acceptance remains separately approval-gated.
