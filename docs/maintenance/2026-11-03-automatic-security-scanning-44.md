# Automatic code-security analysis (#44)

Issue [#44](https://github.com/karamhouse/chartrove/issues/44), imported from
Fasten #124, requests automatic security scanning and links marketplace
SonarCloud/Snyk actions. Full original content/discussion is preserved. The
[old SonarCloud entry](https://github.com/marketplace/actions/sonarcloud-scan)
now directs users elsewhere; the [Snyk entry](https://github.com/marketplace/actions/snyk)
requires a separate SNYK_TOKEN. Neither vendor is provisioned by this change.

The reviewed alternative is GitHub-native CodeQL for the verified public
repository. [Official eligibility/setup](https://docs.github.com/en/code-security/how-tos/find-and-fix-code-vulnerabilities/configure-code-scanning/configuring-advanced-setup-for-code-scanning)
supports a tracked advanced workflow with its built-in token. Existing selected
Actions policy already allows GitHub-owned actions. No external credential,
paid subscription, new service, cloud deployment, publishing workflow or
Actions allowlist expansion is introduced. The retired Development registration
stays disabled; the existing Validation triggers/concurrency and original three
jobs/permissions are structurally and byte verified unchanged.

## Analysis scope and trust boundaries

A matrix job analyzes Go, JavaScript/TypeScript and Actions workflow sources.
Go traces explicit root/desktop builds on macOS using Go1.27.3, while JavaScript
uses the frozen frontend graph with Node26.13.0 on Ubuntu. The application lock,
module/toolchain decisions and exact SQLCipher replacement stay unchanged.
The scanner action is pinned to stablev4.39.2 commit
1fc58d04a4d4374f921bdac573c3811c240f9c26. Its action runtime is Node24, separate
from the app's Node26 requirement. Reviewed source recommends CodeQL2.27.4 with
prior2.27.3 fallback; actual CLI versions must be recorded from execution.
[Supported languages](https://codeql.github.com/docs/codeql-overview/supported-languages-and-frameworks/)
include the current Go1.27/TypeScript6. Default queries are not suppressed and no
source/rule exclusions are added. Platform/build-mode limits remain; this is
not universal coverage of every build tag, Swift/native helper, dependency,
secret, runtime threat or clinical-data behavior.

Only the scanner job receives security-events:write and contents:read for native
GitHub result upload. No repository secrets, content/package/deployment/identity
write permissions, privileged PR trigger or untrusted shell interpolation is
introduced. Checkouts do not persist credentials. Expanded CodeQL database upload
is disabled, SARIF processing is awaited, and expanded output stays on ephemeral
hosted runners. No local scanner/build output or full workflow below50GiB is run.
Real health records and purchased Fasten data are never supplied to scanning.

The candidate passes the installed actionlint and strict structural checks,
with original three jobs/triggers/permissions/concurrency unchanged. The workflow
is staged through the native Git index, without materializing its excluded path;
original sparse patterns and S flag remain intact. The targeted hardening review
finds no critical/high/medium/low issues; its single informational note is the
requested scanning gap. [Pre-application review](2026-11-03-codeql-workflow-hardening-review.md),
[admission and exact source](codeql-admission.json).

## Protected execution and findings

PR [#407](https://github.com/karamhouse/chartrove/pull/407) merged at
07ea27fbdbcdc2596074596e9050e5c5243b64ab after all seven trusted checks passed.
The original Backend/Frontend/Desktop checks and strict rules remain intact;
three language jobs (Actions app15368) and the native CodeQL aggregate
(GitHub Advanced Security app57789) now also protect PRs. No bypass actor exists.
The independent main push completed six Actions jobs successfully; the native
aggregate is PR-specific and is not claimed as a seventh main-push job.

CLI2.27.4 processed three diff-filtered PR analyses with zero results. The first
full main baseline, separately inspected, reports three high go/sql-injection
alerts (#1/#2/#3), all in aggregation SELECT/GROUP/ORDER construction. These are
not dismissed or hidden. A disposable two-owner SQLite control confirms that
function/property interpolation can read another synthetic user's account value,
despite the outer owner filter. A corrected parameterized builder is being
prepared separately; #44 remains open while findings are reconciled.

Main Actions and JavaScript/TypeScript analyses have zero results within reported
coverage. JS logs report527/531TypeScript,107/107HTML,8/8JavaScript. Extraction
log reconciliation covers528/532tracked .ts paths; four hidden Storybook paths
are absent: compodoc-compat.ts, main.ts, preview.ts and typings.d.ts. These two
metrics are retained separately; neither means every tracked file was analyzed.
Go uses both explicit module builds and default queries without exclusions.
Scanning is not comprehensive assurance of all runtime/data/dependency behavior.

[Execution, exact log hashes and qualifications](codeql-ci-validation.json),
[checkpoint after407](2026-11-03-checkpoint-after407.json).
Raw evidence: dist/validation/2026-11-01-consolidation/post-dependency-issue-review/issue44/.

Protected#408 fixes the three original exploitable aggregation alerts; all seven
PR checks and six main-push jobs pass. Its complete main Go analysis still
reports one identifier/error-state flow, unlike its zero-result diff-filtered
PR analysis. This distinction remains explicit.#44 stays open; no dismissal or
suppression occurs. [Exact post408 execution](aggregation-query-ci-validation.json),
[separate structural follow-up](2026-11-03-aggregation-schema-identifiers.md).
