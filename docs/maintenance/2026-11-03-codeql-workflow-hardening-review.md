GitHub Actions Hardening — Validation and private CodeQL candidate

| Severity | Count |
| --- | --- |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |
| INFO | 1 |

Review checkpoint: protected main3e6e9a73eb50bc43f7e67988d42fbed254c3ed91,
before applying the private workflow candidate. Existing workflow SHA256 is
6b06988c3ce46bd047dc0e93ad0abf9a53d8681db16a5c7b22cdebc771f07166.
Candidate SHA256 is08abe669c013731019c4d4c15696266d073bb9ad24dec35b4f1a9916226c5a27.

## INFO — Automatic source-security analysis is absent (#44)

The current tracked workflow begins `name: Validation` at line1. Its three jobs
cover builds, vet and synthetic tests, without an automatic source scanner.
This is the requested feature gap, not a vulnerability finding or claim that
those validation jobs are unsafe. The attributed issue points to marketplace
SonarCloud/Snyk actions. The old SonarCloud marketplace entry now says to use
another action; Snyk requires a separate SNYK_TOKEN. The candidate uses GitHub's
native CodeQL eligibility for this verified public repository, with no new
account, private key, subscription or expanded Actions allowlist. The selected
policy already allows GitHub-owned actions.

## Triggers, injection and privilege

Existing pull_request, push-main and manual triggers are preserved exactly.
There is no pull_request_target/workflow_run or other privileged external trigger.
No PR title/body/head text enters a run script. Matrix language/OS/build-mode
values are constants and used as action inputs, conditions or job labels.
Concurrency metadata is a key, not executable code. Hosted runners remain
ephemeral. All checkouts retain persist-credentials:false.

The original three jobs retain contents:read and byte-identical job definitions.
Only CodeQL receives security-events:write for GitHub scanning-result upload,
plus contents:read. It has no contents/package/release/deployment/identity write
permission and references no repository secrets. CodeQL's documented fork-PR
handling applies; no base-context execution of fork code is introduced.

## Supply chain, build and output handling

All action references use full40character SHAs. The added github/codeql-action
init/analyze use reviewed stablev4.39.2 commit
1fc58d04a4d4374f921bdac573c3811c240f9c26, with its Node24 action runtime.
Its source defaults recommend CodeQL2.27.4 (prior2.27.3); actual runtime CLI
versions/coverage still require observation. Official current language support
includes Go1.27 and TypeScript6. Existing application Go1.27.3/Node26.13.0,
locked dependency installation and disabled persistent caches are retained.

Go traces explicit root/desktop builds on macOS; JavaScript/TypeScript uses
locked dependencies on Ubuntu; Actions scans workflow sources. Default queries
are not suppressed and no path/rule exclusions or skip-queries are added.
Database upload is disabled; findings/SARIF processing go to the repository's
native scanning service. This is source analysis, not medical-record indexing,
SCA, secret scanning, native UI acceptance or proof that all code is secure.
Existing upstream publishing remains disabled. Expanded scanner/build output
lives only on ephemeral runners; no local full build is run below reserve.

The private candidate passes the installed actionlint (1.7.12) and awaits live validation.
Existing strict checks remain required; new checks can only be enforced after
verified execution and source/app provenance. No live workflow/ruleset change
has occurred at this review checkpoint. The excluded workflow remains absent
from the primary filesystem with its original sparse S flag.

Review each change before committing. Nothing has been modified.
