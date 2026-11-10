# Fail-closed encryption setup consumers

## Ownership and original draft checkpoint

Pi owns #507 in the sole checkout, branch `task/encryption-setup-consumer`, from
protected #508 merge `55e5383db038efbf2977db655ece0c9e823282f4`. That parent passed
separate exact-main Validation42758364242; #506 is closed only for its response-cache
boundary. This ten-file frontend product checkpoint is
`8ec14fd0e265976fc9d9095009c52ce16d97dd81`. It is **not accepted or production-qualified**:
the original checkpoint was storage-blocked, with matching-head/protected
merge/separate exact-main qualification pending.

On 2026-11-10, sufficient storage became available. The unchanged exact draft head
`e9eb06d2c8d609def5ecb3f77a574aa29ba55184` passed canonical Node/browser/production/
Storybook validation, all four step exits zero, 865 browser successes/one skip and
normal cleanup. The original preflight failure remains separate evidence. At the
user's explicit request, #509 is now the protected consolidation vehicle for the
original #431/#488/#509 heads. Combined-source checks remain distinct; repository
integration does not certify native GUI, key custody, independent recovery or overall
production readiness. See the [current consolidation scope](production-readiness.md).

## Two distinct boundaries

The existing global interceptor deliberately preserves legacy local401/403 handling.
Other HTTP failures, including400/500, are thrown. Old encryption methods mapped
`data` without checking literal success; their consumers used `toPromise()`, which
also resolved an empty completed stream as undefined. Neither a false envelope nor
an empty completion was a valid setup acknowledgment. Raw HTTP errors also retained
arbitrary response objects, which the generated-key consumer logged.

Generation now requires literal `success === true` plus a nonempty string `data`,
returned unchanged. Validation and setup require literal success, accepting the
backend's existing successful no-data envelope. All three boundaries explicitly
reject empty completion and provide fixed private errors, without raw causes or
console logging. Each request has its own30-second bound; no POST retry is added.
The configured HTTP injection token, API root/prefix/explicit remote URL, exact key
bytes, existing interceptor, authorization, backend and cryptography are unchanged.

A positive setup acknowledgment is not an unlocked-state receipt: standalone setup
may respond before asynchronous reinitialization. Both consumers independently
observe health before success, countdown or navigation. Observation requires literal
success, `standby_mode === false` and boolean `first_run_wizard`. It uses a fresh
non-secret UUID scope per subscription and a unique query value per request, not
key/identity authority. This avoids adding a non-safelisted CORS request header to
configured remote APIs. Polling is nonoverlapping at500ms, individual health requests
are bounded to5 seconds, and the whole observation to30 seconds. Transient restart
failures or malformed metadata never imply readiness. These are separate operation
bounds, **not a30-second total limit for validation plus setup plus confirmation**.
No live CORS/cache/proxy/TLS/native-network interoperability certificate follows.

## Owned views and explicit uncertainty

- At most one active mutation sequence per view; capture the exact key before waits and use it for
  both validation and setup. Do not regenerate while setup is pending.
- Owned HTTP subscriptions, delays and countdowns end on destruction. Epoch/active
  checks suppress late success, navigation and toast updates; superseded generation
  is discarded. Clear form/generated-key references and private view flags on exit.
- Mask restored key entry. Preserve exact generated file/clipboard bytes and manual
  restore behavior. Download initiation is not proof of a saved file; the existing
  acknowledgment now explicitly asks the user to verify and protect the file.
- Setup/confirmation failures retain uncertainty: a submitted key may already have
  changed the server. Keep the saved key and check the server before retrying.
  Unsubscription or leaving does **not** roll back sent work, revoke credentials,
  erase all memory, remove saved files or undo clipboard copies. A health observation
  does not attest that the submitted key was durably stored or establish custody.

## Actual synthetic controls and counterevidence

- Actual registered AppModule HTTP/service baseline on55e:25 controls,20FAIL/5pass.
  False200/malformed envelopes were accepted;400/500 controls instead failed the
  private-error contract because they exposed raw `HttpErrorResponse`, not because
  those statuses resolved successfully. The issue records this correction explicitly;
  retain original logs rather than rewriting the first result. No auth-bypass claim.
- Actual components/templates at API/router/toast/timer seams:8 controls,7FAIL and
  one preexisting duplicate-restore positive. Pending health was ignored, late keys
  and destroyed form secrets remained, generated setup admitted duplicate requests,
  and arbitrary generation/setup errors reached console. The component tests do not
  impersonate a running Wails backend or a normal native GUI.
- The first typed-service draft still completed empty on401/403:37 controls,
  6FAIL/31pass. Explicit empty-stream rejection fixed this gap. An additional actual
  form control found the inherited restored-key input was visible text:17 controls,
  1FAIL/16pass, before masking it. No original behavioral assertions were relaxed.
- One union-Observable TS2349 test preparation failure was corrected with an explicit
  `Observable<unknown>` local. A separate invalid fake-native fixture produced
  4FAIL/69pass: it mixed real HTTP location with a contradictory Wails classifier
  authority and used an invalid API base missing its leading slash. Its exact source
  was privately copied/compared/hashed before replacing the fixture, not changing
  production URL/interceptor behavior to satisfy it.
- Current installed Angular/Chrome Headless execution: **73pass**:37 receipt/private
  error cases,9 health ownership/freshness/budget controls,21 actual component/template
  lifecycle/file/clipboard cases and6 configured root/prefix/remote-API provider
  controls. No real records/provider credentials or live backend/provider requests;
  clipboard writes and file downloads are spied. Native-authority classifier tests
  remain separate existing tests;
  configured HTTP tests are not Wails runtime acceptance.
- Separate unchanged registered HTTP/auth-interceptor compatibility suite: **41pass**,
  including synthetic native URL authority and both configured clients. This is not
  an actual native XHR/Wails runtime or live provider test.
- Standalone native-bridge/tooling Node controls:29 tests,27pass,0fail, **two filesystem
  skips**, not29 executed successes. Source hash check covers all ten frontend files.

## Incomplete complete-suite gate

The owned `frontend/tools/validate.mjs` invocation exited1 **before any validation
step**: `Frontend validation requires at least 50 GiB free host storage.` Read-only
observation showed42GiB available. Its owned wrapper62553 was observed absent, no
validation lock remained, and all ten frozen frontend hashes still matched. Do not
lower/bypass the guard, delete ignored/history/private evidence, or call this a
product test failure or complete frontend acceptance. Full Angular, production and
Storybook builds through the canonical receipt remain pending. Existing hosted
Frontend runs Node/Angular/production checks, **not Storybook**, so seven hosted
checks alone would not close this missing local complete-suite gate.

Known Angular icon/form-registration, deprecated-builder and Browserslist warnings
remain distinct from failures. No new packaged native runtime, normal encrypted
onboarding/GUI, older-OS, live provider, key-custody, independent restore, deployment
or overall production-readiness claim. #474/#479/#504 and #487/#488 remain separate;
no clinical projection trial was revived by this frontend repair.
