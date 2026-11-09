# Local data and security settings

Chartrove is a development preview. This guide describes current behavior and
explicit development options; it does not certify a production installation.
Start with the [project status](../README.md#project-status) and use disposable
synthetic records for development. Existing family records and the purchased
Fasten application's storage are not test fixtures.

## Desktop defaults

The native host embeds the API inside Wails. Normal desktop mode opens no
application API TCP listener and does not require importing a server CA into a
browser. Its defaults are:

| Setting | Desktop behavior |
| --- | --- |
| SQLite encryption | Enabled; keep the recovery key when setting up encryption |
| Database recovery key | Not persisted by the host |
| JWT signing key | Random per installation, persisted in a private mode-0600 file |
| Storage directory | Private mode 0700, with an exclusive instance lock |
| Unsafe unauthenticated routes | Disabled |
| Direct SMART providers | Disabled until an explicit public provider profile is selected |

The compatibility directory remains FastenCommunity. Signing keys and database
recovery keys serve different purposes. Preserve existing private keys and storage
paths; the fork does not migrate or rename existing databases. Read the
[desktop data guide](../desktop/README.md#data-and-startup) and
[provider validation status](../desktop/PROVIDER_VALIDATION.md#current-validation-status)
before selecting a data directory or attempting a provider connection.

For the explicit operator-only encrypted SQLite backup command and isolated
synthetic restore contract, read the [backup guide](maintenance/encrypted-sqlite-backup.md).
It captures committed WAL state into a new encrypted artifact; it does not back up
keys/configuration, restore an installation automatically or authorize live data use.

Use the ordinary account-creation/sign-in flow. A public synthetic test account
used by validation is not a default account for a normal installation. Current
startup instructions do not publish a shared beta username/password.

## Explicit development options

For disposable synthetic records, the desktop accepts --unencrypted together
with an explicitly selected fresh private --data-dir. The
[README test command](../README.md#desktop-development) shows this separate test
workflow. It does not authorize opening or converting an existing database.
--frontend-dir selects development assets; it does not change the data boundary.

Keep web.allow_unsafe_endpoints disabled. Enabling it exposes developer routes
that bypass normal authentication. This guide does not enable them or recommend
them for records. The desktop explicitly forces the setting off.

## Inherited standalone server settings

The standalone backend has a different startup/configuration path. The factory
fallback and repository [config.yaml](../config.yaml) now select IPv4 loopback
`127.0.0.1`. Empty/blank host settings also resolve to loopback rather than an
implicit wildcard. Explicit `web.listen.host`/`FASTEN_WEB_LISTEN_HOST` addresses,
including deliberate wildcard/IPv6 settings, remain operator choices. Existing
private configuration files are not rewritten. Other local users, proxies,
forwarders and explicit external publication are separate boundaries.

Factory HTTPS and SQLite encryption fallbacks remain disabled; the repository
sample explicitly enables both. With a valid private signing key but no configured
encryption key, standalone startup enters its existing STANDBY setup flow before
opening the repository, rather than proving that an encrypted database is usable.
Neither configuration is evidence of production readiness. Review effective network,
key custody and data settings before running a server. PostgreSQL transport/storage
protection is separate; SQLite's encryption option does not protect PostgreSQL.

The image build derives an explicit container-side wildcard bind from the single
public standalone sample using [the guarded producer](../scripts/container-config.sh).
No image environment default overrides an operator-mounted configuration. Both
Compose examples publish only `127.0.0.1:${PORT:-9090}:8080` by default; intentional
remote publication uses `FASTEN_BIND_HOST` (an IPv6 value in short port syntax needs
brackets, for example `[::1]`). This host-publication choice is distinct from the
container's `FASTEN_WEB_LISTEN_HOST`. Remote companion pairing needs deliberate
network/proxy configuration, not merely a QR code. The legacy production Compose
file still selects an upstream image: this change does not advertise, build, deploy
or certify a Chartrove production image, migrate records, or configure TLS/firewalls.
See [the scoped defaults receipt](maintenance/standalone-loopback.md).

Standalone startup rejects an empty JWT signing key and the historical shared
example before opening storage. The example configuration leaves the key empty.
Supply FASTEN_JWT_ISSUER_KEY or a private jwt.issuer.key unique to the installation.
The [contributor signing-key instructions](../CONTRIBUTING.md#server-signing-key)
show how to generate one. Preserve an existing private signing key; changing it
invalidates existing sessions. Do not put keys or configuration dumps in reports.
The desktop already manages its persistent installation signing key.

Session lifetime is configured per installation with `jwt.session.duration` or
`FASTEN_JWT_SESSION_DURATION`, for example `5m` or `2h`. The default remains one
hour; an empty setting also uses that default. Values use Go duration units and
must be at least `1s`. Invalid settings prevent standalone or embedded startup
before storage initialization and cannot create an account through sign-up.
This changes the absolute expiry of newly issued sign-in/sign-up tokens. Existing
tokens keep their signed expiry, including after a restart or setting change.
It does not add a per-user preference, idle timeout, sliding refresh or revocation
mechanism, and does not change persistent signing/encryption keys or stored records.

The existing FASTEN_WEB_LISTEN_HTTPS_ENABLED setting selects built-in HTTPS or
HTTP. Enabled HTTPS uses the server's certificate generation/lookup path. HTTP
behind a separately managed TLS terminator is an operator-managed deployment
choice; do not expose an unprotected HTTP endpoint to an untrusted network.
This baseline does not install a proxy, configure trust roots, deploy a container,
or establish production provider registration.

[Imported discussion #193](https://github.com/chartrove/chartrove/issues/193)
preserves an attributed community Tailscale/Docktail recipe for an upstream
Fasten image. It is reference material, not a validated Chartrove deployment or
a newly implemented integration. The original discussion and its custom-certificate
comments remain available; no custom-certificate API is added by this guide.

## Source and validation boundaries

The current defaults and startup guards are visible in
[desktop startup](../desktop/cmd/fasten-desktop/main.go),
[private desktop storage](../desktop/internal/host/storage.go),
[configuration](../backend/pkg/config/config.go),
[server startup](../backend/pkg/web/server.go) and
[existing signing-key regressions](../backend/pkg/web/signing_key_test.go).
The shared-key guard is an existing integrated fix, commit
6ac39837606de0b6012b40db76f0f1e3efb5bc56; this guide does not rotate data or keys.
Validation uses unsupported synthetic database types for startup-key checks,
private test directories, local fixtures and blocked outbound HTTP.

Public [Chartrove reports](https://github.com/chartrove/chartrove/issues) must
omit health information, identifying details, passwords and keys. Use a synthetic
reproduction. Fasten attribution, licenses and external service identities remain
as recorded in [the project](../README.md#fasten-attribution-and-license).

Repository constructors disable automatic GORM SQL output at every application
log level, including DEBUG. SQL parameters, driver errors and slow-query traces
can contain clinical text and credentials; selected parameter filters alone do
not cover all query paths. Application lifecycle messages and errors returned to
callers remain available. This does not certify every application, proxy, browser
or identity-provider log. Use synthetic fixtures for database diagnostics; do not
enable SQL tracing on health records. The
[daily-use readiness assessment](maintenance/production-readiness-2026-11-05.md)
separates the source correction from unverified backup and runtime acceptance.
