# Current unresolved acceptance boundaries

#101 concerns attachment popup/file upload in the manual-entry wizard; #168
separately reports the manual-upload macOS picker gesture. The original bridge
failure was historical and remains retained in the [attempt receipt](desktop-picker-attempt.json).
After restart, the exact retained native app was observed against disposable
synthetic data. Manual Upload opens the chooser and imports three generated
records. The attachment popup/chooser works, but Save originally returned HTTP
500 because generated-file FormData did not reach file extraction. The bounded
multipart correction now has a fresh native package and actual UI save acceptance,
plus read-only proof of the exact persisted attachment bytes. These separate
reported paths were individually validated, integrated through protected PR
#363 and closed. [Acceptance evidence](2026-11-02-native-attachment-upload.md).

Automatic in-place encounter graph refresh required navigation reload during
acceptance; that behavior is not claimed fixed. The missing-birth-date NaN
profile and invented dashboard age were subsequently corrected in #378/#379,
with raw records unchanged. #186's lab-edit raw-payload replacement boundary
still requires separate review. No date editing, record migration or real-data
repair was performed.

#69 remains unresolved; #137 was individually closed as its duplicate, without
claiming implementation. The shared extractor currently selects the first
scalar date or a Period start/end fallback. Generated model date fields are
*time.Time; the separate SearchParameterDateType bounds/array shape is not stored
by those fields. Returning range JSON would break the existing parser and database
contract. Correct range/array indexing requires a reviewed representation and
migration/reindex decision with synthetic preservation tests; no such migration
or real-record repair is applied or implied by successful date-column portability
checks. #383 corrects only missing Period endpoints in table presentation and
does not resolve this indexing policy.

The maintainer has now chosen retirement of summary/text_field_extract's legacy
repair with archives/records preserved; publication and exact-tip retirement
checks are recorded separately. Selected upstream expansions remain in product
scope, with the specific areas pending clarification. Preserve all prototypes
until the selections are established. Safe work on independent dependencies and
defects continues; these decisions do not authorize blanket issue closure or
retirement of other unique unimplemented features.
