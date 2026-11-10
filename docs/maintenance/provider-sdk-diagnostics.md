# Legacy SDK instance diagnostics and operation errors

## Ownership, acceptance and limits

Pi owned #502 in the sole checkout on `task/provider-sdk-diagnostics`, from protected
#501 merge/main-qualified `a7432c368d9e3dd55488c9d43f86bd3b4f007f7c`. Eight-file product:
`4cbb7af0a513c2580b46ebd939898e06803581a8`; exact docs/head988487fce113a54ffe78be8b1d4f0475a60c218f
passed seven42733463248 (Backend9m38/Frontend3m58/Desktop11m03/actions57s/Go15m28/
JS2m41/aggregate3s). Fresh matching/CLEAN/no unresolved threads preceded protected #503
merge `b33a994571f45897867d7c920029d1f50503b265`,2026-11-10T01:25:44Z. Separate exact-main
Validation42735717066 passed; #502 closed boundedly. This is a separate surface from
accepted #496 rejection-body privacy and #500 request lifetime, not evidence those
receipts promised global SDK privacy or that hosted Desktop erased the local failure.

The boundary is the supplied SDK-instance logger and outward SourceClient method/
construction errors at current application callers. It is **not** an interceptor for
SDK standard-log/stdout/stderr calls, arbitrary transport-error or historical logs,
existing parent logger fields, all summaries/credentials/successful clinical outputs,
or any other SDK instance created without this application boundary.

## Genuine evidence

A private actual non-test-mode SDK diagnostic on clean1ebbf0b4 with both accepted
transport options observed HTTPcalls1/projectionabsenttrue/error-location-leaktrue/
SDK-debug-location-leaktrue, exit1. Exact SDK error/log retained privately, no raw dump
published. All HTTP was intercepted in memory; the path/query canary and opaque token
were synthetic, with no endpoint/provider/discovery/database/profile/health record.

Tracked actual SDK baseline on a7432c36 genuinely failed0.594s at both outward error
and SDK DEBUG location assertions. Correct outgoing path/query/provider Bearer, one
call, no failed projection and rejected-body close remained positive controls.

An initial draft passed that SDK case1.378s but expanded tests found a **draft gap**:
logrus keeps malformed field-error text in unexported Entry state, observable in a
recovered panic even with a constant formatter and cleared exported fields. Expanded
race0.335s genuinely failed that control. Resetting the entire private entry before
forwarding fixes it; expanded race1.386s passes, rather than hiding the negative.

## Owned integration

NewProviderSourceClient replaces the six current direct factory calls (regular sync,
manual probe/import, related import and both default-off diagnostics). It copies the
caller option slice and appends the accepted response-privacy/lifetime options last.
Original stable request or deliberate detached job contexts remain unchanged; no
Gin/JWT copy, job HTTP-abort cancellation, default timeout, dependency or SDK/module-
cache change. Manual SDK construction ignores HTTP options but receives the private
logger and typed operation wrapper; this is not a new manual network-lifetime claim.

- A separate logger owns WithFields/WithError-derived entries. A naïve FieldLogger
  proxy could escape through their returned *Entry. A private hook/constant formatter
  resets raw message, fields, context, caller and internal field-error state, forwarding
  only a fixed diagnostic at the original severity through the existing parent logger.
  Existing parent fields are not newly declared private or erased.
- Pinned root/desktop logrus1.10.2 Entry.Log(Panic/Fatal) forwards severity without
  triggering control flow. SDK Fatal* still delegates the copied parent ExitFunc once;
  Panic* still panics with a reset private entry. Tests retain all seven severities,
  parent level filtering, job metadata and untouched parent writer/formatter/level.
  No shared/global logger, hooks, formatter or output changes.
- The typed SourceClient decorator forwards all eight error-returning operations,
  original arguments/options and successful/partial outputs unchanged. Error() exposes
  fixed operation/cancellation/deadline text and a bounded allowlisted OAuth code only
  when the original error retains that typed classification. Unwrap preserves original
  errors.Is/As inspection; it does not reconstruct chains the SDK already flattened or
  scan arbitrary private error strings. Original causes remain private evidence, not
  safe-to-dump objects.
- Clinical bytes, original locations/IDs, bundles, credentials, versions and upsert
  summaries are not diagnostic redaction targets. They remain functional data requiring
  their existing local ownership/provenance controls, not a new global privacy claim.

## Qualification at this checkpoint

Actual SDK warning/success and production manual-factory controls pass race1.381s:
real SDK non-JSON warning remains visible while original Binary URL/id/content type/
base64 exact bytes (including a decimal lexeme/NUL) remain unchanged. Malformed local
manual input crosses the actual production factory and returns a private error without
losing original credential identity. No socket or real record used.

All eight method-forwarding success/failure/options/partial-output controls, typed
OAuth/cancellation/original-cause inspection, all severities/owned Fatal/Panic/filtering,
actual SDK privacy/lifetime and existing real OAuth compatibility selected race11.254s
pass. Complete models race23.512s, focused handler context/diagnostic/credential-log
regression race8.995s, changed-package vet/diff pass; eight source hashes retained.
Bounded inventory finds six adopted callers and one owned underlying factory, not a
whole type-resolved callgraph certificate.

Complete serialized ordinary models/handler/web passed2.652/97.691/66.268s, captured
actual Go-test0/unchanged15-minute budgets/eight source hashes. Go/configuration source
was frozen through capture; documentation separate. Owned wrapper1870 observed absent.

Initial serialized native invocation genuinely failed: cmd1.809s passed, host20.131s
failed unchanged TestLocalIdentityResolverUsesRealEmbeddedAuthentication first signup
at18.20s, expected200/actual500 with fixed Could not create account. Shell36.389,
SMART6.299, fixture1.247, smoke1.179s passed. No cause established or whole invocation
rewritten green. A bounded control used accepted-a743 production source via private Go
overlay (four handler blobs hash-verified and new helper excluded; host tests unchanged):
host race25.501s passed. Then one unchanged current host-package race23.504s passed,
eight hashes unchanged before/after. Components are thus qualified separately while
retaining the failed invocation. Original15-second/cost14 registration and package
budgets/assertions remain; no scheduler/host/foreground/priority/capture experiment or
blind repeat. Known macOS27.2-object/11.0-link-target warnings retained, not older-OS
acceptance.

Full handler/web race, new native package/runtime/GUI, live provider/consent,
independently operated recovery and production readiness remain unclaimed. #474 global inventory, #479 maximum race and blocked #487 runtime/negative
gates remain; no historical log cleanup or revocation.
