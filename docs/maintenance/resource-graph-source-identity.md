# Resource graph source identity

This correction fixes [Chartrove #444](https://github.com/karamhouse/chartrove/issues/444), independently reproduced while investigating [#158](https://github.com/karamhouse/chartrove/issues/158). The latter retains dwradcliffe's original report and AnalogJ's discussion from [Fasten #502](https://github.com/fastenhealth/fasten-onprem/issues/502).

## Reproduction and correction

On base `b52c080b29c1eb3fae183835b78521644dbee6b5`, import two disposable encrypted bundles under one native owner, with distinct Patient/source UUIDs but equal DiagnosticReport and Observation business IDs. Selecting each report by its source/type/ID returns the first source's Observation for both reports. Graph placeholders contain source UUIDs; expansion lost that source when fetching content by type/ID alone.

Every expansion content query now uses the native owner UUID, source UUID, resource type and business ID. Content predicates require a live native actor, a live source owned by that actor and an undeleted resource. Selected roots use the same live predicates, and root query errors propagate. Table names continue to use the existing resource-type allowlist.

The graph API and generic relationship/reciprocal semantics remain unchanged. A reciprocal Observation.focus link can appear beside actual report results; graph adjacency does **not** establish DiagnosticReport.result membership. The graph remains owner-only. Administrators and delegates gain no implicit access, and explicit source-level VIEW/EDIT, revocation, expiration and audit controls remain on their existing shared-source endpoints.

Raw FHIR, Bundle URNs, native identities, encryption and source provenance remain unchanged. No schema, dependency, SDK, configuration, authentication or frontend change is included.

## Validation

The native encrypted two-source/equal-ID regression failed before the correction. Synthetic aliases deliberately differ from business IDs. Final native coverage checks the exact source and value for direct results and reciprocal focus links, unchanged raw URNs, unrelated-user denial and zero outbound HTTP. The fixture also includes a derived-from input.

Repository lifecycle tests exercise encrypted SQLite and the same contract on disposable PostgreSQL: owner/admin/delegate boundaries, active/revoked/expired grants never broadening the owner-only graph, mismatched native UUID, stale actor identity and resource/source/user deletion. The existing graph suite remains intact.

```sh
go test ./backend/pkg/database ./backend/pkg/web -run '^(TestGraphSourceIdentitySQLite|TestRepositoryGraphTestSuite|TestEmbeddedGraphKeepsSameBusinessIDsInTheirNativeSources)$' -count=1
go test -tags postgres_integration ./backend/pkg/database -run '^TestGraphSourceIdentityPostgres$' -count=1
```

The existing PostgreSQL fixture refuses remote Docker contexts, requires the already pinned local image, uses `--pull=never` and removes only its own bounded synthetic container. Required protected Backend, Frontend, Desktop and CodeQL results and exact-head independent review are recorded on the PR. Original probes, red source/failures, corrective snapshots and receipts remain under the existing private evidence root in `lab-report-references-158-20261106`.

## Separate #158 limitation

The initial report UI prototype consumed direct graph adjacency and compared its cardinality with raw report references. Independent review identified a reverse, non-result Observation.focus link that invalidates that assumption. An actual encrypted native fixture reproduced the extra reciprocal child. A focused frontend regression also covers an unavailable true result replaced by a non-result child with equal cardinality. This is a clinical correctness failure; the UI prototype is withdrawn rather than shipping another heuristic. Its original commit `c9c7fc39ee01ee4c77f59ab42313f9092ba02f90`, original tests and all red/green evidence remain preserved in history/private evidence.

The current importer resolves Bundle fullUrl aliases into generic relationships, but discards the alias map and reference-field provenance after processing. Raw URNs cannot safely be reduced to business IDs, and relationship adjacency cannot reconstruct discarded field identity. #158 remains open: a result-qualified resolution contract and explicit treatment of previously imported records are required. This PR performs no automatic reimport, metadata backfill, raw-FHIR rewrite or migration and does not claim to fix the Labs page. It does not infer that a missing result is absent clinical data.

## Limits and rollback

The existing graph still constructs the owner's relationship graph before root selection. Traversal is not an atomic database snapshot: each content read rechecks current ownership/deletion, but no new multi-query serializability guarantee is made. Accepted #438 optimistic-snapshot and non-atomic multi-record-save limits remain unchanged. The accepted main integration retains PR #443's independent diagnostic-privacy correction.

Source-only rollback uses a normal protected revert and must not change databases, native identities, grants, credentials or retained raw records. Both independent retention worktrees and all unique historical refs remain held. No real clinical data, provider connection, live recovery, deployment or Authentik operation is performed. [#425](https://github.com/karamhouse/chartrove/issues/425) remains a separately authorized operational acceptance gate.
