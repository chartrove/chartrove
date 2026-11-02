# Minimal Karma 6.4.4 replacement

#228's stated Karma 6.4.4 target is adopted through a minimal lock-only replacement.
The complete current proposal also refreshes unrelated transitive artifacts. Its
exact head/diff is preserved in a checksum-recorded bundle with native prerequisite
and isolated strict pack/object verification. [Recovery](karma-head-recovery.json).
That wider lock refresh is explicitly superseded in favor of the bounded target;
this change does not claim to implement all its incidental transitive updates or
resolve all frontend dependency advisories.

The current ~6.4.3 manifest range accepts 6.4.4. Change only Karma's version,
resolved tarball and integrity lines. Direct version comparison shows only Karma
changes; 74 other direct versions and the complete runtime manifest stay unchanged.
The frozen install passes without engine bypasses or an unrelated lock refresh.

Both official npm distributions are downloaded and their SHA-512 integrities
verified against version metadata. All file paths and 87 regular files are
identical; the only three changes are package.json (version/author ordering),
CHANGELOG.md and the bundled publisher local.log. Dependency ranges and executable
code are identical. [Exact distribution comparison](karma-distribution-comparison.json).
The historical installed tree remains in the earlier verified dependency archive.
The full browser suite passes with the ordinary Karma configuration and dots
reporter: 415 tests, one existing skip, exit 0. No diagnostic wrapper or forced
termination is needed in this run. Launcher teardown warnings remain, and no
claim is made that this metadata-only patch fixes the intermittent shutdown hang.
[Validation receipt](karma-validation.json).

No application source, Go graph, FHIR storage/migration, Wails contract, production
asset or test assertion/timeout/threshold changes. The #347 native package and
#348 byte-identical application-asset proof remain the runtime evidence. Fresh
required Backend/Frontend/Desktop checks still gate integration. Close the original
proposal only after this target is safely integrated; retain recovery history and
verify open-PR dependencies and the exact remote tip before retiring its branch.
