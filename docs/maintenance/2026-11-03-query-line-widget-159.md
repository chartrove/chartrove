# Configured complex line chart (#159)

Issue [#159](https://github.com/karamhouse/chartrove/issues/159), imported from
Fasten upstream #506, reports a configured complex-line-widget continuing to show
Diabetes Tracking demo values despite a successful Observation query. The full
original report and discussion were read and retained; the discussion asks how
configuration was supplied and mentions a GitHub Gist, without a resolution.
Current dashboard code already forwards queries to GridStack. The widget itself
has no widgetConfig input or API path and always renders static demonstration data.

## Corrected behavior

The widget now uses the existing DashboardWidgetComponent query/FHIRPath alias
contract. Each configured query supplies an independent styled series. Configured
xAxisKey/yAxisKey (or existing label/data defaults) resolve selected values without
mutating the configuration or raw records. Exactly one scalar coordinate is
required: missing numeric values remain null gaps, zero stays zero and ambiguous
multiple values are rejected with a generic error. Labels retain their text and
time-zone offsets; category plotting does not silently reinterpret dates/units.
The chart makes no clinical calculation or unit conversion.

Configured title/description replace the fixed clinical metrics and inactive
Day/Week/Month controls. Loading, missing configuration, empty results and errors
have explicit states without fabricated data or exposed response details.
Replacing the input cancels pending work and clears previous data before rendering
a new query; removal cancels requests and destroys the actual chart. The Storybook
example uses an in-memory synthetic API provider, never a production endpoint.
No provider, backend API, database/schema, encryption, Wails or dependency changes.
Project attribution and licenses remain unchanged.

## Evidence and limits

Identical six-case browser controls use the real HTTP service/testing backend and
Chart.js canvas. Original main runs all six: one passes and five fail, including
the missing widgetConfig/query path. The corrected component passes all six with
the same spec hash. Two earlier spec compilation errors are retained; neither is
represented as a runtime product failure.

The final eleven cases pass, adding configuration replacement/cancellation,
missing-value gaps, ambiguous-coordinate rejection, independent multi-query series
alignment/styles and the actual registered dashboard GridStack add/save path.
Assertions inspect rendered point positions for zero/2.5, original timestamps,
unchanged raw JSON/configuration, generic failure handling and chart destruction.
All records are disposable synthetic fixtures; no real records or provider calls.
Application and spec type checks pass. Standalone whole-Storybook type invocations
first encounter TypeScript6 inferred project-root errors, then the intentionally
absent generated documentation.json dependency; their exact failures remain.
The changed story is checked separately using the unchanged Storybook type policy
and explicit frontend root. This is not a full Storybook build.

Focused browser coverage is only a consumer slice, not a whole-app coverage gate.
Its private output was disposed, with logs/hashes/free-byte receipts retained.
The live Angular cache setting remains false and the 50GiB full-workflow reserve
is unchanged. No below-reserve full production/Storybook/native package is run;
protected Backend/Frontend/Desktop checks are still required before integration.
No GitHub Gist transport, new native package or clinical visualization accuracy is
claimed. The fix addresses the demonstrated configured-widget failure, not every
possible external configuration-delivery problem.

[Compact validation](query-line-widget-159-validation.json).
Complete original source, issue/comments, harness failures, control logs and
receipts remain in dist/validation/2026-11-01-consolidation/
post-dependency-issue-review/issue159/. Issue closure remains pending protected
integration and a fresh original-content check.

## Protected integration and issue disposition

PR #403 merged at935504cc61c84e599a6e59abd318d7b1dfa8b423 with all three
required jobs passing in run41745782327:572 browser passes, one existing skip,
16Node tests and6.13MB initial output under the unchanged10MB budget.
Complete job logs and hashes are retained. After a fresh complete-content and
protected-main incorporation check, #159 closed as implemented with comment
6619050399. Only its incorporated task branch was retired; the exact task tip is
reachable from main. Both independent retention owners remain held.
[CI](query-line-widget-159-ci-validation.json),
[retirement](query-line-widget-159-retirement.json),
[fresh checkpoint](2026-11-03-checkpoint-after403.json).
