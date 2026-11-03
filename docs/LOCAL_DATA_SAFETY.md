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
fallbacks bind 0.0.0.0, disable HTTPS and disable SQLite encryption. The repository
[config.yaml](../config.yaml) explicitly enables HTTPS and SQLite encryption.
Neither is evidence that an arbitrary deployment is ready for health records.
Review the effective network and data settings before running a standalone server.
PostgreSQL transport/storage protection is configured separately; SQLite's
application encryption option does not establish PostgreSQL encryption.

Standalone startup rejects an empty JWT signing key and the historical shared
example before opening storage. The example configuration leaves the key empty.
Supply FASTEN_JWT_ISSUER_KEY or a private jwt.issuer.key unique to the installation.
The [contributor signing-key instructions](../CONTRIBUTING.md#server-signing-key)
show how to generate one. Preserve an existing private signing key; changing it
invalidates existing sessions. Do not put keys or configuration dumps in reports.
The desktop already manages its persistent installation signing key.

The existing FASTEN_WEB_LISTEN_HTTPS_ENABLED setting selects built-in HTTPS or
HTTP. Enabled HTTPS uses the server's certificate generation/lookup path. HTTP
behind a separately managed TLS terminator is an operator-managed deployment
choice; do not expose an unprotected HTTP endpoint to an untrusted network.
This baseline does not install a proxy, configure trust roots, deploy a container,
or establish production provider registration.

[Imported discussion #193](https://github.com/karamhouse/chartrove/issues/193)
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

Public [Chartrove reports](https://github.com/karamhouse/chartrove/issues) must
omit health information, identifying details, passwords and keys. Use a synthetic
reproduction. Fasten attribution, licenses and external service identities remain
as recorded in [the project](../README.md#fasten-attribution-and-license).
