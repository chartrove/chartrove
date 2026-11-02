# Retired workflow audit

| Severity | Finding groups |
| --- | ---: |
| High | 3 |
| Medium | 2 |
| Low | 1 |

The six files are byte-identical copies from main `2d4c2c197`; the `.disabled`
extension and location prevent GitHub execution. Their registered workflow IDs
were also disabled before repository Actions enablement. This retirement is
part of the explicitly authorized validation-only CI replacement.

- High: Cloud-Deploy delegates AWS access and syncs to the upstream Fasten bucket
  (`cloud-deploy.yaml.disabled`, role-to-assume and aws s3 sync). Docker grants
  package/OIDC writes and publishes via inherited Depot. CDN pushes to fastenhealth/cdn.
  Replacement: retire these operations; validation has only `contents: read`.
- High: CDN uses `cpina/github-action-push-to-another-repository@main`; Release
  uses three `packagrio/*@master` actions. Other publishing actions use mutable
  tags. Replacement: no publishing actions; all validation action refs use full
  commit SHAs verified from the official action repositories.
- High: Release mints an app token and commits/publishes with tests commented out.
  Replacement: no release/app-token workflow. A future release process needs review.
- Medium: CI/Development/CDN/Release inherit unspecified token scopes. Replacement:
  explicit read-only contents permission and no secret inputs.
- Medium: CI downloads artifact `coverage`, but Development emits frontend-coverage
  and backend-coverage, then publishes to Codecov/Chromatic. Replacement: run tests
  directly without external reporting or uploads.
- Low: inherited checkout steps persist credentials. Replacement:
  `persist-credentials: false` on every validation checkout.

Validation uses ordinary pull_request/push/workflow_dispatch, GitHub-hosted
runners, fixed Node/Go versions, no privileged triggers, no shell interpolation
of contributor-controlled content, and no secrets or persistent runner state.
It does not claim whole-application security or native Angular UI acceptance.
