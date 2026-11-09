# Explicit clinical Sources UI: local checkpoint

Pi owns `task/clinical-import-ui`, initially stacked on `task/scoped-smart-bridge` and API #473.
This is not merged or native clinical acceptance.

The Sources card preserves the original native Patient string inside a bounded collection rather than
serializing a parsed numeric Patient. It offers a separate Patient/Observation retrieval action; the
complete raw collection also remains a string. Clinical parsing in the renderer is for display/counts,
not authority or reconstruction. The backend/native parser remains authoritative.

The card shows the backend-verified receiving UUID and username, requires an unchecked-by-default
recording-authority confirmation, and passes expected UUID plus `confirmed: true` to the atomic API.
Each newly retrieved clinical preview resets confirmation. A missing Observation collection is distinct
from a complete zero-Observation response; missing records are not deletions. Provider-origin metadata
remains owner-declared, not provider attestation.

Read, import and cleanup controls exclude overlapping expensive work. Identity is rechecked before read
and write and after asynchronous responses. Stale previews are cleared rather than applied to a different
profile. Import uncertainty retains exactly the same raw preview for safe digest replay; historical
acknowledgment does not claim old data is current. Failed clinical refresh retains only the prior complete
preview, labeled as such, never a partially received replacement. Cancellation and teardown release native
credentials; native cleanup failures ask for application closure rather than claiming release succeeded.

Focused Angular run passed 34/34: real component templates and DOM confirmation, private-canary masking,
Patient/Observation decimal lexical preservation, duplicate/overlapping-action guards, cancellation/late
results, pre/post identity-switch isolation, raw API request contract, retry/historical acknowledgment and
cleanup progress/failure. Tests use controlled service/native promises and HTTP testing transport, not real
provider identities or a running desktop. New tests and implementation were developed together.

The API notification-backpressure correction was integrated while its hosted qualification was still
pending at that earlier checkpoint; API #473 has since merged through protected matching-head checks. The native synthetic fixture now grants/searches Observations and preserves exact
Patient decimals. For each of its independent origins, the first two clinical reads are byte-identical;
third onward changes one high-precision decimal, deliberately exercising new/replay/update. This scenario
is synthetic-only and does not contact real providers or alter normal startup profiles. Its tests validate
both origins with the real manager/parser and refuse reads after disconnect.

All native internal race packages passed (host 25.412s, shell 1.217s, SMART 6.904s, fixture 1.324s).
Frontend structural/compiler validation passed; combining that long validation with the first full-browser
run exceeded a 180-second tool budget just after Chrome connected, so that invocation is not a passing
browser suite. An unchanged standalone rerun passed 733 tests with one skip; existing test-template
`fa-icon` console diagnostics and browser-support/deprecated-builder warnings remain, not suppressed.
Native macOS object-target linker warnings remain. No older-platform claim is made.

A composed native-service flow now passes under race (27.173s focused; full internal run host 25.314s,
shell 29.297s, SMART 7.455s, fixture 1.367s): real encrypted embedded app, real signup/JWT identity resolver,
synthetic TLS/PKCE manager and two pinned origins, native foreign-owner refusal before provider read,
HTTP stale-receiver refusal, initial three-resource import, exact unchanged replay, one high-precision
update, historical acknowledgement, same Patient at distinct origins, disconnect, encrypted close/reopen,
and exact authorized HTTP original downloads with foreign refusal. Only the trusted-window context is
synthetic; this is explicitly not a Wails GUI/renderer or independent-operator certificate.

The ad-hoc-signed native package built successfully at source `1958ae0598280a8893273e9b4daa728226270477`;
the earlier local package was copied and byte/hash-verified privately before replacement. This inspection
artifact precedes the latest API write-deadline correction and composed test, so it is not their qualified
final artifact. It was extracted and its signature verified but **not launched**: installed Cua Driver
0.43.2 reported existing Accessibility/Screen Recording grants through its read-only helper, while its
normal standard-mode daemon refused both background app launch and a prompt-free check with
`permissions_pending`. No action started, synthetic data directory was not created, and no permission
bypass, installation/update, foreground input or real-app capture occurred. This blocks that GUI route,
not safe source/test work; the composed encrypted flow above was pursued independently.

