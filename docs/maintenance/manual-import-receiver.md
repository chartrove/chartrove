# Manual-file receiving identity and native multipart transport

## Ownership and scope

Pi owns #487, branch `task/manual-import-receiver`, in the canonical single checkout.
This is a bounded legacy-file receiving-context correction and the native transport defect exposed
by its actual runtime gate. It is not whole-chart fidelity, household consent, live conversion,
visual GUI, independent recovery, deployment or production-readiness acceptance. Only synthetic
records/accounts and private disposable test directories were used. No dependency, Wails version,
module cache, other repository, real provider or production identity was changed.

## Receiving context

The prior component awaited external conversion confirmation/conversion and then submitted a file
using whichever local session was active later. A reused display name did not pin the same immutable
profile. The browser now captures a nonnil UUID at selection, rechecks it before new external sharing
and local submission, and checks delayed results/errors before displaying anything. Missing or
malformed browser identity refuses selection; browser claims remain stale-context aids, not server
authority. Destruction clears private file/error state and dismisses only its owned confirmation.

`createManualSource(file, expectedOwnerID?)` appends an optional `expected_owner_id` multipart value.
The server compares it with the already verified, copied immutable actor before SDK construction,
persistent source writes or tracked-job creation. Temporary upload copying/parsing can already have
occurred; existing deferred private-file cleanup remains. The assertion is not authentication,
provider permission or a delegation grant.

- Absent field preserves legacy clients and their existing verified-actor handling.
- Present empty/nil/malformed/nonstandard-wrapper/duplicate values, case aliases, query assertions and
  file-shaped assertions fail with fixed HTTP 400 errors.
- A different immutable receiver fails with fixed HTTP 409, without source/job mutation.
- Valid same-owner submissions and intentionally detached tracked-job lifetime remain unchanged.
  Same-name deletion/recreation does not revive the old UUID or JWT.

Already-sent conversion or import is not undone by profile change, navigation, hidden-state clearing
or HTTP abort. Stale conversion results do not start a new local import. Stale import completion
reports uncertainty about the original profile, not replacement-profile success or safe blind retry.
Destruction unsubscribes the local request so unfinished client encoding cannot subsequently reach
transport; an already received server job is not thereby cancelled or rolled back. The existing
explicit external C-CDA warning remains; this work does not make conversion offline.

## Actual native transport correction

Go-only HTTP and browser-mock success missed a real native failure. Clean source
`386cd7de121a79f45f1bc1648112c8bc241b8882` packaged and restored successfully, but its real hidden
WebView failed the new receiver gate. A separately labeled status-only compile overlay established
HTTP 500 with the fixed missing-multipart-file class. Signup, local JWT and authenticated GETs worked.
Pinned Wails has different body-byte/body-stream paths; this observation does not prove a particular
upstream implementation defect.

A shared `frontend/src/assets/desktop/multipart.mjs` implementation is now used by the actual Angular
API interceptor and the custom native test page. Only application-API `wails:` FormData is converted.
Ordinary HTTP multipart and unrelated/external requests retain their original bodies and authority.
The standard `Response(FormData)` encoder supplies MIME bytes and its own Content-Type/boundary;
there is no handwritten MIME encoder or provider credential substitution.

Native preparation limits are explicit: **32 MiB encoded envelope, 16 parts, 128 UTF-16 code-unit field
names, 4096 UTF-16 code-unit filenames and 30 seconds**. Preflight bounds raw values, stream consumption
bounds the encoded envelope, and stream-reader cancellation is requested on failure/timeout. Errors contain fixed text,
not filenames, records or tokens. The already selected JWT is pinned before asynchronous encoding;
it is not reselected from a replacement session. Unsubscription prevents a later transport emission,
not an already accepted backend job. Other clinical limits and the native 45-second smoke budget
are unchanged. Larger native uploads are not qualified; ordinary HTTP behavior is not restricted by
this adapter.

## Evidence and accounting

- Actual component plus real API service/HTTP-mock baseline: **6 failures / 17**, including unexpected
  conversion/import after profile/destruction changes, missing expected UUID, missing immutable
  identity and a prior-profile error canary. Corrected 17, then expanded 19 and destruction controls
  pass. Fixed receiving-profile uncertainty and private DOM clearing are tested.
- Real-JWT encrypted embedded baseline genuinely accepted eight wrong/malformed receiver assertions
  with HTTP 200/source/job mutations and a same-name replacement with the stale expected UUID.
  The first fixture incorrectly assumed signup created zero sources; its log is retained and the
  corrected fixture compares before/after counts, including the reserved local source. Corrected
  baseline race failed in 38.679s. Guarded unchanged controls pass in 36.029s, including valid,
  absent-field legacy, deletion/recreation, stale JWT and no-external-request cases. A preceding
  masked signup-500 attempt is retained, not hidden by the retry.
