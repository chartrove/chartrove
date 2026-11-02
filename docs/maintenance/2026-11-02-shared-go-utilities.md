# Shared Go utility alignment

This bounded current-graph replacement covers #233 (deepmerge 0.2.2), #265
(google/uuid 1.6.0), #266 (strcase 0.3.0) and #269 (Sprig 3.3.0). All four complete
PR diffs affect only root go.mod/go.sum. The desktop already selected UUID 1.6.0
and Sprig 3.3.0; retain those and advance only deepmerge/strcase there. The proposed
current heads remain open until the replacement passes protected CI and merges.
[Exact heads, incremental bundle and strict verification](utility-heads-recovery.json).

Upstream comparisons were read at their version tags. deepmerge's release changes
no Go files. UUID changes hash/UUIDv7 support; this does not change the application's
UUIDv4 ID contract. strcase changes acronym/Camel handling, which the model generator
uses. Sprig changes dictionary/crypto helpers and its dependency graph. No new
application feature or exploit fix is claimed from these utility updates.

The root graph's necessary Sprig closure adds dario.cat/mergo 1.0.1 and advances
semver/v3, xstrings, copystructure, reflectwalk, decimal and cast. Desktop retains
its existing higher transitive selections. Go version directives and every active
replacement are identical before/after: local backend, recovered provider-library
provenance and exact-version SQLCipher replacement remain intact. No GORM/driver,
JWT/JWX, network/crypto family, Wails, frontend manifest or model source update.

To check the generator boundary without rewriting tracked source, compile the
unchanged generator against old and new module graphs, run each in a new isolated
directory using only the two public FHIR metadata inputs, and compare all output.
All 57 generated files are byte-identical. This checks field names, JSON/GORM tags,
table names and generated statements across the dependency change; it neither
runs a data migration nor claims the generator covers every possible FHIR issue.

Both module verifications and vet pass. The full serial root backend suite and
backend executable build pass. Desktop host tests, encrypted embedded-backend
integration (unchanged 120-second limit), model/IPS/PDF tests and host build pass.
Fresh pinned disposable PostgreSQL startup/persistence/reopen checks also pass.
The previously generated related-versions input was preserved before preparation.
All database fixtures use new synthetic storage; no real family/purchased-app
paths are read. Go source and generated database models are unchanged.

Evidence: dist/validation/2026-11-01-consolidation/utility-* includes module deltas,
old/new generator hashes, complete upstream comparisons, logs and exact recovery
tips. Required Backend/Frontend/Desktop CI and a current native package must pass
before claiming integrated current-source acceptance. The earlier 06d71f22b native
receipt predates these module changes. Original-primary coverage recovery remains
unresolved; no blanket clean-vulnerability or all-history-preserved claim is made.
