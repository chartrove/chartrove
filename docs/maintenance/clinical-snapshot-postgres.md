# Clinical snapshots on PostgreSQL

## Scope and ownership

Pi exclusively owns [#479](https://github.com/chartrove/chartrove/issues/479), in the canonical
single checkout `/Users/osama/Code/chartrove/chartrove`, branch `task/clinical-postgres-acceptance`.
This is synthetic repository/runtime qualification, not PostgreSQL production deployment, TLS,
at-rest encryption, independently operated recovery, native GUI or provider acceptance.

All seven checks qualified exact source head `e0c02909e447882f1deea92f200ab3eefbd3ec7b`
(42601840418), with no unresolved review threads. Protected matching-head #480 merged at
`3c1acc588600fe4d5ef8b072696cc6302c2a2da3`; exact-main validation 42605419029 passed.
#479 remains open for the explicitly documented instrumented maximum-load qualification limit.

The existing `postgres_integration` fixture uses PostgreSQL 17.11, the repository-pinned multi-platform
image `postgres:17-alpine@sha256:266d3c257594e195b309ae9d1805e54c788f1d7b375f56bf6c367aac6c917e7b`.
Only this exact image was acquired. No dependency/catalog refresh, Docker context change, existing
server/database, host volume or real records were used. The fixture requires a local Unix socket,
binds an ephemeral port on `127.0.0.1`, generates its own password, bounds CPU/memory/tmpfs and
removes only its own containers. The pinned image remains cached for subsequent tests.

## Reproduced correction

PostgreSQL stores indexed `resource_raw` as JSONB. Its returned text normalizes whitespace, object
order and numeric notation. Comparing that representation directly with newly compacted input
incorrectly classified an identical Patient as updated whenever an Observation changed.

`TestClinicalSnapshotPostgresPersistence` uses the full application migrations and the same explicit
immutable synthetic principals as SQLite. Before correction, its real two-version import failed:
expected one updated Observation, actual two updated resources (4.915s package result). The first
unchanged regression passed after correction (4.578s).

The bounded fix compares against PostgreSQL's own JSONB text representation, inside the existing
private transaction and cancellation lifetime. It does not use floating-point reserialization or
JSONB numeric-value equality. In particular:

- A last-decimal-place change `1.234567890123456789` → `1.234567890123456788` remains an update.
- Changing recorded scale to `1.2345678901234567880` also remains an update, not numeric equality.
- Entry `fullUrl`/`SourceUri` must still match; equal indexed JSON cannot retain stale provenance.
- Whole-snapshot `bytea` originals and their SHA-256 remain byte-exact, independent of indexed JSONB.
  Indexed JSONB is **not** a byte-exact substitute for the original, including numeric notation.
- SQLite retains its existing byte comparison. No schema, provider, parser, identity, timeout or
  global SQL logger behavior changed.

A provisional `Raw(...).Scan(...)` implementation passed the count regression but the subsequently
added real INFO SQL canary failed: GORM's temporary scan recorder bypassed the private parameter
filter and interpolated clinical input (10.391s). This was an uncommitted implementation defect,
not a claim about previously accepted application behavior. The final helper uses `Rows`, retains
parameter filtering, checks query/scan/iteration failures and closes the result. A deliberately
failed normalization query returns a fixed storage error and rolls back rather than dereferencing
a nil SQL row. Raw failure logs and source/profile evidence remain private.

## Runtime controls and receipts

The PostgreSQL tests exercise actual application migration/reopen, exact originals and digests,
unknown extensions, full decimal precision and scale, replay and historical acknowledgement,
owner/Patient/origin separation, admin/delegate original denial, EDIT read-only provenance,
revocation, Patient/Observation entry-only provenance changes, final-audit rollback, late
cancellation, normalization failure, corruption refusal, concurrent replay, source/account deletion
rollback and competing import/account deletion. Synthetic storage principals bypass password hooks;
these tests do not claim authentication or independent consent acceptance.

- The first existing PostgreSQL baseline passed in 48.361s; opt-in retrieval evaluation was skipped.
- A subsequent complete ordinary PostgreSQL selection passed in 178.833s with both bounded retrieval
  evaluation/quality flags enabled. This preceded the final extra scale/corruption assertions.
- The later explicitly selected PostgreSQL race run passed in 474.695s, including those extra
  assertions, all existing PostgreSQL tests and bounded retrieval evaluation/quality. **It excludes
  the separately measured maximum-load test; it is not a complete maximum-load race pass.**
- Unchanged encrypted SQLite persistence/provenance race controls passed in 16.040s. Its complete
  5,001-resource maximum test passed ordinarily in 14.216s. Tagged database vet/diff checks passed.
- After normally integrating accepted #478 at code checkpoint
  `d9b8cfaa8d0bb8b620467025cc717565b7bf21ed`, the complete ordinary PostgreSQL selection passed
  again (175.279s), including the final scale/corruption assertions and both bounded retrieval flags.
  Maximum import/change took 18.452490458s / 20.380358125s, with both originals retained.
  Serialized complete native internal race suites passed: host 22.636s, composed encrypted shell flow
  27.864s, SMART 5.896s, fixture 1.241s. Existing macOS 27.2-object/11.0-link warnings remain.

The ordinary PostgreSQL maximum test imports 5,000 Observations plus one Patient, then changes one
last-place decimal in the complete collection. For 1,204,010 bytes it passed with 5,001 creations,
then one update and 5,000 unchanged resources, and byte-exact originals for both versions:
20.181074s initial import, 18.924578042s changed import, 41.78s test. Writer reservation was reached
after 12.235786042s / 12.948540958s respectively. These are local measurements, not deployment SLAs.

### Known instrumented maximum-load limit

Running that same maximum test under `-race` exceeds the unchanged two-minute import budget.
The diagnostic run failed in 125.232s and recorded **zero writer reservations reached**: failure
occurred during pre-transaction projection construction, not a blocked PostgreSQL writer. CPU
profiling showed substantial Goja projection execution and external/race-runtime cost. Ordinary
execution of the unchanged workload passed as above. This does not establish all possible causes
of latency or qualify the maximum under instrumentation.

A scoped association-batching experiment did not fix the instrumented failure (164.997s package
result, maximum subtest 124.80s) and was removed after the phase evidence. Its exact source/logs were
preserved privately. No deadline, hash cost or assertion was relaxed, and no test was made to skip
this failure. Consequently `go test -race -tags=postgres_integration ... -run Postgres` is **not**
reported green for this slice. The bounded race command explicitly selects the other tests; the
maximum remains a separately reproducible ordinary runtime gate and a documented race-load limit.

## Reproducing the scoped checks

From the canonical checkout, with the exact pinned image already available to the local Docker daemon:

```bash
set -euo pipefail
export GOPROXY=off GOSUMDB=off
CHARTROVE_RETRIEVAL_EVALUATION=1 CHARTROVE_RETRIEVAL_QUALITY=1 \
  go test -tags=postgres_integration ./backend/pkg/database -run Postgres -count=1 -timeout=15m

# This explicitly excludes the known instrumented maximum, not its ordinary runtime gate.
names=$(go test -tags=postgres_integration ./backend/pkg/database -list Postgres \
  | grep '^Test' | grep -v '^TestClinicalSnapshotPostgresMaximumCollection$' | paste -sd '|' -)
CHARTROVE_RETRIEVAL_EVALUATION=1 CHARTROVE_RETRIEVAL_QUALITY=1 \
  go test -race -tags=postgres_integration ./backend/pkg/database -run "^($names)$" -count=1 -timeout=15m
```

These Go module flags prevent dependency downloads, not arbitrary networking; the reviewed fixture
uses actual loopback PostgreSQL. This is not a sandbox/network-denial certification.

All owned disposable containers were removed after successful and failed runs; unrelated workloads
were untouched. These receipts do not qualify a fresh packaged binary, GUI, older OS, production
PostgreSQL security/operations or independently operated backup/recovery.
