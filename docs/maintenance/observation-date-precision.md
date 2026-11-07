# Observation table and chart date precision

Observation tables and chart axis labels now retain partially specified
`effectiveDateTime` strings: `2020` stays `2020`, and `2020-02` stays `2020-02`.
Passing those assertions to the previous full-date formatter displayed January 1
or February 1, adding a month/day the record did not supply.

A small shared formatting helper preserves year-only and year-month strings.
Complete dates and timestamps still use the same Angular `mediumDate`, `en-US`
and existing timezone behavior. Missing observation time remains “Unknown date”;
an issuance timestamp does not replace it. This is not new date validation,
timezone conversion, medical inference or a data repair.

Eight synthetic regressions exercise actual table cells and chart-label arrays.
The unchanged implementation produced four partial-precision failures; the four
full-date/missing-time controls passed. The frozen tests also require unchanged
raw resources and effective values. The earlier accepted resource-card
Observed/Issued labeling and date-identity implementation remains unchanged.

No stored records, schema, import, source grouping, clinical values, authorization,
encryption, dependency, provider configuration or production operation changes.
Source-level checks do not establish live clinical-provider/repeat-update or
operational recovery acceptance.
