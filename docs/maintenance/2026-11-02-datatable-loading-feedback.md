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
native packaging and synthetic UI validation pass; protected integration is pending. Coverage is
generated under private TMPDIR; original frontend/coverage absence remains.
Generated Storybook documentation is retained and its original absence restored.
[Detailed validation](datatable-loading-validation.json).

The native package at source 2f0fd5bfc reuses the fresh production output after
verifying its compiler-input manifest against committed frontend files. All
22,287 files match the compiled, reused and native asset trees. ZIP CRC, strict
signature and ten synthetic native checks pass. Original ZIP/asset contents,
recorded stat fields and all xattrs restore exactly, with zero provenance changes.
The ZIP SHA-256 is
45862b6ce184d20dc68626bcd77759ad4f29fc5dc211705d71959fdd2ef7b67b.
[Native receipt](native-2f0fd5bfc-validation.json).

A separate native UI run signs into a copied synthetic fixture, renders its
Encounter row and opens the correct record via row selection. Pending/error
cases are verified by the controlled browser tests; native responses were not
artificially delayed. The three synthetic records remain and the original
synthetic seed is byte-identical. The app closes normally and PID absence is
verified. Its missing-DOB NaN display remains; the blank Encounter cell is for a
fixture without type and does not independently establish #104's disposition.
No real family or purchased Fasten storage is used.

The initial receipt parser expected Node TAP output; the installed spec reporter
uses informational summaries. Only parsing changed. An initial native wrapper
path variable collided with its preservation-manifest function and stopped before
moving outputs; that attempt and original wrapper are retained. The corrected
wrapper and all preservation results are recorded, without rerunning successful
application tests or changing product source. Earlier warning counts remain 138
NG01354 occurrences, matching the qualified baseline.

The existing #104 manual-encounter field problem and other report/editing issues
are not resolved by loading feedback. #106 stays open until protected integration
and individual reconciliation. Raw source snapshots, both red diagnostics,
unchanged green test hash, build logs, compiler-input manifest and preservation
receipts are retained under
dist/validation/2026-11-01-consolidation/datatable-loading-review/.