- Parser/legacy source compatibility races pass in 39.208s. Complete backend races at
  `03625749e7` pass: web 682.314s, handler 626.247s, middleware 1.727s. Backend source is unchanged
  across the normal accepted-#486 integration and subsequent desktop-only changes.
- Full integrated frontend validation records **788 browser passes/one skip**, Node, production and
  Storybook zero exits, and scratch removal. Both actual registered HTTP clients exercise exact
  file bytes/receiver, generated content type, original-token pinning, cancellation, unchanged web
  multipart/external authority, and existing auth/error controls. Five standalone Node transport
  controls include binary/decimal/Unicode fidelity, duplicate preservation and raw/encoded limits.
- Complete native races passed before the transport follow-up (command 1.538s, host 24.079s,
  composed shell 28.384s, SMART 6.042s, fixture 1.263s, smoke 1.174s). Three later complete local
  native race attempts fail existing real-signup fixtures with masked HTTP 500. Those host/backend
  sources are unchanged; SMART timing also slowed from roughly 6s to 51.764s. Concurrent unrelated
  host load was observed, but cause is not conclusively established. These attempts are **failures**,
  not complete current native-race acceptance. Hashes, assertions, deadlines, priorities, OS settings
  and unrelated processes were not changed. Current command/smoke race targets pass separately.

## Source-bound runtime packages

Clean `c9b5dca385e1c38b41fa9c12be410c163027e42f` was fully built, extracted and strictly
ad-hoc-signature verified. Its actual packaged backup binary passes a separate-process encrypted WAL
restore (8.56s test / 9.217s package). A normal installed standard-Cua background launch with fresh
explicitly unencrypted smoke storage independently publishes a **mode-0600, 416-byte, all-12 true**
receipt. The two new checks perform actual authenticated native multipart requests: wrong receiver
409 with unchanged source/job counts, then correct-owner import with exact synthetic Patient ID,
owner UUID and expected counts. The earlier ten checks still run. This is actual native fetch/API
transport using the production encoder, **not** normal rendered Angular/XHR/visual clinical acceptance.

Earlier materialized-byte compiler experiments remain distinctly labeled: one timed out, an
allowlisted private phase-sidecar variant passed all 12, and a buffered-only repeat eventually timed
out. None is promoted to an exact-source candidate. A deliberate c9 server-guard-bypass negative
also timed out without stage data; it has **not yet calibrated** the receiver assertion. Its archive
is separate and the c9 positive ZIP was restored/byte verified afterward.

Follow-up `eed8536adf8685757289dad30796e07ae6a00e42` adds only bounded fixed-label progress to the
existing smoke control. At most twelve best-effort posts contain an allowlisted ordered prefix,
never records/tokens. Strict input bounds, out-of-order concurrency and snapshot-copy races pass.
The unchanged 45-second timeout includes acknowledged progress and remains failure even if all labels
arrived without final completion. Normal startup has no smoke routes. Its ordinary package reuses
byte-identical tracked c9 frontend inputs, verifies strict signature and packaged encrypted restore.
Its first actual native run timed out after the first **three** checks, before account creation completed;
a same-package retry later timed out after **six** checks. The latter receiver lookup/count/preparation/
submission substage is not established; neither is an all-12 eed success. Subsequent test-only progress
posts are awaited sequentially with a 500 ms per-post abort bound inside the same 45-second deadline,
not sent concurrently. This observability change is not a claimed production transport/root-cause fix. The later clean `7f44f3839131ce03aaf110cc95c13d1c2e9f4805` ordinary package also verifies strict
signature and packaged encrypted restore, with unchanged c9 frontend inputs. Its actual serial-progress
run times out after three checks. A separate read-only owned-PID run observed almost frozen CPU time
while runnable and no receipt by the outer observation bound; scheduling causation is not established.
A private macOS finite own-process activity-hint overlay also times out after three checks and is **not**
adopted. It allowed idle sleep, excluded latency-critical activity, released/deferred within 45 seconds,
and changed no global settings, nice/taskpolicy, foreground or permissions. The exact 7f ZIP was restored
and byte verified; experiment archives remain distinct. The latest ZIP is this exact **7f unaccepted
inspection candidate**, not current-main, published/deployed, notarized or accepted runtime.

Raw logs, profiles, keys, overlay sources, archives, hashes and receipt files remain private. Named
Cua lifecycles ended; receipts do not independently prove process exit. Direct capture still requires
legitimate owner consent; no foreground/input/permission bypass was used. Fresh complete native
qualification, calibrated negative, protected source acceptance and the broader production gates
remain open. A green source suite or historical package is not substituted for those results.
