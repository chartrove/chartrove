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

No database, import, record authorization, encryption, provider connection or
dependency changes are included. The component-specific renderer accepted in
#459 remains unchanged. This is a presentation correction, not clinical-provider
or operational-recovery acceptance.
