# Medical-record wizard validation feedback (#102)

The wizard previously disabled Save for invalid forms without identifying the
section containing the problem. Each invalid section now has a visible badge,
screen-reader text and aria-invalid. Validation guidance explains the disabled
button and disappears when the form becomes valid. The existing save-disabled
condition and all submission/serialization code remain unchanged. This HTML-only
correction does not accept invalid records or alter database/Wails contracts.

Three synthetic DOM regressions fail on the original template for the missing
encounter/medication indicators and guidance. The exact same tests pass after the
correction. All eight targeted wizard tests pass, including existing invalid-save,
one-request, retry/preserved-form and payload-log checks. Full ordinary browser
validation passes 426 tests with one existing skip and exits 0; ten Node checks,
production and Storybook builds pass. No test, timeout or threshold is relaxed.

Source checkpoint `6aca492ad3b317a387de37460d3c7a2b64b26200` has a fresh native
package with ZIP CRC, strict
extracted signature and all ten WebView/native checks passing within the unchanged
120-second limit. Every generated native frontend file matches the validated
corrected frontend. The preservation wrapper atomically retains the three original
build outputs, then verifies restored content, inodes, modification/creation times,
permissions, ownership, flags and extended attributes. This build has no OS-managed
provenance differences. Earlier archives, recovered coverage and purchased Fasten
data remain untouched. Automated native smoke tests use a new private synthetic
data directory with
direct providers disabled. A separate manual-test instance uses the existing
local synthetic-provider mode, its own fresh data/cache/temp directories and no
operator configuration. Native UI automation is unavailable, so #101/#168 remain
open awaiting a human observation; the test instance has not been observed by the
agent. No live profile is moved or inspected for this pending check.

[Validation and recovery evidence](wizard-validation-feedback.json).
[Native package and preservation receipt](native-6aca492ad-validation.json).
Issue #102 stays open until all three required checks pass and this PR integrates.

## Additional issue triage remains open

Fresh body/comment reviews of #82, #88 and #178 confirm actionable requirements.
#82 concerns provider-owned installation registration, #88 asks for secure defaults
and deliberate-relaxation guidance, and #178 includes dashboard-authoring and unit
questions. None is retired from its title or an earlier partial implementation.
Their current snapshots are retained privately for the next relevant review.

## Later integration and native acceptance

#362 merged at `f1ec2ddc6a72901364b5a1aa9582accdf5db1d8a` after Backend,
Frontend and Desktop passed. #102 was individually closed with its evidence.
The Mac restarted; durable package, coverage and history archives were reverified
before continuation. The native bridge subsequently worked and #101/#168 were
actually exercised. [Later observations and attachment correction](2026-11-02-native-attachment-upload.md).
The earlier unavailable-bridge description above records that earlier checkpoint.
