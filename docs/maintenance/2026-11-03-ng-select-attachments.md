# Attachment picker maintenance

Update direct ng-select 23.11.0 to 24.2.1 for #295. Its Angular 22/CDK and Node
26 peers match the existing toolchain; the other 72 direct installed versions
and unrelated manifest/lock bytes remain unchanged. All 15 installed package
files match the SHA-1/SHA-512-verified npm archive. The official tag matches its
registry Git commit; all 21 embedded runtime-map sources and four published SCSS
sources match that commit. Three CSS maps lack embedded sources and are recorded
separately. MIT attribution is retained. The full upstream suite was not run.

Four actual wizard consumer cases pass on old/new versions with identical test
bytes. Medication and procedure pickers retain readonly behavior, multiselection,
hidden selected options, removal, serialized attachment IDs and original attachment
data. Initial tab-label and inline-host teardown diagnostics remain excluded;
the final cleanup still checks for leaked external dropdowns.

The first production build exceeded the unchanged 10 MB initial-bundle limit by
20.95 kB. Lazy-loading the existing wizard alone still exceeded it by 10.54 kB.
Remove unused eager ng-select and standalone wizard/dialog registry entries,
then import the wizard only at its two existing entry points. No template uses
the removed shared exports. Each dialog remains directly imported by its actual
consumer. Capture the clicked Encounter across loading and reject late opens
after either owning view is destroyed. Existing save/dismiss navigation remains.

Three entry-point controls pass before the change; six cases pass afterward,
including real modal creation, navigation, Encounter identity and cancellation.
The production initial total is now **9,973,017 bytes**, below the original
10,000,000-byte limit. Webpack identifies the wizard in a non-initial chunk.
Diagnostic stats are retained outside runtime assets. All 523 browser tests
pass with one existing skip; Storybook passes. The qualified frozen install and
ten Node checks are reused only after unchanged dependency/check-input proof.
Coverage and generated-documentation absence are preserved.
[Validation](ngselect-validation.json).

No Go, database, Wails, record serialization or clinical-data migration changes.
The exact historical head is preserved in a verified checksum-recorded bundle,
with all five unique objects proved present in archive bytes and prerequisites
reachable from main. [Recovery](ngselect-head-recovery.json).
Native package/UI acceptance and protected integration remain pending; #295
and its historical branch stay live until those checks and fresh retirement gates.

The maintainer's separately authorized legacy-narrative retirement is documented
with its own source/ancestry/bundle evidence; no old repair or migration runs.
Specific retained upstream expansion areas remain pending clarification.

Raw evidence: dist/validation/2026-11-01-consolidation/ngselect-review/.
