# Import and graph process diagnostics (#442)

A new synthetic probe on accepted backend source
`b52c080b29c1eb3fae183835b78521644dbee6b5` exposed explicit diagnostic output
outside the accepted automatic ORM logging controls. Even with the application's
logger discarded, `storeFileLocally` echoed the upload filename and graph assembly
printed native source/resource identities to the process streams. A selected-report
query printed an unrelated owned result's identity while building the owner graph.

The correction removes those unconditional filename/graph debug prints. Import,
graph construction, normal error returns, authorization, stored provenance and
encrypted audit records retain their existing behavior. No schema, credential,
clinical-resource rewrite, logger configuration or dependency change. Existing
privacy controls and tests remain intact.

The red-before native regression uses a fresh encrypted SQLite database,
independent synthetic credentials, an explicitly discarded application logger,
and captured stdout/standard logging. A generated four-resource FHIR bundle has
unique fullUrl aliases deliberately different from resource IDs, plus a result
sharing the Patient but outside the selected report. Before the correction, the
filename and native identities reached the capture. Afterward, the same test
passes without those markers, while proving original URN preservation, correct
authorized graph resolution, unrelated-result exclusion, unrelated-user denial
and zero outbound HTTP attempts. Global test sinks are restored on cleanup.

Exact red source, failures and green receipts remain under the existing private
consolidation evidence root in `private-import-graph-diagnostics-442-20261106`.
Required CI and independent review are recorded on the scoped PR.

This does not resolve #158's frontend reference handling, replace existing source
audit controls, establish privacy of every other diagnostic path, or verify
gateway/proxy logging. #425 operational acceptance remains separately gated.
