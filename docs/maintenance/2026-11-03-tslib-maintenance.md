# TypeScript runtime helper maintenance

Replace #252's obsolete-base patch with its isolated current lockfile change:
tslib 2.4.1 to 2.9.1 for the existing ^2.0.0/^2.1.0/^2.3.0/^2.4.0 selector.
package.json, the other lockfile bytes and all other 72 direct installed versions
remain unchanged. Existing separate tslib selectors remain intact. No Go,
database, Wails, stored-record or application source change is included.

The downloaded npm archive passes the registry's SHA-1 and SHA-512 values.
Published CJS/ESM runtimes, declarations, package metadata, copyright and license
are byte-identical to the official v2.9.1 tag. Installed files match that package.
The release's extension-rewrite helper accepts additional mappings; the upstream
Node suite passes all 38 tests without modification, skips or todos. Upstream
Rollup/Vite/Webpack suites were not run; Chartrove's actual production and
Storybook consumers pass instead.

The canonical frozen install, ten existing Node checks, 500 browser tests (one
existing skip), desktop production compilation and Storybook all pass with
unchanged compiler inputs. Original frontend/coverage and documentation.json
absence are preserved. Existing compiler, CommonJS, budget and Node warnings
remain in full logs. [Validation](tslib-validation.json).

The exact historical head ea71bafafa7536b902f199a9ea55ca58442fc4f2 is preserved
in a checksum-recorded bundle. Its prerequisites remain reachable from main;
strict pack validation proves all four unique objects are present in the actual
archive bytes. The original intake bundle and its earlier branch tip remain
linked in the [recovery receipt](tslib-head-recovery.json). The historical branch
stays live until protected integration and fresh PR dependency checks.

Fresh packaging at source b4d3e714fb3a3ccad634f5d3a23c4776ea6aa675 passes
ZIP CRC, strict signature verification and all ten synthetic native checks.
All 22,287 frontend files match the qualified build. Original output contents,
stat fields and all xattrs restore exactly, with no provenance changes.
[Native receipt](native-b4d3e714f-validation.json).

Protected integration remains pending. #252 stays open until integration; its
attributed historical discussion and archived exact head will remain recoverable.
Raw evidence: dist/validation/2026-11-01-consolidation/tslib-review/.
