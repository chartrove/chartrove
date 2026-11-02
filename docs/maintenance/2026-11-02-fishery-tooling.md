# Fishery factory tooling and pinned Storybook execution (#242)

The isolated lock patch selects Fishery 2.4.0 within the existing manifest range.
Only its version, archive URL and integrity change. All other lock bytes, 72
other direct versions, the manifest and all four factory source files are
unchanged. Frozen installation passes; the archive matches registry SHA-512 and
all 13 published files match installation. Runtime dependency lodash.mergewith
remains unchanged. Upstream source and release history were reviewed directly.

The release adds parameter typing and typed-array merge handling. A compatibility
probe preserves nested overrides, arrays, sequences and hooks; the published
Buffer fix changes the previous merged array [7, 8, 3] to Buffer [7, 8]. The
probe initially used a wrong API name; both versions declare rewindSequence,
and the corrected probe passes. Both probe sources/results are retained.
Existing clinical fixture defects involving zero/false fallbacks and the 4.9
integer default are separate work and are not claimed fixed by this update.

All 429 browser tests pass with one existing skip; ten Node checks pass.
Production compilation passes, and every one of its 22,287 output files matches
the freshly validated native frontend from source 9add19dd1 byte-for-byte.
Fishery imports are confined to test/Storybook factories and absent from that
production source map. Go/Wails sources and modules are unchanged, so the
existing native package and actual practitioner UI acceptance remain applicable.

## Storybook on Node 26

The ordinary bare-ng Storybook command failed before compilation because
Compodoc spawns yarn, which is absent from a clean Node 26 PATH. The existing
pinned Corepack-through-npm environment supplies Yarn 1.22.22. The build/serve
Storybook Makefile recipes now run the local CLI under that same environment;
the full static build passes. Contributor instructions identify these targets.
No dependency, engine, browser, timeout or check was relaxed for that correction.

The first preservation precheck expected an existing documentation.json, but
this checkout intentionally had none. It stopped before moving any artifact.
The configured Compodoc generator creates that file during Storybook. Validation
retains the generated JSON with its checksum and restores the original absence;
both failed attempts and the successful pinned command remain retained.

[Validation and preservation receipt](fishery-validation.json).

## Historical proposal preservation

The current #242 head 25300eb7b2d606eca27c878089f82980c77b1ff8 differs from the
original intake tip. A new checksum-recorded bundle preserves its exact one-commit
patch, with all four unique objects proved present in the original pack bytes,
strict pack validation and main-reachable prerequisites. The original intake
bundle and installed 2.2.2/Yarn metadata archive remain intact.
[Recovery receipt](fishery-head-recovery.json). #242 and its branch stay open
until protected integration and fresh dependency checks complete.
