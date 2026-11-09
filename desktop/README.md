# Chartrove desktop

Chartrove is an independent fork of Fasten Onprem, with a revived macOS Wails v3 host around the fork's frontend and backend. This is a local development build, ad hoc signed and not notarized. It does not replace the purchased Fasten app.

## Build

Use Go 1.26 or newer, an Apple Silicon Mac with Xcode Command Line Tools, and Mise Node 26.13.0. The separate desktop Go module pins Wails to v3.0.0. The frontend is on Angular 22 with the committed lockfile. See [the contributor toolchain instructions](../CONTRIBUTING.md) for the pinned Yarn launcher and current validation limits.

With Node 26.13.0 selected in your shell:

```sh
cd frontend
npm exec --yes --package=corepack@0.35.0 -- corepack yarn install --frozen-lockfile --registry https://registry.npmjs.org --network-concurrency 4
cd ..
sh desktop/build.sh
```

The result is `dist/Chartrove.zip`, containing `Chartrove.app` for macOS Apple Silicon only. The script rejects other hosts and conflicting Go targets before starting a build, explicitly enables CGO for darwin/arm64, and verifies the executable's architecture before signing. Signing takes place in an OS temporary directory because macOS File Provider can attach metadata to `.app` directories inside Documents that prevents signing. The ZIP preserves the signed bundle. The build does not install the app or change system trust settings. The app menus, window titles, package name, and Atlantic blue icon use Chartrove branding. Fasten attribution remains in the app footer and project documentation.

The app includes the unchanged repository `LICENSE.md` alongside its existing Fasten attribution. The Desktop CI job also builds this ZIP, extracts it privately, verifies its arm64 executable, ad hoc signature and exact license bytes, and runs `--help` with an explicit unused data path. It asserts that the path was not created: help exits before opening configuration or health-record storage. ZIP export is optional as described below. CI does not notarize the app, launch its WebView, or publish a release. Downloadable distribution and real operational acceptance remain separate gates.

The same bundle contains the independent arm64 operator utility `Contents/MacOS/fasten-backup`, separately ad hoc signed before bundle signing. It is never run automatically by the app. See the [encrypted SQLite backup and isolated recovery contract](../docs/maintenance/encrypted-sqlite-backup.md) for explicit database/output/key-file inputs and custody requirements; it does not load app configuration or connect to an identity provider. It is a full-instance backup tool, not a patient export or administrator API, and it does not support PostgreSQL recovery.

Desktop CI qualifies this exact extracted utility using only a fresh synthetic encrypted database with committed WAL state: correct-key backup, ciphertext receipt/hash, standalone recovery and integrity, reopening, wrong-key rejection, and ignored configuration tripwires. The tagged qualification fails if the binary or selected test is absent; no skipped case counts as acceptance. This artifact check does not repeat or replace the broader source recovery regressions, verify production custody/restore, install or publish the app, or accept #425.

After a successful frontend build, `sh desktop/build.sh --skip-frontend` can rebuild only the native package when frontend inputs have not changed. It uses the existing `dist/frontend` output. A normal build regenerates frontend output and synchronizes the embedded assets to remove stale files.

### Version-bound inspection packages

Validation records the qualified ZIP's source SHA, archive SHA256, byte count and
workflow reference after the existing native packaging/backup checks. Ordinary
PRs, main pushes and manual runs with default inputs export nothing.

A separately authorized operator can request a seven-day GitHub Actions artifact
through manual Validation on `main`: `export_desktop=true` and
`artifact_source_sha` equal to the exact accepted main revision. The first run
attempt is required; mismatched refs/revisions and reruns are rejected before
building. The input binds the request and does not itself confer authorization.
Actual artifact export remains a separate owner decision after source acceptance.

The allowlist is `Chartrove.zip`, `Chartrove.zip.sha256` and `Chartrove.build.json`.
Extraction directories, test databases, recovery-key files and logs are not upload
inputs.
Artifact names include source/run identities, replacement is disabled, and access
uses GitHub's repository artifact permissions. Verify the source, complete workflow
outcome and archive checksum before an authorized operational exercise. The build
retains ad hoc signing. Developer ID/notarization, installation, provider access,
recovery exercises, release publication and deployment have separate acceptance
and authorization requirements.


