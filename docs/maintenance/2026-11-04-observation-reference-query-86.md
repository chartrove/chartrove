# Observation reference queries (#86)

The original Fasten #306 report, imported as [#86](https://github.com/karamhouse/chartrove/issues/86),
names comboValueQuantity, encounter, subject and valueQuantity subpaths. At accepted
main `43fed3855683448176a0298c8976974f9a042fe3`, actual generated-model extraction
of those fields already works. Both module graphs also pass quantity queries,
including a zero-valued component. The [baseline receipt](indexing-85-86-validation.json)
retains those observations separately from the following new implementation.

Actual subject/encounter queries failed because the query engine explicitly
rejected every reference search parameter value and never expanded its JSON
array. The correction adds exact matching against each stored `reference` member
through the existing owner-scoped, bound-parameter query path. Each condition
uses an independent existence predicate, allowing repeated AND conditions to
match different entries in the same array without duplicating resource rows. Literal
reference strings remain unchanged; there is no network resolution or fetch.

Unsupported modifiers (including identifier/type qualification) and blank values
are rejected explicitly. Modifier rejection stays in query-value parsing so the
existing aggregation JSON-property contract is preserved. No aggregation compiler,
owner predicate, schema, model generator, raw record or derived index is weakened
or rewritten. Exact matching does not normalize absolute/relative/URN references,
implement chaining, or claim complete FHIR-server search conformance. Internal
reference translation and IPS mapping remain separately unresolved in #138.

Ordinary encrypted two-owner Observation fixtures cover subject and encounter
matches, no match, OR/AND combinations, multiple fields, array deduplication,
quantity subpaths, unsupported modifiers, empty/blank values and encrypted cold
reopen. No other owner's resource is returned. All records and credentials are
synthetic, and each SQL pool closes before its private test directory is removed.

The identical corrected baseline has eight supported-query failures and three
passing unsupported-query controls. Final root and desktop selections each pass
38 leaf cases, including all 14 reference controls and existing query/aggregation
regressions. An initial candidate test incorrectly assumed a concrete pointer
slice; its decoding failures are retained and excluded from product proof. The
corrected test consumes the same JSON representation as normal callers. Exact
source/log hashes and retained qualifications are in [the receipt](reference-query-86-validation.json).

Protected CI and integration remain pending. #86 remains open until acceptance;
no existing historical row is reprocessed. PostgreSQL query execution and a new
packaged native UI are not claimed. No local full build was run below the
unchanged 50 GiB reserve.

Final review reproduced an additional array-AND defect in the first published
candidate: requiring both common and specific references returned zero rows
instead of one. A correlated existence predicate per condition corrects it; the
new regression passes in both graphs. The prior candidate's seven green checks
are retained as historical evidence only. The source change requires fresh
protected checks and is not accepted on the preceding head's results.
