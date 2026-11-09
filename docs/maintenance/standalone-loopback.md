# Standalone loopback and explicit container networking

Pi owns [#494](https://github.com/chartrove/chartrove/issues/494), branch
`task/standalone-loopback`, in the canonical single checkout. This is a source/default
change, not deployment, migration, key rotation or a production-readiness certificate.
Validated product checkpoint: `0da518aea06696cc906171e0ecd8ab4d263f5eab`.
Normal accepted #493 parent integration: `7a57a5ac0b849fc77a73405c81170f26dddf4c00`;
all ten configuration/producer/address/test source hashes stayed unchanged. Integrated
focused config/web races passed 2.760/3.326s.

## Effective defaults, not one isolated fallback

Both the configuration factory and committed `config.yaml` now select `127.0.0.1`.
The CLI automatically reads the working-directory sample, so changing only the factory
would not change effective sample/CLI behavior. Address construction treats empty/blank
hosts as loopback, uses the standard host/port join, and preserves explicit IPv4,
hostname, wildcard, IPv6, previously bracketed IPv6 and zone forms. No baseline test
opens a wildcard TCP socket.

The image needs a deliberate **container-side** bind for ordinary port forwarding.
Its build derives the container configuration from the single public standalone sample
using `scripts/container-config.sh`, replacing exactly one expected loopback line.
No image ENV default was adopted: Viper environment precedence would override an
operator's explicitly mounted loopback configuration. The producer refuses missing/
ambiguous source markers and existing output; it does not edit the input or user config.
Other YAML bytes, HTTPS/encryption choices and empty key placeholders stay unchanged.
The final stage copies this derived build output instead of overwriting it with the
standalone sample. The existing nonpublishing container workflow now also triggers on
`config.yaml` and `scripts/container-config.sh` changes for PRs and main; permissions,
pinned checkout, matrix, no-publication steps and 30-minute budget are unchanged.
This narrow dependency-filter change passes installed `actionlint`; it is not a full
workflow security audit. Both affected platform jobs are required alongside the seven
usual checks for this task's publication qualification. A new public-script marker extends the existing synthetic Docker
context regression; private context exclusions are unchanged.

Both Compose examples default **host-side publication** to
`127.0.0.1:${PORT:-9090}:8080`, retaining deliberate `FASTEN_BIND_HOST` overrides.
This is separate from `FASTEN_WEB_LISTEN_HOST` inside a container. An IPv6 host value
in short publication syntax needs brackets. Existing running containers and private
configs are not modified; later operator-initiated recreation will use the new defaults
unless explicitly overridden. Remote access/pairing and reverse proxies need deliberate
network configuration. The legacy production Compose image reference is retained, not
endorsed/refreshed into a Chartrove production release.

## Evidence and limits

- Factory/sample baseline: two genuine expected-loopback/actual-wildcard failures
  (`config` package 0.185s). The initially proposed missing Docker ENV assertion was a
  candidate-design check, not a demonstrated runtime defect; that approach was rejected.
- A behavior-preserving address-construction extraction made the actual server address
  testable without storage/listeners. Five genuine reds (`web` 1.963s): default, empty,
  blank and unbracketed IPv6 loopback/wildcard. Existing explicit/bracketed controls stayed
  available. Fourteen final address cases include empty brackets and IPv6 zones.
- Compose's old implicit-publication source contract failed in 0.292s; both directly
  coupled examples now keep default publication private without changing their images.
- Initial focused ordinary config/web 1.353/1.901s and race 3.713/8.030s passed. Expanded
  config/web race 2.057/3.329s includes producer preservation/failure/help/usage, sample,
  explicit environment/programmatic choices, Compose source contract and address controls.
- Installed Docker Compose v5.7.0 parsed both files with default, explicit wildcard and
  bracketed IPv6 override: six successful expected host/9090/8080 results. An explicit
  empty env-file and synthetic values avoided reading local `.env` data; a deliberately
  absent Unix Docker socket prevented daemon operations. This is parsing/interpolation,
  **not an actual network exposure or production-image test**.
- `shfmt` and `shellcheck` pass the producer. An initial formatting gate stopped before
  Go execution; it was not a behavioral test red. No dependencies/toolchains were updated.
- Complete ordinary config/web validation passed with actual wrapper exit 0:
  0.619/329.951s, serialized packages and unchanged fixture budgets. Full config race
  passed 1.721s. Changed configuration/web vet and diff checks passed. This is not a
  fresh complete web race or native GUI certificate.
- The reviewed Docker context regression passed original/current synthetic controls with
  the new producer-script marker: local Unix socket/context builder, `FROM scratch`, no
  RUN/registry/runtime container/production context. Only disposable fixture outputs were
  removed normally; context settings and unrelated containers were untouched. This does
  not build or run the actual application image.
- Go/configuration/producer source hashes match before and after complete validation.
  No budgets, bcrypt cost, global process settings or dependency versions changed.

Private raw evidence is not committed. Existing first-run STANDBY behavior matters:
with a valid private signing key but absent encryption key, startup enters setup before
repository creation; the repository-level missing-key error does not prove startup
aborts. No existing-record disclosure is demonstrated here. Signing-key guards,
registration/auth/consent, encryption policy and native no-API-TCP behavior are unchanged.
Loopback does not protect against other local users, forwarding, explicit exposure or
badly configured proxies, and does not prove at-rest protection, TLS, recovery,
older-platform, running-application network or deployment acceptance.

## Protected acceptance and independent main qualification

Exact `2c0ae15ae67c16580119daa3a6c6b44839b917e4` passed all seven42701210530 plus affected
native-platform image builds42701211447: amd64 9m08s, arm64 6m51s. Fresh matching head,
CLEAN/no unresolved threads and ready verification preceded protected #495 merge
`5bf2ca0ff56806f26e28f8429b688bdc18ed12a6` at 2026-11-09T21:35:44Z. Separate exact-main
42703989021 and both container jobs42703989082 passed. #494 is closed boundedly.

PR image IDs were amd64 `sha256:ddf3502fb71f20b3437f7b5c403288804034c7f4c7776a310c4234da57b116bf`
and arm64 `sha256:e456a7bcc44248393e02f2b44fa67748982244f1d2e0e05b13b7f0c418e22614`.
These are actual source-bound built-image identities, not registry/release publication,
bit-for-bit reproducibility, a running app/network test or deployment evidence. Superseded
e624 runs were not relabeled as new-head passes after the substantive trigger-filter fix.
