# Scalar Observation coded values

Scalar Observation CodeableConcept values keep the existing preference for
supplied narrative text, then the first coding's display. If optional display
text is absent, the formatter now retains that coding's supplied code. An empty
coding array produces no value instead of throwing during clinical display.

The inherited data-absence formatter uses the same fallback, so a supplied
`not-performed` code remains `not-performed (data absent)`. No terminology lookup,
translation, clinical inference, numeric reading or combined coding is generated.
Only the first coding is selected, as before. Original source objects remain
unchanged. Primitive value-string formatting is unchanged.

Six focused regressions cover code-only values and absence reasons, empty arrays,
text/display precedence, multiple codings and existing no-value behavior. Against
unchanged source, three cases failed and three controls passed in the actual
cached Jasmine engine with the project's class-field compiler setting. The
initial adapter run with different class-field semantics is retained separately
and is not project-aligned acceptance evidence. These specs also run in the
existing protected browser suite; the focused check does not replace that suite.

The initial source review also identified a consumer interaction: selecting a
chart from only the first reading could send later categorical codes into a
numeric pipeline. Six further browser regressions demonstrated four failures
(later coded/absent readings selected a bar chart, a later pending reading threw,
and an initial pending reading hid the series) and two positive controls.

Visualization selection now uses capabilities shared across the complete series.
Incompatible or missing readings fall back to the existing escaped table,
retaining record order and every row; compatible quantitative series retain
their chart. The old private helper assertion that permitted a mixed bar series
is updated to the corrected table contract; the new quantitative control still
requires an actual chart. Both new regression files remain unchanged after RED.
This does not add unit conversion, change record grouping or certify clinical
unit equivalence; existing date and reference formatting remain unchanged.

No database, import, record authorization, encryption, provider connection or
dependency changes are included. The component-specific renderer accepted in
#459 remains unchanged. This is a presentation correction, not clinical-provider
or operational-recovery acceptance.
