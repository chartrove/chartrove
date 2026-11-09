# Private, profile-scoped companion credential handoff

Pi owns [#492](https://github.com/chartrove/chartrove/issues/492) on
`task/companion-token-privacy` in the canonical single checkout. Product checkpoint
`1b65d92e0b44a6d5d487b0a529e1160430f3263d` was normally integrated with accepted API #491
at `2e414141ffb014f1fb54d868865fe031854696f4`; all three frontend source hashes stayed unchanged.

## Actual defects and boundary

The Settings token consumer logged the entire issuance response, including its signed access JWT.
It retained prior token/QR state after some failures or dialog closure and allowed late HTTP/QR
callbacks to populate state after profile replacement or component destruction. Malformed successful
responses and legacy intercepted HTTP-error-as-next-message behavior could also be mistaken for
successful handoff/removal. This does not establish historical log cleanup or a production breach.

The component now uses typed, narrowed responses, fixed payload-free errors and the configured
application API/client rather than a hardcoded root URL/manual local-storage header. It preserves
explicit successful QR/manual sharing and existing device-name/expiration semantics. Its active
server profile must be confirmed before pairing; browser immutable UUID checks are stale-context
aids, **not authentication authority**. The signed server actor and UUID-bound token record remain
responsible for authorization.

A pairing generation and owned modal reference prevent old closures/results from clearing or
repopulating a newer handoff. Secret fields clear on a new attempt, close, observed identity change
(change detection/storage event), failure and destruction. Pending client subscriptions stop;
late QR encoder promises recheck context. Only the owned dialog is dismissed. No global auth-state,
provider, token-expiration or native transport policy changes. HTTP operations have 30-second client
observation bounds; these do not change backend registration/clinical/native-smoke budgets.

Fixed uncertainty messages instruct the owner to review saved devices before retry. Closing,
unsubscription or state clearing is **not server rollback, revocation, memory zeroization or undo of
an already copied/shared credential**. Malformed device lists are not silently treated as zero, and
revocation requires a confirmed success response. The raw QR section is labeled as a sensitive
connection credential, not development output.

## Test-first and qualification evidence

- Eight genuine baseline failures use the actual component/template/AppModule HTTP providers,
  with HTTP/modal/QR encoder boundaries controlled. They cover credential console disclosure,
  prior-secret persistence, malformed success, close/late-QR, replaced profile and destruction.
- Unchanged original assertions pass, then twenty controls cover competing clicks, old-modal
  completion, identity change before change detection, cross-window storage notification, absent
  immutable identity, QR rejection privacy, intercepted-500 removal, malformed metadata,
  discovery failure, default name/zero-expiration/auth header and unconfirmed server profile.
- A later 17-fail/3-pass expansion was a fixture synchronization defect: after mocked profile HTTP
  completed, the newly guarded DOM button needed `detectChanges`. That correction retained the
  assertions; twenty pass. The earlier genuine baseline failures remain separate.
- A missing top-level Corepack/Yarn shim caused command exit 127 with no tests. The installed
  Angular CLI was invoked with the existing Node; manifests/lockfiles/dependency versions were
  not changed. The normal full runner continues to use its existing pinned Corepack/Yarn route.
- First full runner 25158 recorded Node/browser/production exit 0 (792 browser successes and the
  existing skip). Its outer tool ended after 1800 seconds during Storybook. The owned orphan
  finished naturally, logging success after 1100.47 seconds, but its final exit was not captured.
  The incomplete receipt was **not rewritten**. Original logs/scratch/lock metadata were preserved;
  all four former PIDs were confirmed absent without signals before relocating only that stale lock.
- Fresh complete runner 16549 on integrated `2e414141ffb014f1fb54d868865fe031854696f4`
  records all four phases exit 0 and wrapper exit 0:
  `dist/validation/frontend-2026-11-09T20-08-08.450Z-16549/receipt.json`.
  It includes 792 browser successes/existing skip, Node tests, desktop production build and
  Storybook. Its disposable scratch was removed normally. Existing builder/Browserslist,
  template/icon/form and Storybook polyfill/asset-size diagnostics are not suppressed or repaired.

Raw logs, token fixtures and browser data remain private. This is source/browser validation, not
normal rendered Wails/interactive mobile QR scanning, renderer partitioning, production identity,
older-OS, independently operated recovery or overall production readiness. The separate #487
native multipart runtime/negative-calibration gates remain open.

Exact `5391dca276443d1384278d72c667f7cc83ba4eea` passed all seven42693108010.
Fresh head/CLEAN/no-unresolved-thread/ready verification preceded protected #493 merge
`921f42adc2cba7d80e016cd0ecb901dd26c0f014` at 2026-11-09T21:00:43Z. Separate exact-main
42698388769 is pending at this checkpoint; protected acceptance is not a claimed main-run result.