## Data and startup

[Local data and security settings](../docs/LOCAL_DATA_SAFETY.md) distinguishes desktop defaults, inherited server settings and explicit synthetic-test options.

The default data directory on macOS is `~/Library/Application Support/FastenCommunity`. The host creates it privately, keeps a persistent JWT signing key in a mode-0600 file, and locks it against a second desktop instance. This compatibility directory name is intentionally retained through the Chartrove rebrand, together with bundle identifier `house.karam.fasten-community`, database filename `fasten.db`, key files, and bridge service names. No data migration is performed. This directory is separate from the purchased application's storage.

Database encryption is enabled by default. The fork's existing encryption-key setup and restore screens handle creation and unlock. Save the recovery key when prompted; the host does not persist it. Existing purchased-app databases and their migrations have not been validated and must not be used as test inputs.

For a separate development database, launch the extracted executable with an explicit private directory:

```sh
Chartrove.app/Contents/MacOS/chartrove --data-dir /path/to/private/test-directory
```

`--unencrypted` is available only for disposable synthetic-data testing. `--frontend-dir` selects a development asset directory. Normal startup embeds the fork's `/api` routes in Wails without opening an application API TCP listener or requiring hosted installation registration.

The [direct SMART connection UI](DIRECT_SMART.md) uses a separate Go-host service for Chartrove-owned public registrations, system-browser authorization with S256 PKCE, and in-memory Patient retrieval. It is disabled unless `--smart-provider-config` selects an explicit public JSON profile file. An active authorization temporarily opens only its registered IPv4-loopback callback listener. Sources shows a Patient preview before an explicit import saves a provider-scoped manual snapshot. Cancellation, discard, and screen teardown release the session. Snapshots do not automatically synchronize. Keychain persistence and refresh remain future work; no production registration is supplied.

The native bridge uses the runtime served by the pinned host. It validates external URLs, restricts privileged actions to the main window, and restricts legacy callback completion to the popup for the matching state. Legacy OAuth callback data is delivered directly to the main window, not broadcast to other provider windows. Epic sandbox OAuth and a synthetic Patient read have been verified in an isolated native test frontend. The legacy production workflow still depends on retired upstream infrastructure; see [provider validation](PROVIDER_VALIDATION.md) for its tested boundaries. Direct SMART avoids that dependency but requires Chartrove-owned registrations and live interoperability validation. Hosted CCDA conversion remains unvalidated.

## Tests

The dated observations below describe earlier Fasten Community builds. They are historical evidence, not validation of a newly branded Chartrove package. See [current Chartrove branding validation](BRANDING_VALIDATION.md) for the new package checks.


From the repository root:

```sh
go test ./backend/pkg/web ./backend/pkg/event_bus -count=1 -timeout=120s
go test -race ./backend/pkg/web ./backend/pkg/event_bus -count=1 -timeout=120s
cd frontend
node --test desktop-tests/bridge.test.mjs
node node_modules/@angular/cli/bin/ng.js test --watch=false --browsers=ChromeHeadless
```

From `desktop/`, run `go test ./...` for host tests. To exercise the embedded backend against the desktop module's actual dependency graph, also run:

```sh
go test github.com/fastenhealth/fasten-onprem/backend/pkg/web -run TestEmbedded -count=1 -timeout=120s
```

The native executable supports `--smoke-test --unencrypted --data-dir <fresh-private-directory>`. Smoke mode refuses a nonempty or final-component symlink data directory, an omitted explicit data directory, or configured direct SMART providers. It tests the real WebView/Go binding, disabled direct SMART service, unconfigured-provider rejection, fresh database API, synthetic signup/sign-in, wrong-receiver rejection/no source-job mutation, owner-confirmed native multipart import, URL rejection, JSON API errors, popup callback delivery, and isolation from an unrelated window. It prints a `DESKTOP_SMOKE` JSON result, quits automatically, and exits nonzero on failure. Its local test routes do not exist in normal mode.