The extracted inspection package's independent `fasten-backup` executable also passed
`TestPackagedBackupBinaryRestoresSyntheticWAL` (6.384s package / 5.97s test), under the desktop dependency
graph with the packaged-binary integration tag. This is an automated separate-process WAL restore control,
not independent human custody/key-loss operation or a new final-head artifact qualification.

Next: matching packaged GUI, actual repeated/changed clinical UI import and backup-copy recovery.
The import result now links into the source-specific permission/record viewer, with navigation accepted
only when the requested source is already present in the authorized source list. Its read-only simple
Observation preview takes exact numeric lexemes from `resource_raw_text`, not rounded JS numbers. The
pinned jsonc-parser dependency is reused without installation/version changes. Units, unit codes,
comparators and partial date/time strings are displayed as recorded, with no normalization, clinical
interpretation or added precision. Duplicate/ambiguous fields, components, absent-value reasons,
unsupported choices and bounded-complexity failures leave the current exact-JSON fallback available.
This does not claim that every legacy dashboard/lab display is precision-preserving, or that indexed
current-resource JSON is byte-identical to a separately retained whole-snapshot original.

Focused helper/component tests passed 36/36; full browser suite passed 739 with one skip. The source-specific
DOM test deliberately gives a rounded parsed value while retaining a long exact raw decimal and verifies
the latter and a partial date are shown; malicious-looking unit text remains escaped. False/zero, trailing
zeros, exponent spellings, wide quantity numbers, duplicate fields, complex/invalid values, oversized/deep
inputs, allow-listed navigation and clearing private preview state are covered. Frontend structural/compiler
validation passed. Existing test-template/form/browserslist warnings remain recorded. These newly developed
presentation controls are not a pre-existing test-first regression or native GUI qualification.

After integrating the current scoped native bridge normally, a **new matched local candidate** built at
clean source `de437c482c0b5f435b5fb6a81fc32624bdb556b3`, including the original-write deadline and safe native
account-context adapter. The previous `1958ae0598` ZIP was preserved and byte/hash-verified before
replacement. The current ZIP hash/source/build log remain private. Extracted ad-hoc signature verification
passed; its actual packaged backup executable passed the separate-process WAL integration test again
(4.225s package / 3.83s test). All native internal race packages, including composed encrypted clinical
restart, passed at that source (host 25.709s, shell 31.443s, SMART 7.125s, fixture 1.281s).
The matched app is not launched or GUI-qualified; the existing permission-gated route is not bypassed.
Broad legacy pooled-context failures remain explicitly tracked in #474 and cannot be treated as a pass
merely because this local candidate builds or the clinical scope tests succeed. Exact original export and precision-preserving Observation display still
need user-facing operational verification. No live provider, production identity or deployment is authorized
by this checkpoint.

After normal integration of accepted #475 and the bounded #474 repair (#476), the local UI source is
`262e685080b775279f8aff4f7dcc9ec21234ae28`. Complete frontend validation passed there (Node, 739 browser
passes plus one skip, desktop production compilation and Storybook); evidence is retained in the normal
ignored validation directory. An earlier combined 300-second attempt was interrupted during Storybook,
not counted as a complete pass. Its now-exited owner's lock/generated documentation were preserved
privately before a clean full rerun; no other process or repository was stopped or changed.

One concurrent heavy frontend/native run failed two synthetic signup assertions with HTTP 500 after
roughly 35 seconds. Registration has a 15-second budget including cost-14 bcrypt; oversubscription is
consistent with that result, but the masked HTTP response does not itself prove the internal error.
The unchanged complete native suites passed when serialized without competing frontend builds (host
23.613s, shell 28.412s, SMART 6.405s, fixture 1.286s). No product deadline, password cost, identity control
or test assertion was relaxed. Existing platform/linker and frontend warnings remain visible. This newer
source does not retroactively requalify either earlier package, nor grant GUI or independent acceptance.
