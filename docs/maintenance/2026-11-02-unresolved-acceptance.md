# Current unresolved acceptance boundaries

#101 concerns attachment popup/file upload in the manual-entry wizard; #168
separately reports the manual-upload macOS picker gesture. The original bridge
failure was historical and remains retained in the [attempt receipt](desktop-picker-attempt.json).
After restart, the exact retained native app was observed against disposable
synthetic data. Manual Upload opens the chooser and imports three generated
records. The attachment popup/chooser works, but Save originally returned HTTP
500 because generated-file FormData did not reach file extraction. The bounded
multipart correction now has a fresh native package and actual UI save acceptance,
plus read-only proof of the exact persisted attachment bytes. These are separate
reported paths, individually evidenced, awaiting protected integration and
individual closure. [Current acceptance](2026-11-02-native-attachment-upload.md).

Automatic in-place encounter graph refresh required navigation reload during
acceptance; that behavior is not claimed fixed. A synthetic patient without a
birth date displayed NaN Years Old. Those remaining observations and #186's lab
edit raw-payload replacement boundary require separate review. No date editing,
record migration or real-data repair was performed.

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