Native completion also publishes `desktop-smoke-report.json` in that selected private directory.
The bounded (8 KiB), owner-only receipt is published only after its temporary file is written,
synced and closed. Publication refuses an existing file or report symlink; it never overwrites
prior evidence. Failed private temporary evidence is retained. A failed receipt write makes the
smoke run fail; normal startup and `--help` do not write a receipt. The receipt records native test
completion, not independently verified process exit or a power-loss durability guarantee.

For non-interfering automated runtime qualification, add `--smoke-background` to the existing smoke
arguments. It is rejected without `--smoke-test` and keeps the main, callback and observer windows
hidden. It does not grant capture permission, enable providers or use a foreground fallback. A
hidden native WebView/service test is **not** visual/interactive GUI, encrypted-GUI, key-custody or
production acceptance; those gates remain separate. Current exact-source positive/negative native
package evidence is in the [runtime receipt](../docs/maintenance/native-smoke-runtime.md).
The native file adapter uses the standard multipart encoder with a 32 MiB encoded-envelope/16-part
limit and 30-second preparation budget; normal HTTP and external multipart remain unchanged.
The smoke deadline stays 45 seconds. Up to twelve best-effort, fixed-label progress posts allow a
failed timeout receipt to retain acknowledged stages, not records/tokens or a success claim.
See the [manual receiver receipt](../docs/maintenance/manual-import-receiver.md) for actual native
file-path failures, source-bound twelve-check success and outstanding later runtime/race gates.

Observed on October 26, 2026: seven embedded-backend integration tests passed, including encrypted create/unlock/reopen, bad-key rejection, concurrent health polling, persistence, and synthetic FHIR import without outbound HTTP or temporary-document residue. The cancellation regression passed with the real event bus. Desktop host tests and four bridge tests passed. The frontend suite reported 286 successes and one skipped test. The packaged executable passed all eight native smoke checks after extraction and passed strict signature verification. Its encryption onboarding UI was also verified visually; a staged build's actual UI signed into the synthetic account after restart and reached the dashboard.

On October 27, the original fixture-import timeouts were reproduced and fixed: generated models now share compiled JavaScript programs, acquire exclusive initialized runtimes, clear their current resource before returning them, and reuse parsed FHIRPath expressions without caching patient results. The generator was updated with its 56 model outputs. The graph suite passed in 16.596 seconds and resource handler suite in 11.887 seconds with the same 120-second limit. Concurrent resource-isolation tests passed under the race detector. The stale generated database mock was also regenerated with the pinned MockGen version after the broader handler suite exposed missing interface methods.

The full database package performs many additional migrations and imports; applying a 120-second limit to that entire package still times out later in unrelated tests. Full-suite validation uses the repository's usual `go test ./backend/... -count=1` command with Go's default timeout. That full suite passed; the database package took 253.471 seconds and the handler package 31.149 seconds. `go vet ./...`, desktop module tests, and embedded-backend tests using the desktop dependency graph also passed. No test assertions, fixtures, encryption settings, or repository timeout configuration were changed. Missing generated `related_versions.json` from the initial revival was resolved with the existing targeted generator.

## Restored dependency provenance

The fork requires `github.com/fastenhealth/fasten-sources v0.6.25`, whose original repository and Go proxy archive were unavailable. A surviving public mirror retains release commit `18c9783658fdd109483dac88ae55cb13a4b68408`. The replacement is constrained to that exact original version and pins `github.com/andhines/fasten-sources v0.0.0-20241106031723-18c9783658fd`.

Rehashing the mirror archive using the original module/version prefix produced the exact checksum already recorded by this fork:

```text
h1:PpgI/jkd3a8y9gYMX5tLdmrAXu8b8KhIU97rek88QCY=
```

Its `go.mod` checksum also matched the original. This restores the pinned release's contents rather than substituting a newer provider library. Both root and desktop module files retain the explicit replacement for reproducibility.

`healthit.yaml` was excluded throughout implementation. The installed purchased app and its medical records were not read or changed. Test databases, temporary apps/downloads, and the community app's newly created synthetic WebView cache are disposable test state; they are removed after validation. No changes were pushed.
