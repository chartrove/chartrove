# Retired frontend scaffold

These byte-identical .disabled files preserve the old TSLint configuration and
Protractor starter scaffold. [SHA-256 manifest](manifest.json). Git history also
retains their original paths and attribution.

Installed Angular build tooling 22.2.5 has no tslint or protractor builder.
Before retirement, both ng lint and ng e2e exit 127 before running any checks.
The sole Protractor assertion expects the generated starter welcome paragraph
“fastenhealth app is running!”, not a Chartrove workflow. It is not reported as
passing or as providing current browser end-to-end coverage.

Remove the unsupported targets and package scripts, plus their exclusively
configured developer dependencies: codelyzer, tslint, protractor, ts-node,
@types/jasminewd2 and jasmine-spec-reporter. The unreferenced legacy
karma-coverage-istanbul-reporter is also removed; the active karma-coverage reporter
and its HTML/LCOV/text reporting remain unchanged. Runtime dependencies and all
retained direct package versions stay unchanged. No lint rule or working test
assertion is relaxed to manufacture success.

The supported baseline remains Angular production compilation/type checking,
Karma browser tests and the native WebView/Go integration checks documented in
CONTRIBUTING. A maintained frontend linter and full browser journey suite would
need a separate configured, validated integration; these retired commands do not
provide those capabilities. This does not claim native smoke checks replace all
browser end-to-end coverage.

The pre-change installed dependencies are retained in the verified archive at
dist/validation/2026-11-01-consolidation/frontend-dependencies-before-legacy-tools-20261102.tar.zst
(SHA-256 983c70481647a788aa5ffce2cd763649a85b7b32a23f2be21b40d8fa37fb2243).
The original initial-install archive remains untouched. CLI failures, complete tar
member validation and lock/version comparison are retained in the same evidence
directory. This retirement must pass the active builds/tests before integration.
