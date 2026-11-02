# Current unresolved acceptance boundaries

#101 concerns attachment popup/file upload in the manual-entry wizard (linked
upstream PR 330); #168 separately reports the trusted click in the manual-upload
area not opening the macOS picker. They are overlapping reports, not proven
duplicates. The latest task-built app was started with a new explicit unencrypted
synthetic data directory and providers disabled. A minimal synthetic Patient
bundle was prepared. Cua's exact-app control failed twice with
`Sky Computer Use native pipe startup failed`, before any accessibility state or
screenshot was obtained. No picker gesture or import was executed or proven.
Only the exact task-owned process was stopped after verifying its full command;
its synthetic data, fixture, logs and failed-control receipt remain retained.
[Attempt receipt](desktop-picker-attempt.json). Both issues remain open; the ten
passing packaged smoke checks do not substitute for this missing acceptance.
No host security settings, permission prompts or installed control software were
changed to work around the unavailable bridge.

#69/#137 remain unresolved. The shared extractor currently selects the first
scalar date or a Period start/end fallback. Generated model date fields are
*time.Time; the separate SearchParameterDateType bounds/array shape is not stored
by those fields. Returning range JSON would break the existing parser and database
contract. Correct range/array indexing requires a reviewed representation and
migration/reindex decision with synthetic preservation tests; no such migration
or real-record repair is applied or implied by successful date-column portability
checks. These two reports are not closed as duplicates or as already fixed.

The previously pending summary/text_field_extract narrative disposition and
upstream expansion scope decisions also remain unanswered. Coverage recovery and
routine continuation are not answers to these product/data choices. Safe work on
independent dependencies/defects continues; none of these boundaries authorizes
blanket issue closure or retirement of unique unimplemented features.
