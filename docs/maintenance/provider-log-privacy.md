# Provider credential request log privacy

The production-readiness reassessment found that the legacy create/reconnect source handler logged the
entire decoded `SourceCredential` at INFO. That structure contains access/refresh/ID tokens, PKCE verifier,
private dynamic-client JWK fields, Patient identity and display text. A rejected request could therefore
copy secrets into logs before provider lookup or storage. Raising the level to INFO did not protect them.
Additionally, JSON binding errors for invalid timestamp fields echoed submitted values into both logs and
the error response.

## Correction and boundary

`CreateReconnectSource` no longer logs decoded credential payloads. Binding failures retain a fixed parsing
diagnostic and HTTP 400 without raw decoder errors. Catalog lookup and other existing failure diagnostics
remain unchanged. This does not change provider registration, successful request/response contracts,
credential persistence, source ownership, grants or synchronization. Existing API credentials are not
rotated, and historic logs are not inspected or deleted.

This is a focused correction, not an assurance that every log path is safe. Legacy dynamic-client
registration/token error-body logging in `backend/pkg/models/source_credential.go` is addressed by the
[separate response correction](provider-response-privacy.md); its acceptance is recorded there rather
than inferred from this handler fix. Ingress, proxy, SDK, browser and identity-provider logs require their
own acceptance. No real provider request, real token or medical record was used.

## Synthetic regression evidence

The permanent tests invoke the real handler with synthetic decoded credentials and a missing local-catalog
endpoint, forbidding database calls. They check every private canary against logs and responses at DEBUG,
INFO and WARN, while requiring the catalog-error diagnostic to remain. A second test exercises actual
invalid-timestamp binding rather than mocking the parser.

Before correction, the original assertions failed on the INFO/DEBUG credential payload and malformed-date
log output. WARN remained a closed control. After correction, the unchanged tests passed under `-race`:

```sh
go test -race ./backend/pkg/web/handler \
  -run '^TestCreateReconnectSource(BindingErrorsDoNotEchoInput|DoesNotLogCredentials)$' \
  -count=1 -timeout=60s
```

Local execution used synthetic fixtures, disabled dependency downloads (`GOPROXY=off`, `GOSUMDB=off`) and
the private task temporary directory. It is ordinary project regression validation, not an independent
sandboxed penetration-test certification. A restrictive macOS sandbox probe could not even execute
`/bin/echo` (exit 134), so no sandbox qualification is claimed. An initial race-build invocation exceeded
the tool's 120-second wall timeout before a test result; the warmed rerun completed successfully
(`handler`, 5.473s). Exact raw failure/success logs are retained privately, not uploaded as medical evidence.
Full hosted validation remains required before integration.
