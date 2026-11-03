# GORM family compatibility and encryption boundary

This corrected replacement covers five overlapping Renovate proposals together:
#232 gormigrate/v2 2.1.7, #278 datatypes 1.2.7, #279 PostgreSQL driver 1.6.3,
#280 SQLite driver 1.6.0 and #281 GORM 1.31.2. Both root and desktop graphs
select this family. Gormigrate 2.1.7 and PostgreSQL 1.6.3 require GORM 1.31.2;
reviewing these as independent historical patches would obscure the resulting
graph. Required imported transitives are pgx/v5 5.10.0, MySQL driver 1.5.6,
go-sql-driver/mysql 1.8.1 and edwards25519 1.1.0.

The uncorrected graph selects mattn/go-sqlite3 1.14.22, bypassing the inherited
replacement that matched only 1.14.17. The existing encrypted-header/wrong-key
test actually failed with “database is not encrypted.” Its failing log remains
preserved. Both modules now replace every selected mattn version with the **same
existing** jgiannuzzi SQLCipher source, pinned to
1.14.17-0.20230719111531-6e53453ccbd3. This changes the replacement match, not
the cipher implementation or its file format. The nominal mattn 1.14.22 version
must not be described as an upgrade to the stock SQLite implementation.

Go minimum 1.26.0, Wails, Fasten module/import contracts, provider mirrors and
the graphs' existing unrelated version differences remain intact. No application
schema, credential policy, encryption key, database path or migration source
changes. No real personal or purchased-app records were inspected or opened.
This does not implement an unencrypted-to-encrypted migration or resolve #185.

## Compatibility evidence

The first complete applied backend run also exposed cross-table query pollution:
paginated list requests could carry a previous table's qualified user/source
filter into the next table, producing “no such column: fhir_account.user_id.”
Each table query now starts from a cloned configured GORM session; user/source
filters, pagination and existing per-table pagination semantics remain intact.
A real synthetic repository regression covers two populated resource types,
another source, another user, one-result limits and exhausted offsets.

Identical final tests pass on the old graph/original source, fail on the new
graph/original source and pass with the session correction. An initial fixture
reused global source/type/resource identities across users and overwrote its own
earlier rows; that fixture also failed on the old graph. Distinct synthetic IDs
correct the fixture without changing assertions. An initial source-overlay path
used the wrong macOS path casing and was ignored; the qualified comparison uses
the exact directory returned by Go. All failed controls remain preserved.
One focused handler run without a private TMPDIR failed during its manual-upload
setup (HTTP 500); a diagnostic using a fresh private TMPDIR passes. The original
setup failure's underlying cause remains unverified and is not called fixed.

An explicitly private synthetic fixture was created by the old root graph using
the actual embedded application. The helper blocks external HTTP, refuses an
unmarked directory and refuses creation over an existing database. It creates
a synthetic administrator and imports Patient and Observation resources,
including zero, false and a fractional timestamp with +03:00 offset.

Separate copies cold-open under the corrected new root graph and the corrected
desktop graph. They reject a wrong key, accept the test key and credentials,
retain the administrator role, preserve both raw records canonically and retain
an encrypted header. The old graph also cold-reopens a copy previously opened
by the new graph with the same checks. The immutable original synthetic seed's
bytes and recorded metadata remain unchanged. Initial root cold-open validation
used an intermediate pair of exact-version replacement matches; its effective
selected cipher module is JSON-identical to the final version-independent match.
Final root PostgreSQL and desktop embedded/cold checks use the final graph.

The opt-in PostgreSQL fixture passes on the old root, new root and new desktop
graphs: each has 12 pass events, no failures or skips. These exercise both pgx
query modes, parameter binding, fresh migrations, timezone persistence/reopen,
legacy account handling and startup-failure pool cleanup. The pinned local
PostgreSQL 17 Docker fixture uses loopback, tmpfs, a generated secret and no host
database volume. Its owned containers are independently confirmed absent after
testing. The required pinned image was downloaded; no unrelated cache was pruned.
This does not establish every production PostgreSQL schema or resource query.

Complete MVS metadata changes include upstream test/unused-driver dependencies
that are not imported by this application. The complete before/after graph is
retained (54 root changes; desktop additionally sheds an unused xerrors entry).
The root backend/test imported-module comparison and current desktop import
inventory restrict imported changes to the ten intended family/transitive
modules. The first comparison incorrectly assumed complete graphs contained only
imported dependencies and stopped. A desktop old-graph comparison also required
an exact old-root manifest descriptor because the actual root had been updated;
the remaining xerrors difference is recorded explicitly. No failed guard was
represented as a passing check, and no application source was altered by these
diagnostic comparisons.

Applied-checkout verification and vet pass for both graphs. The complete backend
run has 342 pass events and 15 skips; desktop has 33 pass events and two skips.
All embedded desktop-graph checks pass (12 pass events) under the unchanged
120-second limit. Expanded final PostgreSQL checks include cross-table pagination
and exhausted offsets: old/original, corrected root and corrected desktop each
pass 14 events with no failures or skips; no owned containers remain. The first
final PostgreSQL admission failed before tests because its pinned fixture image
was unexpectedly absent (actor unknown). The exact documented image was pulled
again; all original failures remain retained, and no unrelated cache was pruned.
[Compact validation receipt](gorm-family-validation.json). Protected Backend,
Frontend and Desktop checks are required before integration. Native WebView behavior and a fresh packaged app are not claimed
from Go tests. The last local package predates the Chart.js/GridStack updates.
No full build was run solely for the already merged artifact-retention change;
storage admission and all existing assertions remain unchanged.

## Recovery and disposition

All five exact historical heads, original full intakes and fetched patch payloads
are retained. The shared 7,852-byte Git bundle has SHA-256
c588137f3bde424f8b84a3d9b9f9db8be9fcb6c3b6bf1064f72496667ef60943.
Strict pack validation and original-pack offset checks prove all 20 unique
objects are present; prerequisites are reachable from preserved main.
[Shared recovery](gorm-family-heads-recovery.json).

Raw logs, failed probes, checksums, module descriptors, synthetic fixtures and
the cold-open helper are retained under
dist/validation/2026-11-01-consolidation/gorm-family-review/.
Relocation records map original temporary paths to this durable directory without
rewriting the original command receipts. Original PRs remain open until protected
integration and fresh individual whole-content/head guards. Branch retirement
also requires exact-tip leases, all-PR consumer checks and ownership checks;
the independent artifact-retention branch/worktree is outside this batch.

## Protected publication

Merged #394 at 4a4efc69f3c031e1a08b0ac01581229947ae92bc after Backend, Frontend
and Desktop checks passed on exact head 497861ad460ea9743da74bbe65f6f8ccfcf21dd9.
Full original-content/head guards passed separately for each of the five original
PRs before closure as implemented by the corrected replacement. All five
historical tips and the merged task branch were retired in one atomic exact-tip
lease operation after fresh consumer, reachability, archive and worktree ownership
checks. The independent artifact-retention owner remains untouched.
[Retirement receipt](gorm-family-retirement.json). The fresh checkpoint has
163 issues, 15 PRs and 22 origin heads; this remains an incomplete consolidation.
