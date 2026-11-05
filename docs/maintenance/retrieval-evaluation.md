# Native-storage retrieval evaluation

For [#419](https://github.com/karamhouse/chartrove/issues/419), retain existing
source-scoped metadata/reviewed-transcript search as the default. This bounded
synthetic evaluation does not justify Typesense, a vector database or a model
service. It adds reproducible opt-in fixtures and measurements, not a runtime
index, migration, new query endpoint or clinical-data processing path.

Historical ideas and attribution remain in [#211](https://github.com/karamhouse/chartrove/pull/211)
and [Fasten #594](https://github.com/fastenhealth/fasten-onprem/pull/594): groundmuffin
and TheoAtTechStack/TechStackApps contributed optional retrieval/chat, pagination,
composite identities and provenance concepts. Their preserved original head/base
and 189 unique commits remain historical references. Browser-direct global search,
exposed indexing credentials and owner-unchecked indexed detail are not adopted.

## Measured scope

The [machine-readable receipt](retrieval-evaluation-2026-11-05.json) records all
40 query cases, fixture/source blob IDs, raw-receipt hash, optimizer summaries and
index costs. Raw runs, including unsuccessful drafts, are preserved privately.

- Apple M2, 24 GiB, Go 1.27.3; encrypted SQLite 3.42.0 through the existing locked
  SQLite3MC driver. FTS4/Unicode61 works; FTS5 returns `no such module: fts5`.
  No extension, build flag or dependency was changed to manufacture availability.
- PostgreSQL 17.11 in the existing pinned, loopback-only Docker fixture: 2 CPUs,
  768 MiB RAM, 512 MiB tmpfs, no host volume, remote database or image download.
  PostgreSQL does not inherit SQLite's application encryption; deployment volume,
  backup and transport encryption would remain the operator's separate boundary.
- Two owners with 1,000 and then 10,000 synthetic reviewed documents each; repeated
  FHIR IDs across their sources; about 443 bytes of text per document. Separate
  VIEW delegate and ungranted administrator. No actual patient records or models.
- A common word in every document and a rare word in 1%; owner/delegate queries,
  three warmups, 25 samples per case, nearest-rank p95, five-second deadline,
  deterministic resource-ID ordering and the existing 50-result limit.
- Baseline is the actual current API repository query. Index candidates rejoin
  the current clinical rows and repeat the production source/ownership/grant
  predicates, plus current reviewed-origin and decoded-text checks. They return
  metadata/provenance only. Index membership never authorizes a request.

Delegated query timings at 10,000 documents per owner, milliseconds:

| Engine / method | Rare p50 / p95 | Common p50 / p95 |
| --- | --- | --- |
| Encrypted SQLite, existing query | 85.688 / 148.622 | 1.360 / 6.488 |
| Encrypted SQLite, FTS4 candidates + live recheck | 57.960 / 64.352 | 129.423 / 145.450 |
| PostgreSQL, existing query | 46.153 / 50.133 | 3.797 / 8.508 |
| PostgreSQL, GIN projection candidates + live recheck | 34.219 / 37.589 | 2.738 / 3.520 |
| PostgreSQL, materialized GIN candidates + live recheck | 30.592 / 32.016 | 4.452 / 6.453 |

At 20,000 total indexed rows, FTS4 allocated about 12.1 MB of additional database
pages; insertion took 915 ms, excluding initial virtual-table creation and corpus
seeding. The final encrypted SQLite file including schema/corpus/index was about
36.4 MB. The PostgreSQL projection, stored body/vector, primary key and GIN together
occupied about 26.7 MB; creation/population/vector/index/ANALYZE took 1,096 ms. These
are different storage representations and timing scopes, not interchangeable
claims about index-only size or production ingest throughput.

`EXPLAIN` matters: the ordinary PostgreSQL form chose primary-key probes instead
of GIN. Materialization used GIN for the rare candidate set, but the outer ordered
clinical scan still checked about 4,901 texts to return 50 hits. SQLite used its
FTS virtual table while common-term candidate expansion was slower than the early
50-row baseline. An index's existence does not prove the planner or end-to-end
query benefits. No optimizer settings, grant checks or current-content checks
were disabled. Earlier draft runs and their changed query shapes remain retained.

Observed whole-evaluation durations ranged from 21.7 to 44.4 seconds with the same
query shapes; the earlier actor-deletion assertion was superseded by independent
preconditions in the current receipt. This variability is a reason to avoid a
precise capacity or relative-performance promise.

These are observations, not an optimal-query or capacity claim: host load was not
controlled, warm data is not cold-cache behavior, PostgreSQL is resource-capped,
corpora are short/repetitive and only two single-word distributions were sampled.
There is no ranking, clinical relevance, multilingual, abbreviation, negation,
phrase, concurrent-ingest or large-attachment acceptance here. A different query
plan or realistic corpus could favor indexes. Full text uses token semantics;
the current API uses literal substring matching. Do not silently replace one with
the other. [SQLite FTS4](https://www.sqlite.org/fts3.html),
[FTS5 availability/build](https://www.sqlite.org/fts5.html#compiling_and_using_fts5),
[PostgreSQL GIN](https://www.postgresql.org/docs/17/textsearch-indexes.html) and
[CTE materialization](https://www.postgresql.org/docs/17/queries-with.html#QUERIES-WITH-CTE-MATERIALIZATION)
provide the underlying engine contracts.

## Isolation and invalidation acceptance

Both engines and corpus sizes pass synthetic checks for duplicate FHIR-ID isolation;
ungranted administrator/unrelated-owner denial; expired and revoked grants; deleted
records/sources/actors; and stale indexes after text or reviewed-origin changes.
Source deletion and actor deletion have independent successful retrieval
preconditions: the synthetic source is restored before deleting/recreating its actor,
and the old UUID must fail current-user lookup specifically, not merely source access.
The candidate statement also denies access when called after revocation without
repeating the earlier metadata precheck. Returned source/type/resource IDs retain
provenance; transcript and owner fields do not enter result metadata. Existing
middleware handles session expiry/revocation and native identity; this database
fixture does not replace HTTP/OIDC acceptance.

A future runtime index would still need transactional projection maintenance,
versioning and delete/rebuild tests, bounded per-source candidate work, encrypted
SQLite portability (including Desktop), PostgreSQL schema compatibility, backup
and export handling, query-language limits and retained synthetic regressions.
An index lag may reduce recall; it must never restore permissions or deleted text.
Recheck source authorization and current content immediately before any excerpt,
model request, result detail, citation or conversation replay. Do not store grants
in embeddings, index documents or history as an authority. Revocation must apply
on the next retrieval even if a browser retains an old result identifier.

## Processing choices and next decision

| Option | Privacy and authorization | Resources and operations | Current decision |
| --- | --- | --- | --- |
| Existing local lexical search, no model | No new data destination; live source predicates and metadata-only results already work | Existing encrypted SQLite/PostgreSQL; no service, model download or index lifecycle | Keep default; measured baseline remains useful |
| Native full-text projection, no model | Derived clinical text stays in the database security boundary; rejoin current records/grants; never browser-direct global access | Extra disk/write/backup cost, planner and tokenizer compatibility, transactional invalidation/rebuild required | Viable future prototype if query semantics/latency need justifies it; not enabled |
| Fully local model on the application host | Authorized selected text only; disable prompt/response payload logs; isolate histories by native UUID and reauthorize citations | CPU/GPU/RAM, model weights, context/KV cache, latency, updates and offline supply chain need a measured budget; none tested/downloaded | Architecture/resource approval required before introducing a model |
| Separate privately hosted model service | Adds a data destination and authenticated network boundary, even if privately operated; scope inputs and retention explicitly | Separate availability, TLS/credentials, tenancy, observability, backups and incident ownership | Architecture and destination approval required |
| External provider API | Clinical text leaves Chartrove; destination, retention/training policy, deletion, region, logging and credentials require an explicit decision | Network dependency, usage cost, rate limits, provider changes and outage behavior | Not authorized; no clinical records sent |

Any future RAG must treat retrieved text as untrusted evidence, preserve source
citations and uncertainty, avoid diagnosis or autonomous record/tool mutation,
and avoid granting a model broader retrieval rights than the requesting user.
Reauthorize before sending context and before serving stored outputs; revocation
cannot retract text already seen by a recipient, so minimize copies and define
history/cache retention before adding them. Search must remain usable without chat.

The follow-on [retrieval-quality corpus](retrieval-quality.md) measures exact
identifiers, phrases, Unicode, abbreviations and contradictory/negated statements
against explicit evaluation-only modes and relevance/latency criteria. It preserves
the current literal default. Select a native index/query shape only for a demonstrated
remaining need. #419 remains unresolved; these evaluations are not full-text/RAG
implementation or model/data-destination approval.

## Reproduction and rollback

From the normal primary checkout with existing Go dependencies, Docker socket and
pinned PostgreSQL image already available:

```sh
CHARTROVE_RETRIEVAL_EVALUATION=1 go test -p 1 -tags postgres_integration ./backend/pkg/database -run '^TestRetrievalEvaluation' -count=1 -v
```

The fixtures use only fresh temporary SQLite files and disposable loopback/tmpfs
containers and clean up their own resources. They refuse remote Docker contexts;
ordinary tests skip this evaluation without the opt-in variable. Required protected
Backend, Frontend, Desktop and CodeQL checks still apply to the PR.

Rollback is removing the evaluation tests/documentation; runtime schema, API,
authentication, indexing and data destinations are unchanged. No production
migration, model download, service deployment or live data operation is included.
