# Contributing to Chartrove

Chartrove is an independent fork of [Fasten Onprem](https://github.com/fastenhealth/fasten-onprem). Keep upstream authorship and copyright notices intact. New issues and PRs belong in [karamhouse/chartrove](https://github.com/karamhouse/chartrove).

## Toolchain

- Go 1.26 or newer; root and desktop modules retain separate dependency graphs.
- Node 26.13.0, pinned in `mise.toml`, with the committed Yarn 1.22.22 configuration.
- Angular 22.x through the local, locked CLI (currently core 22.2.4, CLI/CDK 22.2.5, and TypeScript 6.0.3).
- macOS with Xcode Command Line Tools for native desktop builds.

Install only the locked frontend dependencies when needed:

```sh
cd frontend
npm exec --yes --package=corepack@0.35.0 -- corepack yarn install --frozen-lockfile --registry https://registry.npmjs.org --network-concurrency 4
cd ..
```

Node 26 does not bundle Corepack. The command above invokes a pinned Corepack through npm and preserves the committed Yarn selection. Use `node node_modules/@angular/cli/bin/ng.js <command>` for the local CLI without a global Yarn or Angular CLI installation. The Angular workspace explicitly selects Yarn.

Use `make build-storybook` and `make serve-storybook` for Storybook. These targets
run the local Angular CLI under the same pinned Yarn so Compodoc can invoke its
child command on Node 26.

Angular 22 officially supports Node 26 and TypeScript 6.0. Keep TypeScript within the selected Angular compiler's supported range and report concrete compatibility failures; do not reinstall retired Node runtimes or bypass engine checks. See [the maintenance baseline](docs/MAINTENANCE_BASELINE.md) for validation evidence and the sequential migration history.

## Pull requests and automation

Changes to protected `main` require a pull request with up-to-date successful
Backend, Frontend and Desktop checks and resolved review conversations. No
reviewer approval minimum is configured for the single-maintainer workflow. Force pushes
and branch deletion are blocked on `main`; do not bypass validation to merge.

Only the validation workflow is active. Inherited cloud, CDN, container and
release publishing sources are retained in the
[retired workflow archive](docs/retired-workflows/README.md); the old Development
registration stays disabled. Use short-lived task branches and retire them only
after integration and preservation of any unique history. GitHub issues and the
[consolidation checkpoint](docs/MAINTENANCE_CONSOLIDATION.md) record outstanding
work; closed historical proposals do not imply unimplemented features exist.

## Repository layout

| Path | Purpose |
| --- | --- |
| `frontend/src/app` | Angular views and application services |
| `frontend/src/assets/desktop` | Native bridge adapter |
| `backend/pkg` | Go backend, FHIR handling, database, and HTTP routes |
| `desktop/cmd/fasten-desktop` | Native Wails host entry point |
| `desktop/internal` | Storage, window authorization, bridge, and native smoke checks |
| `desktop/build.sh` | Chartrove app packaging |

Go module/import paths and bridge service names still contain legacy Fasten identifiers. They are compatibility contracts, not display branding; a module migration needs coordinated backend, frontend, and bridge changes and is outside the branding update.

## Desktop development

See [desktop/README.md](desktop/README.md) for builds, flags, and signing. Use an explicitly selected, newly created private directory for testing. Never use the purchased app's storage or an existing health-record database as a fixture.

```sh
sh desktop/build.sh
```

For frontend iteration, select the existing `desktop_prod` configuration. A native executable can serve a built asset directory through `--frontend-dir`; this flag does not select or migrate a database.

## Validation

From the repository root:

```sh
make prepare-backend
go test -p 1 ./backend/... -count=1
cd frontend
node --test desktop-tests/*.test.mjs
node node_modules/@angular/cli/bin/ng.js test --watch=false --browsers=ChromeHeadlessCI
```

The Angular 22 frontend uses production compilation/type checking and Karma
browser tests. The unsupported TSLint/Protractor starter targets are retired with
[their original source and rationale](docs/retired-frontend-tooling/README.md).
A maintained frontend lint or full browser journey command is not currently
configured; do not report the retired commands as passing validation.

Frontend coverage goes to a fresh private temporary directory for each Karma
invocation; the command prints its location. To retain a report at a chosen path,
set `CHARTROVE_COVERAGE_DIR` explicitly. Reports at that explicit location may be
replaced by the run, so select a new report directory rather than a retained
snapshot. Ordinary tests do not write to `frontend/coverage`.

`make prepare-backend` generates the ignored `backend/resources/related_versions.json` required by the backend executable. It does not tidy, vendor, or upgrade dependencies. The backend test targets use this preparation step; dependency refresh remains the separate `dep-backend` target. Mise selects Node through your shell activation; `.envrc` no longer activates the historical Nix flake. The retained flake is not the supported development baseline.

Database fixtures use test-scoped directories and close their repositories before cleanup. Run validation with `TMPDIR` set to a newly created private task directory so any leftovers from failed or interrupted tests can be identified and removed after their processes exit. Never point tests at an existing health-record directory.

From `desktop/`:

```sh
go test ./...
go test github.com/fastenhealth/fasten-onprem/backend/pkg/web -run TestEmbedded -count=1 -timeout=120s
```

The native `--smoke-test` mode requires `--unencrypted` and an explicit fresh `--data-dir`. It refuses an existing database. It validates the real WebView bridge and uses synthetic records only.

During automated GUI checks, verify process exit after closing the test window.
Do not query a closed app through automation that can relaunch it without the
explicit test arguments. Every new test launch must select and verify its
disposable data directory again.

The opt-in PostgreSQL driver regression requires Docker with a local Unix socket.
Pull its pinned PostgreSQL 17 image, then run from the repository root:

```sh
docker pull postgres:17-alpine@sha256:266d3c257594e195b309ae9d1805e54c788f1d7b375f56bf6c367aac6c917e7b
go test -tags=postgres_integration ./backend/pkg/database -run '^TestPostgresIntegration$' -count=1
```

The test creates a loopback-only disposable container with a generated password
and in-memory data directory, and removes it afterward. It accepts no existing
database connection or host volume. To check the desktop dependency graph, run
the same command from `desktop/` with the package path
`github.com/fastenhealth/fasten-onprem/backend/pkg/database`.
This checks driver parameter binding, fresh repository migration, synthetic
account and FHIR date persistence, and reopen. It does not establish compatibility
of every advanced resource query or an existing production database.

Choose checks appropriate to the change. Do not relax timeouts, encryption, or test assertions to make a failing check pass. Distinguish a source check from a successful build and a real native UI check.

## Server signing key

See [local data and security settings](docs/LOCAL_DATA_SAFETY.md) for the different desktop/server defaults and explicit development options.

Server startup rejects an empty key or the historical shared JWT signing key
before opening a database. Configure `FASTEN_JWT_ISSUER_KEY` or `jwt.issuer.key`
with a private value unique to the installation. Generate a new value with
`openssl rand -base64 32` and retain it using your existing secret-management
process. Preserve an existing private signing key; this change does not rotate
keys or modify stored records. The desktop retains its per-installation key.
Startup errors name the setting without printing its value. Configuration dumps
and signup-token console logging are removed.

## Branding, data, and provider boundaries

Use Chartrove for product-facing names and link project support to this repository. Preserve Fasten attribution and external service identities. [BRANDING.md](docs/BRANDING.md) records the compatibility names that remain intentionally stable.

Leave `healthit.yaml` and existing health records untouched. Keep `fasten.db`, existing storage paths, key management, and migration behavior stable unless a separately scoped data migration has been approved.

The Epic sandbox connection uses published synthetic test data and upstream sandbox infrastructure. It does not grant production access. See [the provider validation record](desktop/PROVIDER_VALIDATION.md) and [Lighthouse replacement assessment](desktop/LIGHTHOUSE_REPLACEMENT.md).

## Pull requests

Keep changes focused and describe the resulting behavior and validation. Preserve commit attribution when adopting upstream contributions. Migrated draft PRs target historical bases to preserve their original comparisons; review and retarget/rebase them before merging. Existing review transcripts are historical context, not approvals for this fork.

The [upstream Fasten documentation](https://github.com/fastenhealth/docs/tree/main/technical) is useful historical context. Its hosted-service, deployment, and provider-registration instructions may be outdated for Chartrove.
