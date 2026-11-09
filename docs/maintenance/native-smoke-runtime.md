# Packaged hidden native runtime receipt

## Ownership and boundary

Pi owns #483 in the canonical single checkout, branch `task/native-smoke-receipt`.
This qualifies an automated **hidden native WebView/service** path, not visual or interactive
clinical GUI acceptance, encrypted GUI operation, provider acceptance, independent recovery,
notarization, older macOS or overall production readiness. Normal app startup is unchanged.

Installed Cua 0.43.2's existing owned daemon stayed in standard mode with automation prompts
suppressed. Its legitimate background launch route works; exact-window capture still requires
owner-approved direct capture consent. No permission/profile bypass, update, screen capture,
foreground fallback, input into another app or real record/provider/identity was used.

## Implementation

Existing `--smoke-test` still requires explicit disposable `--unencrypted` storage and no provider
configuration. Its directory now must be absent or empty and not a final-component symlink before
storage opens: merely lacking `fasten.db` did not establish freshness. Nonempty directory controls
preserve their existing evidence. `--smoke-background` is rejected without smoke mode and hides
main/observer/callback windows, avoiding delayed popup activation. The 45-second native budget,
password hashing and assertions are not relaxed.

Native completion writes `desktop-smoke-report.json` under the selected private directory and
retains existing `DESKTOP_SMOKE` stdout. The report is bounded to 8 KiB and mode 0600. A private
temporary file is written, synced and closed before atomic hard-link publication; an existing file
or final symlink is never replaced. Failed temporary evidence remains private. Post-publication
unlink warnings preserve the completed receipt and are logged separately. This is not a power-loss
durability guarantee. Receipt failure makes the smoke result fail; normal startup/help do not write
it. A native-completion receipt does not independently prove process exit or its status.

The old unconfigured-provider smoke call had one argument although `SMARTService.Connect` now
requires provider ID plus a local JWT. Accepting any exception could pass on argument/identity
failure. It now runs after actual synthetic sign-in, passes that ephemeral token as the second
argument and accepts only the fixed provider-data rejection. The custom local smoke page does not
persist its token, use local/session storage or invoke cookie-setting authentication. This does not
qualify normal Angular renderer/profile storage partitioning.

## Actual package evidence

Clean checkpoint `1dcde76900c02c3e7e50b790d3529549c4119c29` was built normally, extracted and
strictly ad-hoc-signature verified. Its actual packaged backup binary passed the separate-process
synthetic encrypted WAL restore (4.07s test / 4.534s package). Previous exact-a594 ZIP was privately
preserved and byte/hash verified before replacement. Source, full build logs and hashes remain
private; no candidate archive or raw evidence is uploaded.

The exact extracted app was launched through normal Cua with a fresh explicit test directory and
`--smoke-test --smoke-background --unencrypted`. Launch reported focus suppression and no ready
visible window. Independent filesystem readback found its mode-0600 native receipt with all ten
checks passing:

1. Native binding round trip.
2. Direct SMART disabled without registration.
3. Fresh database API.
4. Local account creation.
5. Local sign-in.
6. Authenticated native unconfigured-provider rejection.
7. Native URL validation.
8. API errors separated from SPA routing.
9. Owned popup callback round trip.
10. No callback disclosure to the unrelated observer window.

A **negative-only compile overlay** of that checkpoint replaced only the application's token
argument with a synthetic invalid token. The separately retained/strictly verified overlay package
is explicitly not an exact-source candidate. Through the same standard Cua route and a different
fresh directory it reported `ok:false`, five earlier checks and
`Error: direct SMART rejects unconfigured provider`. Thus an identity rejection is no longer counted
as a successful unconfigured-provider control. No tracked source was changed; the exact positive ZIP
was restored and byte verified after the experiment. Both named driver lifecycles were ended; no
unrelated app/process was stopped.

Accepted diagnostic #482 was then normally integrated at
`779f20988d88971b2ea03b39fff92fd49ed61177`, followed by fresh-directory hardening. Complete
post-parent native races passed: command 1.562s, host 24.271s, composed encrypted shell 28.185s,
SMART 6.089s, fixture 1.223s, smoke 1.250s. Native vet/diff checks passed. Existing newer-macOS
object/older-link-target warnings remain. Both declared backend/desktop Go module integrity checks
also passed; this alone is not a reproducible-build/whole-supply-chain certification.

A second clean, full build at `68e3c9732e6438c8dbf8ba4dbeda276c3207144a` includes accepted #482
and the strengthened directory guard. Extracted strict signature and the actual packaged encrypted
WAL restore passed again. Its pre-GUI background-without-smoke invocation exited 1 and did not create
the selected unused data path. A fresh normal-Cua hidden launch again published all ten checks as
`ok:true` in a mode-0600, 339-byte receipt. The native test page is unchanged from the calibrated
negative-control checkpoint. That archive is retained as an exact-68e inspection candidate, not a
later receipt-only commit, protected-main release or published/deployed artifact. A newer distinct
b916 candidate and its limits are recorded in the [application token receipt](application-token-boundary.md).

The native slice qualified exact head `2a03fea8a4932f5d57ac42e06a5a8c6246eea732` with seven
checks (42624220876) and no unresolved threads, then protected matching-head #484 merge
`0252c677ffd3c56900eacf60d7c214401a7effe3`. Separate exact-main 42627751541 attempt 1
passed. This is source acceptance, not publication/deployment of a candidate archive.

Subsequent pending #487 expands the custom smoke page to twelve checks using real authenticated
multipart requests and the same bounded encoder as the Angular interceptor. Raw FormData actually
failed in the native host despite Go/browser green. Exact-c9 subsequently passed all twelve checks;
later exact-eed timed out after three checks before account creation completed. Bounded fixed-label
progress now makes such timeout stages observable, never success. Its guard-bypass negative is not
yet calibrated. See the [manual receiver receipt](manual-import-receiver.md) for exact source,
failed attempts, archive status and outstanding native-race/runtime gates. Historical ten-check
acceptance above does not qualify that later file path or candidate.

Actual visual receiving-profile/authority/clinical import/repeat-change/export/profile-switch/logout
and restart workflows remain unqualified. The earlier exact-a594 clinical inspection app was left
running in its own synthetic unencrypted directory; this run did not interact with it or infer
permission from old TCC booleans. Independent key custody/recovery and deployment gates remain.
