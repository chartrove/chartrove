# Observation component values

The Observation card displays each supplied `component` code and value in source
order, alongside the existing parent observation. This includes separate pressure
measurements; it does not derive a combined reading or interpret clinical meaning.
Repeated component codes remain separate. Each component uses its own supplied
reference range, never the parent's range.

The card reuses existing Quantity, integer, boolean, string, CodeableConcept and
data-absence formatters. Zero, false and explicit empty strings remain present.
Angular text bindings escape supplied labels and values. A missing or unsupported
value displays “Value unavailable”; it is not replaced with a numeric reading.

Existing formatter limits remain: only the first supplied reference range is
displayed; Range, Ratio, SampledData, time, dateTime and Period value variants have
no newly implemented formatter. Components are not new standalone resources,
chart series, search entries or inferred DiagnosticReport members. Parent values,
dates, record/source links and the original raw resource remain unchanged.

Seven synthetic browser regressions cover paired quantities, primitive boundaries,
Unicode/escaped coded text, absence reasons, repeated codes with distinct ranges
and units, parent/provenance preservation, and scalar-only behavior. The frozen
tests first demonstrated six missing-component failures against the unchanged
card, with the scalar-only control passing. Existing unrelated skipped tests are
not component acceptance evidence.

No import, persistence, schema, authorization, encryption or provider connection
changes are included. This does not establish Epic clinical import/repeat-update
acceptance or operational authentication/recovery acceptance under #425.

Reference: [FHIR R4 Observation components](https://hl7.org/fhir/R4/observation-definitions.html#Observation.component).
