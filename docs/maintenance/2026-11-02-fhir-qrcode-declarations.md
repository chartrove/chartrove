# FHIR and QR declaration alignment

Adopt @types/fhir 0.0.44 and @types/qrcode 1.5.6 for #225/#227. Their registry
artifacts pass published SHA-512 integrities, contain no executable JavaScript or
install hooks, and retain their dependency ranges. Preserve the existing Node
18/26 declaration selections; the historical QR proposal's incidental selector
reshuffle is superseded. Change only the FHIR manifest range and two lock blocks.
All 72 other direct versions stay identical. [Artifact review](type-package-artifact-review.json).

The new FHIR declarations expose five compile errors in existing annotations.
Remove unused RatioRange/Range imports from the R4 reference-range model. Type
LForms' polymorphic export as Bundle<FhirResource>, allowing the existing
Observation discriminator to narrow code/valueString access correctly. Do not
cast everything to any or change clinical conversion logic. The initial build
fails with the original annotations; after these corrections, production,
422 browser tests/one existing skip, five bridge tests and normal pinned Yarn
Storybook/Compodoc all exit 0. No compiler check, assertion, timeout or threshold
is relaxed. [Validation](type-package-validation.json).

All executable JS, CSS and other application assets match the fd5194396 native
ten-check package. One main source map differs only in sourcesContent for the
two edited TypeScript files; its source list, mappings and names are unchanged.
Both generated frontend trees/maps remain preserved. This is precise runtime
invariance, not a claim that every output byte is identical. The active FHIR
version, backend models/Go graph, data/migration and Wails contracts remain
unchanged; newly supplied R4B/R5 declaration files do not enable those versions.

The complete old installed declaration folders were archived with native metadata
options before replacement, checksum recorded, compressed-stream tested and all
tar members traversed. An initial BSD-tar @-operand attempt failed before copying
packages; its verified empty archive is retained, and the corrected ./@types paths
were successfully archived before installation. [Preservation](type-declarations-preservation.json).
The exact current proposal tips and full diffs are in a verified incremental
bundle with strict isolated pack/object proof. [Recovery](type-package-heads-recovery.json).

Required fresh Backend/Frontend/Desktop checks still gate integration. Reconcile
#225/#227 only after the targets are safely integrated, retaining exact-tip and
open-dependency guards for branch retirement. No real records or external provider
service are used. Original-primary coverage recovery remains unresolved.
