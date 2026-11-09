# Legacy provider HTTP rejection-body privacy

## Draft scope and ownership

Pi owns #496 in the sole checkout, branch `task/legacy-provider-response-privacy`,
from accepted `5bf2ca0ff56806f26e28f8429b688bdc18ed12a6`. Validated local product
checkpoint: `65d7b54fa2e60c549d3db9d344c14756faebd684`; protected publication remains
pending at this checkpoint.
The preserved `task/provider-response-privacy` name belongs to historical #471, not
this task; a failed branch-name collision changed no source or branch ownership.

This is a bounded HTTP non-success response boundary for the existing pinned SDK,
not a replacement SDK, dependency refresh or whole-provider privacy certificate.

## Actual baseline and repair

The pinned SDK's non-dynamic refresh logs OAuth errors through the standard logger;
those errors include arbitrary provider descriptions. Its resource-request path also
reads rejected bodies before clipping the returned string. Three actual SDK
400/401/500 synthetic refresh cases genuinely failed 0.586s: returned errors and
standard-log output contained the description canary. Safe `invalid_grant`, credential
preservation and response-close controls already passed. No real endpoint, credential,
provider discovery, database or health record was used; an in-memory transport received
every request.

`models.WithProviderResponsePrivacy`, appended after other HTTP-client options:

- Copies the configured HTTP client and wraps its transport without mutating global
  clients/transports, SDK/module-cache source or application/standard-log configuration.
  Explicit test clients are copied too; the option never enables SDK test mode.
- For non-2xx replies, explicitly reads at most 4096+1 bytes and closes the original.
  Only an exact, unambiguous allowlisted OAuth error name survives. Descriptions,
  arbitrary codes, malformed/oversize bodies, duplicate/case-aliased error fields and
  trailing JSON become fixed `provider_request_failed`. JSON and the pinned OAuth
  client's supported form/text response formats retain safe classifications.
- Replaces failed-body metadata and provider-controlled status reason before SDK
  diagnostics. Status code remains. Other headers, including redirect locations, are
  not silently removed; this is not header sanitization or a redirect-policy change.
- Returns successful response objects/bodies untouched: no read, close, JSON conversion,
  floating-point roundtrip, decimal/clinical normalization or token alteration.
- Preserves the supplied context as parent, rather than inventing Background or copying
  Gin/JWT. Six current source-client factory sites adopt the option: regular sync,
  manual probe/import, related import and two default-off diagnostics.

The read bound describes this wrapper's explicit reads per reply, not aggregate retry
traffic or arbitrary underlying transport `Close` behavior. Existing caller/client/SDK
latency policies remain; no new global deadline or scheduler policy was adopted.

## Qualification at this checkpoint

- Original three actual-SDK assertions pass with the option, not a fake OAuth failure.
  Actual SDK rejected FHIR requests also retain status, preserve outbound bearer use,
  close the body and reject rather than project the response, without echoing its body.
- Existing real OAuth compatibility uses the same option: valid-token reuse, refresh
  rotation, omitted-refresh preservation, invalid-grant rejection. Added form/text
  rejection cases preserve the same safe classification. Original assertions remain.
- Allowlisted/unknown/ambiguous/malformed JSON and forms, success status/byte identity,
  exact decimal lexeme, read bound, status-reason replacement, response-object isolation,
  retained location header, unchanged transport errors and copied client settings are
  covered. Parent deadline/owner value/cancellation and explicit test-client copying
  are checked at the option/transport boundary.
- First selected SDK/compat/dynamic-regression race passed 17.974s; expanded selection
  22.590s. Final complete models race passed 79.803s; existing handler credential-log
  controls passed race 7.480s. Handler compilation 1.510s explicitly ran no tests.
  Changed models/handler vet and diff checks passed.
- Complete serialized ordinary models/handler/web validation passed
  4.630/226.193/103.613s, captured Go-test exit 0 and unchanged 15-minute package/fixture
  budgets. All eight tested source hashes match before/after validation; source was
  frozen through actual exit capture, with documentation edited separately.
- Native cmd/host/shell/SMART/smoke compiled 0.706/2.148/1.615/0.210/0.158s with
  explicitly no tests run. Existing macOS object/deployment-target warning scope is
  retained, not an older-OS/runtime certificate. No full handler/web race, native
  runtime, hosted publication or actual provider acceptance is claimed yet.

## Explicit residual scope

SDK request URLs, response headers, successful diagnostics/resource IDs, arbitrary
transport errors, 2xx OAuth error responses, all local job/error persistence and
historical logs are not certified. Preserving the constructor context does not prove
that every SDK-manufactured HTTP request uses its lifetime; the legacy GetRequest
constructor is a separate source lead for #474. No global logger redirection, historical
cleanup, credential revocation, zeroization, provider consent or live registration is
claimed. #474 global inventory, #479 maximum race, #487 native runtime/negative gates,
normal GUI capture and independently operated recovery remain open.
