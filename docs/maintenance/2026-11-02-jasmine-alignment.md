# Coherent Jasmine tooling for the existing engine

A real browser diagnostic confirms that the current Karma suite executes Jasmine
4.6.0. karma-jasmine 5.1.0 explicitly depends on jasmine-core ^4.1.0 and resolved
its own nested 4.6.0 copy. The separate top-level core 5.1.2 supplied HTML assets,
while declarations described 3.5.14. A top-level version bump alone would not
upgrade the ordinary browser engine.

The corrected baseline explicitly declares the same core 4.6.0, allowing the
adapter and reporter to share one installation. All files of the active engine
are byte-identical before/after. Declarations advance to 4.6.6, and the HTML
reporter advances to 2.3.0 with a satisfied compatible peer. The final frozen
install adds no new Jasmine peer warning. The first candidate's peer warning,
corrected by explicitly declaring the shared core, remains in the raw evidence.
No dependency override or test-engine downgrade is introduced. All 71 unrelated
direct installed versions and production dependency selections remain unchanged.
Renovate holds core and declarations on this adapter's major; a future engine
upgrade must coordinate adapter, loader, reporters and declarations.

Two new Node guards independently reproduce the old installation/type mismatch
(two assertion failures, exit 1) and pass afterward with the five bridge tests
(seven passes, exit 0). They compare resolved engine installations and actual
engine/declaration majors, without hardcoding a chosen version. CI and the
contributor command include these guards. The only workflow change is the Node
test glob; permissions, triggers, action pins, commands otherwise, check names
and strict protections are unchanged. Actionlint passes. The cached index edit
cleared the workflow's skip-worktree bit; it was restored immediately, with the
workflow still absent and sparse patterns unchanged.

An ordinary full browser run passes all 423 existing tests with one existing skip
and exit 0. A separate real HTML-reporter diagnostic passes and reports the same
4.6.0 engine. Diagnostic fixtures were exclusively created and moved to retained
evidence afterward; the first probe's missing old declaration metadata and its
compiler failure are retained, not presented as a product regression. The full
suite's assertions, exclusions, timeouts, coverage configuration and browser flags
are unchanged. Production and Storybook builds pass, with original documentation
metadata restored. Every one of 22,287 produced application files matches the
latest native package's frontend byte-for-byte; only its prior Finder metadata
is absent. No new native run is claimed for this development-only change.

[Validation and complete evidence hashes](jasmine-alignment-validation.json).
Post-capture checks found newly OS-created destination .DS_Store files
changed after the first hash manifest. All non-Finder evidence still matches.
The original manifest and cache paths remain intact; current cache bytes are
snapshotted with their earlier hashes recorded. The earlier cache bytes were not
snapshotted and are not claimed recovered. These were absent in the original task
source and are unrelated to the retained pre-task repository/Finder snapshots.
The stable evidence manifest excludes only these mutable cache paths and includes
the observation record and current-byte snapshots. No validation output or
original user metadata was rewritten to force a checksum match.

[Original installed packages, both old core copies and metadata archive](jasmine-installed-preservation.json).
[All five exact historical tips, verified bundle and unique-object proof](jasmine-heads-recovery.json).
Required fresh Backend/Frontend/Desktop checks still gate integration.

## Individual proposal dispositions after integration

- #237: replace the historical 3.10 declaration update with 4.6.6 matching the
  actual engine; the original 3.10 version is not installed.
- #247: adopt the reviewed HTML reporter 2.3.0 target.
- #245: retire the separate core 5.13 proposal; it does not upgrade this adapter's
  engine, and the corrected shared 4.6 baseline replaces the redundant 5.x copy.
- #304: retire the standalone core 7 proposal for this baseline; no engine-7
  migration is implemented or claimed.
- #297: retire declarations 7 for the active engine 4; no type-7 adoption is claimed.

These dependency targets are not application features. The historical proposals
stay open until the corrected baseline integrates through protected main. A future
coordinated engine migration remains possible using the preserved patches; this
change does not force unsupported majors through the current adapter.
