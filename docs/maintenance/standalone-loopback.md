# Standalone loopback and explicit container networking

Pi owns [#494](https://github.com/chartrove/chartrove/issues/494), branch
`task/standalone-loopback`, in the canonical single checkout. This is a source/default
change, not deployment, migration, key rotation or a production-readiness certificate.
Validated product checkpoint: `0da518aea06696cc906171e0ecd8ab4d263f5eab`.

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
standalone sample. A new public-script marker extends the existing synthetic Docker
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
older-platform, Docker runtime or deployment acceptance. Protected publication is pending.
