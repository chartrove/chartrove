# React 19 Storybook tooling replacement for #326

React and ReactDOM development dependencies move from 18.3.1 to 19.3.1, and
React declarations from 18.3.31 to 19.3.0. Existing Storybook 10.6.5 already
resolved React 19 for its broad peer range. This coalesces that existing runtime
with the explicit root development dependencies. The Angular application has no
production React imports. Seventy other direct versions and unrelated lock
blocks remain unchanged; only obsolete React 18 helper selections are removed.
Go, database, Wails, provider, attribution and record contracts are unchanged.

Official registry archive SHA1/SHA512 checks pass. Every installed published
React (27), ReactDOM (43), and declarations (16) file matches those archives.
Frozen installation, ten Node checks, 523 browser tests (one existing skip),
production and Storybook builds pass. No coverage threshold or production budget
was changed. Original absent coverage/documentation metadata state is preserved.

The old production/native frontend reuse gate failed: emitted runtime local
variable names, maps and index references differ. Its unminified mapped runtime
sources and all other application chunks/styles/polyfills are identical. A fresh
native package is required; old package byte equivalence is not claimed.

Interactive Storybook acceptance remains unverified. The browser connector was
unavailable. A staged copy of installed Chrome and an official Chrome-for-Testing
archive both failed strict signature verification before launch. Installed Chrome
was untouched. Failed diagnostics and copied metadata are retained privately;
no browser was re-signed or signature check bypassed. The task-only loopback
static server was stopped and its PID absence verified.

The historical branch remains live pending protected integration and fresh
exact-tip/dependency checks. Its incremental verified bundle preserves all five
unique objects; original intake bundle/reference remains separately recorded.

[Validation](react19-validation.json), [recovery](react19-head-recovery.json).
Raw local evidence: `dist/validation/2026-11-01-consolidation/react19-review/`.

Fresh native package source `d0b98e1a56d48ab14efd5d44251c93400bc63b49` passes ZIP CRC, strict extracted
signature and all ten synthetic native checks. All 22,289 current compiled,
reused and native asset files match. Original output content, stat metadata
and xattrs restore exactly. ZIP SHA-256 `cad5fe98662678f6058425fdf71213cead2718548088624458f6d23bffa78a38`.
[Native receipt](native-d0b98e1a5-validation.json).

Integration: #387 merged at `1220f9208f8982bebb312e4ae78da31836d71103`
after all three protected checks passed. #326 is individually reconciled, and
its verified exact historical tip and merged task branch were manually retired
after fresh dependency review and guarded atomic expected-tip deletion.
[Retirement](react19-retirement-1.json).
