# Retire unsupported frontend starter tooling

Installed @angular-devkit/build-angular 22.2.5 provides neither tslint nor protractor
builders. Before any change, ng lint and ng e2e both exit 127 before running checks.
The only E2E assertion expects the Angular starter welcome message, which is absent
from Chartrove. This is not working application coverage and is not reported as a
passing test. [Exact source and hashes](../retired-frontend-tooling/manifest.json)
retain all five original scaffold/config files as byte-identical .disabled copies.

Remove only those unsupported targets/scripts and seven developer-only packages:
codelyzer, tslint, protractor, ts-node, @types/jasminewd2, jasmine-spec-reporter,
and unused karma-coverage-istanbul-reporter. The active karma-coverage reporter,
unit tests, assertions, thresholds and timeouts remain. The last package was not
referenced by the configured reporters. ts-node and the webdriver/reporter types
were referenced only by the retired configured Protractor scaffold. A maintained
frontend linter/full browser journey suite remains unconfigured; no substitute
coverage equivalence is claimed. Angular compilation and current browser/native
checks remain the supported baseline.

Before modifying installed packages, archive the complete current node_modules
with ACL/xattr/file-flag preservation. Zstandard integrity, full tar traversal and
SHA-256 pass. [Dependency archive receipt](frontend-dependencies-before-legacy-tools.json).
The initial old dependency archive remains untouched. All retained direct package
versions (76) and the complete runtime manifest are unchanged. The regenerated
lock removes 112 unused artifact tuples and adds no artifact/version/integrity
triple. No unrelated update or engine bypass is included.

Frozen install, five bridge tests, 404 browser tests/one existing skip, and desktop
frontend production build pass. Storybook's first direct CLI attempt fails before
compilation because Yarn is unavailable on PATH. The documented pinned Corepack/
Yarn script succeeds with Compodoc enabled and a new output directory. This is a
normal manager-invoked build, not a compodoc=false workaround. Generated metadata
is retained separately and the prior documentation state is restored.
[Validation summary](frontend-tooling-validation.json).

Retirement applies to #226 (@types/jasminewd2), #251/#312 (ts-node) and #300
(codelyzer). Their complete current diffs affect only package/lock files; exact
unique heads are [strictly archived](legacy-tooling-heads-recovery.json). Close
only after protected integration, explaining removal rather than falsely claiming
the proposed upgrades implemented. Archive/dependency/expected-tip checks still
apply to branch retirement.

Required CI and a current packaged native check remain mandatory before claiming
integration. No frontend runtime package, backend/driver/Wails contract, real
health record or purchased-app directory is changed. Original-primary coverage
recovery is still pending; #344 now protects future normal report runs by default.
