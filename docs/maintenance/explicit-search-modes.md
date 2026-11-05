# Optional search modes (#419)

Source proposal, subject to independent PR review and protected acceptance. Existing
clients omit `mode` and retain the current literal substring query, response shape,
50-hit pages, ordering and source authorization. Explicit `literal` is equivalent.

The smallest useful addition is `exact_id`: byte-exact resource ID equality within
the selected source and resource type. Titles, transcript terms, case folding,
Unicode normalization and prefix matching cannot establish identity. A bound SQL
predicate uses the existing source/owner/grant/deletion query; no index or schema
change is necessary. The optional capability leaves existing repository interfaces
and clients intact. Unknown modes fail instead of falling back to another meaning.

Ordered token phrases are evaluated separately. The accepted semantics require NFC,
Unicode simple lowercase, adjacent maximal letter/number/mark tokens within one
field, and current reviewed-import provenance. Native full-text tokenizers differ
between SQLite and PostgreSQL and cannot silently substitute for these semantics.
A portable candidate can scan current authorized rows with a precompiled linear
token matcher; its actual query/decoding/authorization cost must be measured. It
will be omitted if it misses the existing correctness or performance criteria.
Normalized stored projections/UDFs/index maintenance would expand this increment
and are not automatically justified by desire for a phrase button.

All-token matching is deferred: a note denying a different symptom can contain all
the same words. Its lower exact-set coverage and additional explanation do not
justify expanding this first increment. There is no intent classification, clinical
interpretation, terminology expansion, ranking or model judging.

UI: an explicit mode selector beside the existing search field, initially
“Literal text”; optional “Exact resource ID” searches only the ID and explains
case/spacing sensitivity. Any offered phrase mode must explain punctuation/spacing,
Unicode normalization and CJK substring differences. Changing mode cancels pending
results, clears the selected detail and resets pagination. Applied mode/text must
remain bound together during navigation; source changes reset to literal. Details
continue through the existing authorized endpoint, never trust an old hit as access.

Bounds: preserve the 4 KiB request-body ceiling, valid UTF-8 query up to 128 runes,
existing UI input limit, pages 0–100000, 50 metadata hits and five-second repository
deadline. Exact IDs are not trimmed/normalized; blank input is rejected. Every
request uses current native UUID/source ownership and active VIEW/EDIT grants.
NUL is rejected for exact IDs. Explicit mode must be `literal` or `exact_id`;
null, empty, non-string and unknown modes return 400 rather than silently changing
meaning. The original API/repository method handles literal requests directly;
the new optional repository capability handles only exact IDs.
No body/credentials/owner fields are added to results or logs. Cancellation must
stop database work through request context; UI unsubscribes discard stale replies.

Acceptance uses the frozen 35-case corpus and stored 4/16/128 KiB profile, with
hand-authored expected mode IDs from the accepted receipts. Check actual production
methods, not the test-only reference scan: owner/delegate/admin/unrelated actors,
misowned/duplicate IDs, expiration/revocation, record/source/actor deletion, current
provenance, malformed mode/input, pagination, cancellation and Desktop transport.
Paired, interleaved repeated measurements compare the actual candidate and literal
control on the same synthetic data. Keep the predeclared per-case p95 <=250 ms and
no >25% family p95 regression; retain failures and raw samples. Stop any failing
mode, without loosening semantics or thresholds. No deployment is included.

Prepare one focused PR only for the accepted subset. Protected Backend, Frontend,
Desktop, three CodeQL language checks and native CodeQL, independent review and
required acceptance precede merge. Rollback of an optional mode restores the prior
selector/handler behavior without data or identity migration.

## Measured decision

Ship only exact-ID equality. The [actual implementation receipt](explicit-search-mode-acceptance-2026-11-05.json)
records three interleaved blocks with 3 warmups and 25 measured samples per case
and block, aggregating all blocks rather than selecting a favorable run. Both
engines check all 35 frozen cases; timing uses the established 35 short-profile
cases and 14 representative long-profile cases at 1,000 records per owner. This
measures the production repository methods, including live authorization and DB
reads, not a reference matcher or frontend/network SLO.

| Engine/profile | Highest exact-ID case p95, ms | Highest family p95 ratio vs literal |
| --- | --- | --- |
| Encrypted SQLite, short | 1.849 | 0.202 |
| PostgreSQL, short | 13.381 | 0.463 |
| Encrypted SQLite, long | 4.341 | 0.040 |
| PostgreSQL, long | 113.213 | 0.253 |

Every candidate case remains below 250 ms and every family below 1.25×. Host load
was variable, including slower literal controls; these paired measurements do not
establish production capacity, concurrency or a default-search SLO. The existing
literal SQL, request defaults and handler entry point stay unchanged. Exact-ID
queries narrow through the existing composite equality index and compare encoded
bytes additionally: [SQLite byte comparisons](https://www.sqlite.org/datatype3.html#collation)
and [PostgreSQL conversion to bytea](https://www.postgresql.org/docs/17/functions-binarystring.html).
This guards legacy/custom case-insensitive collations without normalizing identity.

The actual phrase prototype used a precompiled linear token matcher and streamed
authorized current rows. It matched the corpus, but encrypted SQLite family p95
ratios were 1.529–3.086 against the literal control, failing the unchanged 1.25×
gate. It is not shipped or exposed. A normalized persistent projection or engine
specific function would require a separately justified lifecycle/design increment;
silently using different native tokenizer semantics would fail correctness.

An unused generic Go wrapper for literal mode also showed over-limit long-profile
timing variance (1.499× PostgreSQL and 1.539× SQLite). It was removed; HTTP literal
requests already used the original method directly. Final acceptance measures only
the new exact-ID method against that unchanged method. Both failed candidates,
their code, raw samples and errors remain retained privately; no thresholds or gold
sets were relaxed. Per-case percentile and Unicode assertions were added after the
timed run; the measured production code and exact original harness are retained.

Focused acceptance covers the default/explicit literal equivalence, strict mode
validation, byte-exact ID case/spacing/canonical-Unicode differences, a deliberately
case-insensitive SQLite column, owner/delegate/admin isolation and native UUID
replacement, current transcript/provenance changes, expiry/revocation/deletion,
pagination/cancellation, HTTP request limits, UI mode labeling/stale cancellation,
and native-authenticated encrypted embedded/Desktop backend requests. Actual GUI
deployment, live IdP/provider operations and real medical data remain outside scope.
