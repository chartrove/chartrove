# Canonical aggregation compiler output

After protected#408, full main analysis2091397150 reports one high
[go/sql-injection alert#4](https://github.com/karamhouse/chartrove/security/code-scanning/4)
at the FROM/Table clause. GitHub independently marks original alerts1/2/3 fixed,
with no manual dismissal. Main JS/Actions analyses have zero results within their
qualified coverage. [Exact execution](aggregation-query-ci-validation.json).

The remaining flow follows a caller-supplied Name initially stored in an output
object and returned with an error for unknown fields, forbidden primitive
properties or missing complex properties. All callers check that error and abort
before adding a FROM clause. On successful validation, the name is overwritten
with a server-generated schema key. Runtime tests reject SQL-shaped/unknown
identifiers before execution; no additional cross-owner read is reproduced.
This is a conservative error-state flow, not a claim that#408 still permits the
five proven function/property exploits, which are fixed and regression-tested.

The compiler now keeps the requested name in a separate local variable. It
constructs SearchParameter output only from a matched server schema key and
returns an empty AggregationParameter on every error. Invalid identifiers never
enter compiled output. Error messages, valid columns, bound property paths,
function allowlist, query order/count behavior and caller interfaces remain.
This removes ambiguous invalid state structurally; no finding, source path or
query is excluded, suppressed or dismissed. No schema, key, dependency, native
contract or real record changes. The parameterized#408 builder stays intact.

Five invalid-output cases expose one old pass/four old failures; all five pass
after correction. Nine runtime isolation/binding cases pass both before and
after, distinguishing this hardening from a newly reproduced exploit. Both
root/desktop focused suites pass56events/50leaf cases, including all existing SQL
snapshots and encrypted-key/reopen cases. [Source and log hashes](aggregation-schema-validation.json).
Protected CI and a new full main analysis remain pending;#44 remains open.
No local full frontend/native workflow is admitted below the unchanged50GiB
reserve. No new packaged native qualification is claimed.

The separately tested text-only concept correction is committed at
20a82b097bfa2b8f7b815b4652a8f3c340723488 on the unpublished maintenance-owned
branch task/text-only-concepts-135. Its source, tests and evidence are preserved
for the next independent integration. Both separate retention owner branches/
worktrees remain held, with no cleanup or branch transfer implied.

Raw evidence: dist/validation/2026-11-01-consolidation/post-dependency-issue-review/aggregation-schema-identifiers/.
