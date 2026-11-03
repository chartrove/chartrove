# Preserve explicit configuration while unlocking

While reviewing #272 with disposable configuration files, the old configuration
snapshot lost explicit Set precedence after an AllSettings/JSON/MergeConfigMap
copy. Embedded database unlock used this same copy. Environment variables could
therefore replace an explicitly chosen database location and turn explicitly
disabled unsafe routes back on. This is reproduced on the existing Viper 1.12.0;
no dependency update is part of the correction.

Copy the already resolved, JSON-detached settings into the new configuration's
explicit override register using Set. Preserve the separate immutable snapshot,
original request configuration, key validation, database opening and atomic route
publication. This keeps the original explicit database/loopback/unsafe-endpoint
choices effective when AutomaticEnv is enabled. No schema, SQLCipher replacement,
provider registration, persisted record or migration changes are introduced.

## Reproduction and validation

Two new regressions enter and fail under both actual application module graphs
before the fix. One environment override points at a disposable unrelated text
sentinel, while the explicit path points at a fresh database; the sentinel remains
byte-identical in both failing and passing cases. The other verifies that unlock
retains false/loopback settings and registers no unauthenticated unsafe routes.
No real health records or purchased application directories are opened.

Identical regression bytes pass afterward. Both full web suites pass 16 tests;
both unlock race suites pass four; both full vet checks, 28 desktop tests and the
server build pass. Two desktop packages have no test files, with no named test
skips. The encrypted desktop web suite retains the 120-second timeout. Module
files remain byte-identical, and every tracked frontend input matches the qualified
c8c6e4a9f build. Fresh native packaging at 6368cdcc3 passes ZIP CRC, strict signature and
ten synthetic native checks. All 22,287 qualified frontend files remain identical.
Original output content/stat fields and all xattrs restore exactly, with zero
provenance changes. [Native receipt](native-6368cdcc3-validation.json).
Protected integration remains pending.
[Detailed commands and receipts](unlock-configuration-validation.json).

An initial temporary overlay used the checkout alias's lowercase path, while Go's
root package path used uppercase Chartrove. A later desktop overlay saw the
opposite path spelling. Both entered zero tests and are retained as discarded
diagnostics, not passing checks. The qualified red results come from the actual
checkout with both named tests entered and failed in each graph.

Raw source-before snapshots, overlay diagnostics, real red/green logs and sentinel
checks are retained under
`dist/validation/2026-11-01-consolidation/unlock-config-review/`.
The separate Viper candidate and broad graph changes remain unadopted temporary
review evidence; historical #272 remains open.
