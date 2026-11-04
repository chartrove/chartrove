# Aggregation query isolation

The first full main CodeQL Go analysis2091216033 on07ea27fbdbcdc2596074596e9050e5c5243b64ab
reports high go/sql-injection alerts1/2/3 at GROUP BY, ORDER BY and SELECT.
The original ProcessAggregationParameter checks schema membership and modifier
presence but interpolates arbitrary function and property text into raw SQL.
Five read-only payloads on disposable two-owner SQLite rows return the other
synthetic account's placeholder secret despite the outer user_id predicate.
This is a confirmed query-isolation defect, not a false positive or evidence
that any real installation has been exploited. No real records are inspected.

## Corrected construction

Aggregation expressions now use GORM clause expressions throughout SELECT,
GROUP BY and ORDER BY. JSON paths are bound values, including quotes, question
marks and nested property syntax; they cannot introduce a subquery. Schema
identifiers resolve from the server-generated lookup and are dialect-quoted.
Functions compile to application-owned literals: count, min, max, sum, avg,
total, group_concat, date, datetime, time, julianday and unixepoch, without case
sensitivity. Unsupported function syntax, including load_extension, is rejected.
A star order field permits only count. This narrows the former arbitrary-SQL
function input to supported one-argument aggregates/date grouping; it is not a
general SQL execution interface. Existing source consumers use count and max.

Count-by expansion works on copies so a saved caller query remains reusable.
Field validation, ownership predicate, filtering, date-descending ordering,
nondate-ascending ordering, token concatenation, count labels/values, limits and
offsets remain. Internal SELECT column ordering is now label/value; JSON keys
are unchanged. SQL snapshots explicitly assert quoted columns and ordered bound
paths, rather than accepting interpolated text. The unsafe exported test helper
is replaced by an internal expression compiler; no DatabaseRepository, HTTP,
Wails, database schema, migration, key, dependency or raw-record contract changes.

## Validation and limits

Eighteen identical ordinary Go consumer cases run against the original and the
corrected builder. The original has nine passing and nine failing leaf cases:
seven SQL-syntax admission/isolation failures and two count-by input mutation
failures. Five earlier read-only controls explicitly show cross-owner synthetic
account values; the marker is disposable test data. Added dry-run assertions
also prove malicious property text stays in bound variables and outside SQL.
The final root/desktop suites include these cases, all existing18SQL snapshots,
parameter validation and existing seven SQLite/encryption cases. Their exact
executed counts, exits and source/log hashes are recorded in
[the validation receipt](aggregation-query-validation.json).

Positive cases cover counts of all rows and tokens, max/sum with zero values,
nested paths, date groups, default/date/key/quantity ordering, unchanged saved
query input and another owner's excluded999value. Synthetic fixtures live under
t.TempDir and SQL pools close before directory disposal. Initial fixture errors
and failed intermediate build/assertion attempts remain in raw evidence; they
are excluded from acceptance. The final fixture suppresses only model ID hooks
for explicit synthetic owner IDs; production hooks and schema remain unchanged.

Protected CI and a fresh full main CodeQL baseline remain required. No alert is
dismissed or query suppressed; #44 remains open at this source checkpoint. The
existing seven required validations and strict main rules are unchanged. No
local full frontend/native build is admitted below the50GiB reserve, and no
current packaged macOS qualification is claimed.

Raw evidence: dist/validation/2026-11-01-consolidation/post-dependency-issue-review/aggregation-security/.
