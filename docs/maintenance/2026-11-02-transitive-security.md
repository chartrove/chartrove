# Transitive frontend security replacements

The current lockfile still contained the targets of historical draft PRs #198,
#201 and #206. Their complete changed-file lists were reviewed; all affect only
frontend/yarn.lock. This batch applies their relevant changes to the maintained
Angular 22 graph, without restoring obsolete framework dependencies.

- word-wrap: 1.2.3 to 1.2.5, matching #198.
- semver: affected 5.7.1 / 6.3.0 / 7.3.8 selections to 5.7.2 / 6.3.1 / 7.5.4,
  covering #201's applicable scope. Existing 6.3.1 and 7.8.5 entries stay intact.
  All current semver entries fall outside [GHSA-c2qf-rxjj-qqgw](https://github.com/advisories/GHSA-c2qf-rxjj-qqgw).
- follow-redirects: 1.15.2 to 1.16.2. The historical #206 target 1.15.4 is now
  insufficient: upstream's [cross-realm sensitiveHeaders fix](https://github.com/follow-redirects/follow-redirects/security/advisories/GHSA-mrwp-f3q8-3g9p)
  requires 1.16.2; it also includes earlier array-form-header corrections.

Only five version/resolved/integrity blocks change. All original dependency
ranges and dependency edges remain unchanged. Package metadata and tarball
integrities were obtained from the npm registry, and pinned Yarn 1.22.22 performs
a successful frozen install on Node 26.13.0. No manifest or Go graph changes.

A local compatibility probe verifies every semver lock entry, actual TSLint and
Compodoc semver consumers, option parsing and word wrapping, and two synthetic
loopback cross-origin redirect cases. Both object-form and array-form credentials
are removed, including custom header names supplied from a separate JS realm.
The probe passes; its initial harness errors (unavailable optional parser and
incorrect expectations for two documented consumer return values) are retained
alongside the corrected probe. These are compatibility checks, not a claim of
application exploitability or comprehensive frontend vulnerability clearance.

The full browser suite passes 404 tests with one existing skip, CLI exit 0;
Chrome's local teardown required forced termination after assertions. The
production desktop frontend build passes with existing 9.92 MB/CommonJS warnings.
Further local coverage writes only to the task's redirected output. Required
GitHub Backend/Frontend/Desktop checks must independently pass before merging.
Native packaged UI acceptance is a separate gate. Existing data and build outputs
were not used as test fixtures.

Retained local evidence: dist/validation/2026-11-01-consolidation/security-*.
[Exact reviewed PR heads](2026-11-02-security-pr-review.json). Close those PRs only
after this replacement integrates, citing its merged revision; original branches
still require archive, dependency and exact-tip verification before retirement.
CryptoJS #199 remains open: changing its PBKDF2 defaults needs separate review.
