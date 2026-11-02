# Resource-table request feedback (#106)

The shared FHIR table and its fallback template previously displayed the empty
result message while requests were pending, had no HTTP-error feedback and
allowed an older page response to overwrite the current page. This correction
adds visible loading status, grid aria-busy state and generic error feedback
with Retry. Empty-result text appears only after a successful empty response.
Changing pages clears stale rows and unsubscribes the obsolete request; closing
the table also unsubscribes pending work. Existing query parameters, renderers,
selection behavior, pagination size, API routes and database/Wails contracts
remain. No resource contents or server error payloads are added to logs or UI.

## Reproduction and validation

Fourteen DOM/HTTP tests instantiate the actual Encounter and fallback components,
their real datatable module and FastenApiService. Only the external HTTP transport
uses Angular's testing backend. They cover pending, empty, populated and failed
responses, retry, late paging responses and component teardown. Handwritten
synthetic Encounter/RiskAssessment fixtures provide independent expected values;
no provider endpoint or real health record is used.

The first red run stopped after seven cases because unhandled HTTP failures
escaped the component. It is retained as an interrupted diagnostic, including
its browser disconnect, rather than a complete baseline. Wrapping those calls
in fakeAsync with explicit not.toThrow assertions turns the same unhandled
failures into assertion failures without suppressing them. All 14 tests then
fail against unchanged production source. The exact same test bytes pass after
the correction; no test timeout, assertion or coverage threshold is relaxed.

The full ordinary browser command passes 443 specs with the one existing skip.
All ten Node bridge checks, production and static Storybook builds pass. Existing browser-list,
builder deprecation, CommonJS and bundle-budget warnings remain. Fresh
native packaging and protected-main integration are pending. Coverage is
generated under private TMPDIR; original frontend/coverage absence remains.
Generated Storybook documentation is retained and its original absence restored.
[Detailed validation](datatable-loading-validation.json).

The native package will reuse this fresh production output after verifying its
compiler-input manifest against the committed frontend. Original dist ZIP and
asset trees will be preserved/restored with contents, recorded stat fields and
xattrs checked, as in prior batches. Synthetic native validation will use an
explicit disposable data directory with production providers unconfigured.

The existing #104 manual-encounter field problem and other report/editing issues
are not resolved by loading feedback. #106 stays open until protected integration
and individual reconciliation. Raw source snapshots, both red diagnostics,
unchanged green test hash, build logs, compiler-input manifest and preservation
receipts are retained under
dist/validation/2026-11-01-consolidation/datatable-loading-review/ when complete.
