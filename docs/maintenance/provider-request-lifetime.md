# Legacy SDK request lifetime

## Draft ownership and scope

Pi owns #500, a bounded follow-up to #474, in the sole checkout on
`task/provider-request-lifetime`, from accepted/main-qualified
`390005131e536a32d07d15453a6e7317b43cdb5f`. Validated local product checkpoint:
`add8c7bb92821b54b514519d1f94e42189188942`. Protected publication remains pending. No provider/account/consent,
credential policy, dependency, global HTTP/logger or SDK/module-cache source changed.

This repairs the configured transport boundary for SDK-manufactured requests, not all
SDK implementations, aliases, URLs, headers, logging, retries or job cancellation.
#496 correctly preserved constructor context but did not certify that GetRequest
used it when constructing its own request.

## Actual baseline

A private actual non-test-mode SDK diagnostic on clean42219322 used a valid synthetic
opaque credential, the accepted response-privacy option and in-memory transport:
constructor canceled before GetRequest, yet HTTPcalls1/request-context-canceledfalse/
returned-errorfalse; diagnostic exited1. No socket/discovery/database/real provider
or health record was used. Pinned GetRequest uses http.NewRequest, not constructor
lifetime, at base_client.go265.

Actual SDK tests on39000513 then genuinely failed1.346s in three controls:

- Pre-canceled constructor still returned successful data.
- A live constructor deadline was missing on the generated request, despite a complete
  valid success response.
- An active response read ignored constructor cancellation. Controlled fixture release
  returned EOF instead; the owned request goroutine was released/waited, not stranded.

The live-at-read control separately prevents an apparent fix that cancels as soon as
RoundTrip returns headers, before the body can be consumed.

## Integration contract

WithProviderRequestLifetime is appended after base HTTP-client/privacy options at all
six current production factory sites: regular sync, manual probe/import, related import
and both default-off diagnostics. Those callers keep their existing stable request or
detached job contexts. HTTP abort does not become tracked-job cancellation; no Gin,
application JWT or replacement profile is copied into jobs.

- Nil lifetime is not replaced with Background. A Background/value-only constructor
  with no cancellation/deadline needs no wrapper or invented default timeout.
- Where needed, configured and explicit test clients are copied, preserving timeout,
  jar, redirect and underlying transport policy without enabling test mode.
- A pre-completed constructor or request returns its standard cancellation class before
  calling the base transport. Otherwise a cloned request retains its own values and
  cancellation, adds constructor cancellation and exposes the earlier deadline.
  The link does not copy constructor values or private cancellation causes into
  the request; whatever values/cause its original context already carried remain
  its caller's responsibility, not a new global error-privacy guarantee.
- Links remain active through response consumption. EOF/Close, transport failure or
  no-body response stops callback/deadline registrations; release never cancels callers.
  A cloned response/body wrapper preserves status, headers, length and exact successful
  bytes, including decimal lexemes. It is intentionally not body-object identity.
- Callers must still Close on early stop/non-EOF read error, as the SDK GetRequest does.
  This cannot force a noncompliant transport/body to honor context or bound arbitrary
  Close behavior. No new aggregate retry, whole SDK latency or global privacy guarantee.

## Qualification at this checkpoint

Actual SDK controls pass race1.415s; expanded transport lifecycle/registered-callback
cleanup, EOF/close-failure/error/no-body, both cancellation/deadline directions, original
request/value isolation, client-copy and nil/Background controls pass1.423s. Standard
cancellation/deadline causes remain observable without leaking private constructor
causes. Existing actual OAuth/FHIR rejection-privacy and real token compatibility
assertions remain, paired with the new option; selected race10.632s.

Complete models race24.856s and focused actual-handler context/diagnostic/credential-log
regression race8.843s pass; changed models/handler vet and diff checks pass. Bounded
production factory inventory finds the six adopted sites, not a type-resolved whole
callgraph certificate. All nine tested source hashes are retained privately.

Complete serialized ordinary models/handler/web passed2.099/87.027/64.876s, captured
Go-test exit0 with unchanged15-minute package/fixture budgets. All nine source hashes
match through actual exit capture; documentation edits were separate. Owned wrapper75502
was subsequently observed absent. Full serialized native core races passed
cmd1.596/host22.903/shell27.464/SMART6.036/fixture1.237/smoke1.200s with unchanged
source; known macOS27.2-object/11.0-link-target warnings remain, not older-OS acceptance.
No complete handler/web race, new native package/GUI/runtime, hosted publication, actual
provider, independently operated recovery or production-readiness claim is made.
#474 global inventory, #479 maximum race and blocked #487 runtime/negative gates remain.
