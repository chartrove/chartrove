# Clinical projection qualification: unsuccessful controls

## Ownership and disposition

Pi owns #504 under #479, in the sole checkout on
`task/clinical-projection-concurrency`. Parent/source baseline:
`b33a994571f45897867d7c920029d1f50503b265`, protected #503 plus separate exact-main
Validation42735717066 success. No clinical optimization was accepted, committed or
published. This receipt preserves failed controls, not a maximum-race fix.

The three-file Go trial was preserved byte-for-byte in private task evidence, with
`cmp`, SHA256, tracked patch and status verified before restoring only this task's
tracked file and removing its two new files. The JavaScript prototypes never entered
the working tree. Branches/worktrees, other tasks, ignored files and historical
evidence were not deleted. Stop this local trial series; no repeated fanout/cache or
host experiments are proposed by this receipt.

## Unchanged actual maximum gate

Actual pinned disposable PostgreSQL:
`postgres:17-alpine@sha256:266d3c257594e195b309ae9d1805e54c788f1d7b375f56bf6c367aac6c917e7b`,
present on the selected local Unix-socket Docker daemon, Linux/arm64. The reviewed
fixture used `--pull=never`, loopback ephemeral publication, generated secret,
768MiB/2CPU, 512MiB tmpfs and owned cleanup; no existing database, host data volume,
context mutation or image/catalog acquisition. No test-labeled containers remained.

Existing `TestClinicalSnapshotPostgresMaximumCollection` ran with `-race`, original
15-minute package budget and unchanged two-minute import budget, 5,001 resources,
1,204,010 bytes, 16MiB cap, assertions, precision, digests and exact-original checks.

| Source/control | Package / test seconds | Result | Writer reservations |
| --- | --- | --- | --- |
| Clean accepted b33 baseline | 176.171 / 174.18 | `context deadline exceeded` | 0 |
| Finite two-worker Go trial | 135.921 / 133.87 | `context deadline exceeded` | 0 |
| Two-worker plus private absence-only JS overlay | 308.562 / 303.41 | `context deadline exceeded` | 0 |

All actual Go-test exits were1. Three candidate Go hashes, and the overlay's private
JS/configuration hashes where relevant, matched through completion. These are
pre-writer failures, not demonstrated PostgreSQL deadlock or preparation failures.
Package/test durations include work outside the import timer; their differences do
not identify host, fixture, Goja or scheduling causation. No timeout or assertion was
relaxed, race instrumentation disabled, resource skipped, writer moved earlier or
work moved outside the import budget. Ordinary maximum success remains distinct.

## Bounded Go trial, not adopted

The original projection construction was factored without changing extraction, raw
serialization, presentation or URI handling. A serial feature control failed4.257s
because two independent builders did not start. Finite two-worker unit race2.319s
then passed index order, immutable input, active-worker bounds, caller cancellation,
no partial success, already-canceled scope and error controls. A panic control joined
its blocked sibling before rethrowing in the caller goroutine, preserving caller
recovery instead of an independently fatal worker panic.

Workers kept the original import context and waited for every owned builder before
return. Generated search extraction/Goja exclusive checkout remained unchanged; no
SQL/writer reservation until all projections were ready. Those unit properties did
not make the actual maximum race pass, so the trial was not adopted.

## Read-only microprofiles and rejected hypotheses

Private programs used only synthetic projections, no database or collection import.
Goja/bootstrap/FHIRpath AST warmup preceded CPU profiling. Host state was uncontrolled;
none of these figures predicts a production speedup or 5,001-resource deadline.

- Warmed256 sequential race projections19.954881458s; CPU samples12.66s across20.07s
  wall time/63.09% coverage. Goja VM44.55% and array-reduce28.04% are overlapping
  cumulative figures. Compilation/parser only0.20s/1.58%: a constant invocation
  compile cache was not adopted as an evidenced maximum fix. Bootstrap programs and
  per-runtime FHIRpath ASTs were already cached.
- Same initial private binary/input256, one run per mode: one worker20.96122625s,
  two6.260826833s; four-worker variant with existing two-idle-VM pool13.371519334s.
  No race/DB; larger fanout/pool changes were not adopted and churn was not proved.
- Typed-union prototype: paired two-worker256 original19.335184458s versus
  prototype23.460715167s. No speedup evidence; not adopted.
- Absence-only prototype: original59.571295333s versus prototype47.651016958s.
  Absolute timings differed substantially from prior pairs; no causal or speedup
  acceptance. Its actual unchanged maximum overlay still failed as recorded above.

The typed-union differential initially compared raw JSON strings and encountered
native Go-map object-key enumeration differences. This was a private diagnostic
comparator issue, not an application assertion changed to green. The corrected
comparator ignored only object-key order, retaining array order, `__path__` and exact
`json.Number` lexemes. Original-self402 and paired402/67 generated Patient/Observation
scripts/six synthetic fixtures plus eight conservative fallbacks passed. This was
not complete FHIR equivalence or sufficient performance evidence.

The absence-only private prototype always compiled the original expression and
cached only static plans. It bypassed execution only when every relevant simple
concrete Patient/Observation top-level field and primitive `_field` was absent;
present/null/empty/extension fields, complex/abstract/other-type paths kept the
original evaluator. Original-self737 and paired737/67 scripts/11 fixtures plus eight
fallback controls passed, including null/non-string type, primitive extensions and
unknown-capital fields. No resource/result cache was introduced. These controls did
not overcome its actual maximum failure and do not justify shipping that prototype.

## Remaining boundary

#479/#504 remain open. Exact source, private trial copies, programs, CPU profiles,
comparison failures, corrected controls, full logs and exits remain private; no raw
logs, synthetic JWTs, configurations, keys or archives are published here. No SDK,
module cache, dependency, generated catalog, idle-pool size, OS setting, scheduler,
priority, foreground/capture policy or test budget changed. #487's stopped host
experiments/runtime negatives, normal visual consent, independent recovery and
production operations remain separate. A different runtime or projection design
needs its own evidence and scope; none is approved by this failed-control receipt.
