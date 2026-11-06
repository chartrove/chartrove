# Lab report reference and source identity

This source correction follows [Chartrove #158](https://github.com/karamhouse/chartrove/issues/158), imported from [Fasten #502](https://github.com/fastenhealth/fasten-onprem/issues/502), originally reported by dwradcliffe and discussed with AnalogJ. The original issue and Synthea discussion remain intact. The additional graph-read defect is tracked in [#444](https://github.com/karamhouse/chartrove/issues/444).

## Reproduced behavior

On base `b52c080b29c1eb3fae183835b78521644dbee6b5`, a disposable encrypted import with independently generated synthetic records preserves `DiagnosticReport.result.reference` URNs. The current importer already resolves Bundle fullUrl aliases into relationship metadata. The Labs page instead split the raw URN on `/`, requested an invalid resource type and failed to finish loading.

The aliases deliberately differ from resource business IDs: stripping a URN prefix would select the wrong identity. A second native reproduction imported two distinct patient/source bundles with equal report and result business IDs. Graph placeholders retained their source UUIDs, but expansion omitted that source in the database content query and returned the first source's result for both reports. This is new graph-read evidence, separate from accepted record-mutation and tab-editor corrections.

## Source behavior

- Expand every graph vertex by the native owner UUID, source UUID, resource type and business ID. Content predicates also require a live native actor, a live source owned by that actor and a live resource. Apply the same live predicates to selected roots. Propagate root-query failures.
- Preserve the existing graph API and its owner-only authorization contract. Administrators and delegates do not acquire graph access to other users' records. VIEW/EDIT grants, expiration, revocation and audit behavior remain with the existing explicit shared-source endpoints.
- Load selected lab groups from direct normalized DiagnosticReport-to-Observation graph edges. Never interpret a reference URI as authorization, guess an ID, fetch an external URI or rewrite raw FHIR. Nested derived-from input Observations do not add report groups.
- Verify the returned report's exact source/type/ID and each direct Observation's source. Require normalized result identities to account for the distinct raw references. Deduplicate groups and omit absent LOINC codes instead of inventing an `undefined` code.
- Show a terminal error for unavailable, ambiguous or incomplete reference metadata. A genuinely empty report completes normally. Cancel replaced report/page requests and subscriptions when leaving the page, clearing stale display state on navigation/failure.
- Keep the established subsequent owner-authorized, cross-report historical code-trend query unchanged. A selected report selects trend **codes**; the chart is not restricted to results from that single report or source.

FHIR Bundle fullUrl resolution is described in the [FHIR R4 Bundle reference rules](https://hl7.org/fhir/R4/bundle.html#references). The existing pinned importer supplies this resolution; this change adds no SDK, schema, dependency or configuration migration.

## Validation

Synthetic fixtures cover aliases distinct from business IDs, equal IDs across native source namespaces, zero-valued results, Unicode and unchanged raw URNs. Native encrypted HTTP import/graph tests check correct-source results, unrelated-user denial and no outbound HTTP. Repository lifecycle cases check owner/admin/delegate boundaries, grant/revocation/expiration non-escalation, mismatched native UUID, stale actor identity and resource/source/user deletion. The same lifecycle contract runs against disposable PostgreSQL through the existing tagged local-Docker fixture.

Frontend cases cover exact selected identity, direct-edge membership, deduplication, missing/ambiguous roots, unavailable references, actual API request/response shape, empty/error rendering and request cancellation. Exact red source/logs and focused green receipts are retained privately under the existing consolidation evidence root in `lab-report-references-158-20261106`. Required protected Backend, Frontend, Desktop and CodeQL checks plus exact-head independent review are recorded on the PR.

Targeted commands (use only disposable fixtures):

```sh
npm test -- --watch=false --include='src/app/pages/report-labs/report-labs*.spec.ts' --include='src/app/pages/empty-report-requests.spec.ts'
go test ./backend/pkg/database -run '^(TestGraphSourceIdentitySQLite|TestRepositoryGraphTestSuite)$' -count=1
go test ./backend/pkg/web -run '^TestEmbeddedGraphKeepsSameBusinessIDsInTheirNativeSources$' -count=1
go test -tags postgres_integration ./backend/pkg/database -run '^TestGraphSourceIdentityPostgres$' -count=1
```

The PostgreSQL fixture requires the existing pinned local image and a local Docker Unix socket, refuses remote contexts and uses `--pull=never`; it creates and removes only its own bounded synthetic container.

## Limits and rollback

This does not migrate old missing relationship metadata or implement contained/identifier-only resolution. Different reference aliases resolving to the same result cannot be proven equivalent from this graph response and produce an explicit error rather than a partial chart. This conservative count check is not a new URI resolver. Results without LOINC codes remain outside this existing LOINC trend view.

The existing graph endpoint still constructs the owner's relationship graph before selecting roots; this patch does not redesign its scaling or make graph traversal an atomic database snapshot. Each content read rechecks current ownership/deletion, but it is not a multi-query serializable transaction or a new concurrency guarantee. Accepted #438 optimistic snapshot and non-atomic multi-record save limitations remain unchanged.

Source-only rollback is a normal protected revert of this change. It must not change databases, native identities, grants, credentials or the retained raw medical records. No production connection, recovery, deployment, Authentik operation or real patient-data processing is performed. [#425](https://github.com/karamhouse/chartrove/issues/425) remains a separate operational acceptance gate.
