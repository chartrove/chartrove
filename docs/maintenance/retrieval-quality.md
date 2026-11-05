# Synthetic retrieval quality and explicit query semantics

For [#419](https://github.com/karamhouse/chartrove/issues/419), evaluate query
meaning before choosing another index. Existing source-authorized literal search
remains the default. This fixture changes no runtime query, API, schema, index,
authentication, dependency version or data destination. Historical contributions
and the earlier native-index measurements remain in
[the storage evaluation](retrieval-evaluation.md).

## Corpus and relevance judgments

`backend/pkg/database/testdata/retrieval-quality-v1.json` contains authored synthetic
documents, query intents, explicit relevant document IDs, engine-specific literal
expectations and reasons. No patient data, external terminology, generated model
judgments or medical ground truth is used. Each engine receives two native owners
with duplicate document IDs, an explicit VIEW delegate, an ungranted administrator,
an unrelated owner, and a misowned-row canary. Current source/owner predicates and
reviewed-import provenance are checked for every query. Only metadata is returned.

Cases cover exact resource/accession identifiers and prefixes; order, whitespace
and punctuation in phrases; separated/reordered words; composed/decomposed French,
Greek case and CJK substrings; literal SQL wildcard characters, quotes and paths;
ambiguous abbreviations and full labels; negative, present and historical statements;
contradictory allergy notes; reviewed-body versus external-title/body provenance;
and genuine no-match queries. Contrary statements remain separate documents. A
term hit is not a current-condition assertion, clinical reconciliation or proof
that an abbreviation has one meaning. Even manually reviewed transcripts may be
wrong; this is retrieval acceptance, not OCR accuracy or clinical acceptance.

## Explicit modes, evaluated without enabling them

| Mode | Meaning | Deliberate limits |
| --- | --- | --- |
| `literal` | Actual current repository/API substring query over ID/title and allowed reviewed text; SQL wildcards escaped; database-native LOWER | No canonical Unicode normalization, phrase tokenization, exact identity or abbreviation inference |
| `exact_id` | Byte-for-byte resource ID equality within the authorized source/type | Not an accession/body query; no prefix, case folding or fuzzy identity |
| `token_phrase` | Adjacent ordered whole tokens within one field; NFC + Unicode simple lowercase, maximal Unicode letter/number/mark runs | Punctuation/whitespace separate tokens; no stemming, accent removal, stopwords, transliteration, CJK segmentation or expansion |
| `all_tokens` | Every query token occurs within one field, in any order/distance, using the same tokenizer | Can join unrelated or differently negated clauses; not a clinical inference or phrase |

Alternatives are bounded reference scans in `_test.go`, using current production
database source/owner/grant/deletion predicates and re-reading current provenance
and text. They are intentionally not proposed production implementations. Empty
token queries match nothing. No operators, implicit mode guessing, cross-field
word joining, ranking, model, index or search grammar is introduced. Results retain
stable resource-ID order and the current 50-hit bound. Corpus cases stay below that
bound; this is not relevance ranking or a top-k/pagination benchmark.

## Acceptance criteria established before measurement

1. Every hand-authored literal expectation must match the actual engine query.
   The selected explicit mode must match each query's stated relevance set exactly;
   no missed, invented, duplicated or wrong-provenance IDs. Other mode/intent
   combinations are measured, not assumed interchangeable.
2. Zero cross-owner/admin/ungranted hits and zero stale results after grant expiry,
   revocation, current-text/provenance change, record/source deletion or native UUID
   replacement. Each invalidation has a positive precondition. No body/owner/secrets
   in returned metadata. SQLite remains encrypted on disk.
3. Report per-case precision/recall and exact-set correctness for all four modes,
   grouped by intent family. Empty/empty scores 1; empty/nonempty scores 0. Prefer
   exact sets over averages, which can hide a broken language or identifier family.
   Selected-mode gold is task-specific: selecting it after reading the query intent
   is not a deployable classifier or evidence that one mode beats all others.
4. Opt-in timing: 1,000 documents per owner, two owners; 3 warmups and 25 serial
   end-to-end delegated queries per case/mode. Report p50 and nearest-rank p95,
   including authorization and database reads, excluding fixture/label assertions.
   An exploratory interactive budget is p95 <=250 ms per case on the measured host.
   It is not a flaky CI wall-clock assertion or a production SLO. A candidate runtime
   replacement must also avoid >25% p95 regression in any intent family on controlled
   repeated runs, preserve all literal-only use cases via explicit mode selection,
   and pass concurrency/cold-cache/large-document/desktop/resource acceptance.
5. Another index requires a demonstrated remaining query/latency need, not merely
   better scores from changing semantics. Compare identical semantics on both engines
   with transactional maintenance, deletion/revocation and storage/backup costs.

Ordinary Backend tests run the small SQLite corpus; the existing optional
`postgres_integration` tag runs the same gold sets in a fresh pinned loopback/tmpfs
PostgreSQL container. Opt-in timing expands only test-owned databases. No extension,
image or model downloads, live credentials, production sessions or clinical records.

```sh
go test -p 1 ./backend/pkg/database -run '^TestRetrievalQualitySQLite$' -count=1 -v
CHARTROVE_RETRIEVAL_QUALITY=1 go test -p 1 -tags postgres_integration ./backend/pkg/database -run '^TestRetrievalQuality' -count=1 -v
```

Retain the exact corpus hash, source revision, engine/version/locale, host and raw
results with each receipt. Host load, warm serial execution and short synthetic
documents limit generalization. These tests do not replace HTTP/session, browser,
desktop or operational Authentik acceptance. Protected Backend, Frontend, Desktop
and all required CodeQL checks still apply. Rollback removes fixtures/docs only.

## Observed results, 2026-11-05

The [source-hashed receipt](retrieval-quality-2026-11-05.json) retains every
case/mode's IDs, precision, recall and percentiles. Encrypted SQLite 3.42.0 and
PostgreSQL 17.11 (`en_US.utf8`, pinned local Alpine image) passed the 33-document,
35-query gold sets and independent lifecycle checks. The opt-in run used 1,000
documents per owner (33 diverse documents plus 967 administrative distractors),
two owners, four modes, 3 warmups and 25 measured queries per case/mode: 7,000
measured requests across both engines. The full run took 57.272 seconds on an
ARM64 macOS host with a resource-capped local PostgreSQL container.

| Mode | Exact relevance sets / 35, SQLite / PostgreSQL | SQLite per-case p95 range, ms | PostgreSQL per-case p95 range, ms |
| --- | --- | --- | --- |
| Actual literal baseline | 22 / 23 | 1.400–1.957 | 5.555–8.743 |
| Exact resource ID only | 7 / 7 | 4.604–5.281 | 7.096–12.536 |
| Token phrase reference scan | 29 / 29 | 8.102–16.141 | 10.403–18.406 |
| All-token reference scan | 28 / 28 | 8.313–9.595 | 10.527–14.860 |
| Mode selected by the stated intent | 35 / 35 | 1.400–16.141 | 5.846–18.406 |

These are deliberately mixed intents, not a claim that an ID-only mode should
answer clinical-word queries. Intent-selected coverage is an oracle using the
fixture label, not an implementation or classifier. All measured cases met the
exploratory 250 ms budget. Reference scans were slower than the baseline and do
not meet the proposed replacement non-regression criterion; no runtime replacement
is justified by this run. One small corpus does not establish a production SLO.

Concrete differences:

- Exact `LAB-0042` identity returned one intended ID instead of three substring
  hits (precision 1 versus 1/3 for that intent).
- The phrase `blood pressure` found two spacing/hyphen variants versus zero literal
  hits. Explicit all-word search found separated/reordered words; phrase search
  correctly rejected reordered wording.
- Normalized `échographie` found both composed/decomposed documents; literal SQLite
  found neither and literal PostgreSQL found only the composed case. Greek lowercase
  likewise exposed the engines' differing native lowercase behavior.
- Whole-word `Na` returned the one abbreviation document instead of five substring
  hits; it did not infer equivalence to `sodium`. Both meanings of `MS` remained
  separate; no terminology service or clinical disambiguation was used.
- Token phrase/AND lost the useful `血圧` substring and broadened punctuation/quoted
  wording. AND also matched a note denying a different symptom for `denies chest pain`.
  Literal/phrase selection is necessary; silently changing the default would regress
  valid cases. Negated and contradictory statements were preserved as evidence.

The first fixture run exposed manually omitted literal `Na` substrings; those
expectations were corrected after inspecting the authored words (exertional,
pulmonary, provenance and nasal), while the intended whole-word relevance stayed
unchanged. An initial PostgreSQL metadata probe used an unsupported `SHOW lc_ctype`;
the receipt now reads the database catalog. Both original failures and all measured
samples remain retained privately. No source authorization or production search
behavior was changed to obtain passing results.

The next useful increment is a focused explicit query-mode design with user-facing
semantics and bounded current-source retrieval, using these gold sets. Any production
implementation still needs frontend/desktop consistency and HTTP authorization and
pagination acceptance. More realistic long-document/concurrent/cold-cache workloads
are required before an index or a performance promise. #419 remains open.
