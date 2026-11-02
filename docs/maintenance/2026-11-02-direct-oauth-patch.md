# Direct OAuth client at the reviewed patch

#224 proposes @panva/oauth4webapi 1.2.1. The publisher's registry deprecation and
shipped README state that the package moved to oauth4webapi. The scoped 1.2.1
artifact contains only re-exports plus a floating ^1.2.1 dependency; the complete
historical lock actually selects underlying oauth4webapi 1.4.1.

This corrected replacement directly pins **oauth4webapi 1.2.1**, updates the two
service import specifiers and replaces only the corresponding lock entry. The
deprecated wrapper is retired; its incidental 1.4.1 delegate is not adopted or
claimed implemented. This is the reviewed historical patch, not a claim of the
latest OAuth SDK or production-provider acceptance. All 73 other direct installed
versions, other manifest fields and all remaining lock bytes are unchanged.

Both scoped artifacts and the direct artifact match registry integrity digests.
The direct package has no dependencies or installation hooks. Its complete shipped
JavaScript is byte-identical to installed 1.2.0 except for the version string and
one signing-key error message; algorithms and exports are unchanged. Declarations
change only a documentation comment. MIT license bytes are identical, and the
built license notice retains the exact author/text under the new package name.
[First-party source comparison](https://github.com/panva/oauth4webapi/compare/v1.2.0...v1.2.1).

The complete current installed scoped leaf (five files matching its original
registry artifact), file metadata and .yarn-integrity were archived before the
frozen install. Native tar preserves ACLs/xattrs/flags; zstd integrity and complete
tar traversal pass. [Installed-leaf preservation](scoped-oauth-preservation.json).
The exact current #224 tip is also preserved with all five required unique Git
objects and reachable main prerequisites. [Git recovery](oauth-patch-head-recovery.json).

Frozen install, production build, ordinary full browser suite (423 passes, one
existing skip, exit 0), five bridge tests and Storybook pass. The real-SDK synthetic
transport regression still exercises token forwarding/storage and prevents token
console logging; it contacts no actual provider. Original documentation metadata
is restored. Against the #353 native application, 22,283 generated files match.
Only main JS/map, its index reference, and the renamed/relocated MIT notice differ,
with the two Finder snapshots explicitly excluded from generated output. Complete
source-map contents change only the two import specifiers and two reviewed SDK
strings. No clinical algorithm, database/Wails contract or data migration changes.

[Validation and artifact evidence](oauth-patch-validation.json). Existing build
warnings and original artifacts remain retained. The [fresh native package](native-f30970052-validation.json)
passes build, ZIP CRC, strict extracted signature and all ten synthetic WebView
checks using a new explicit unencrypted data directory with providers disabled.
Original outputs match contents, inodes, permissions, ownership, birth times,
mtimes and all xattrs after restoration. All three protected checks still gate
integration. The original proposal remains open until that integration.

Integration: #357 merged at `6ba97e57c4e0cd7b21848b9368b76c5cb0bec92e` after all
three required checks passed. #224 was individually superseded with its unadopted
1.4.1 delegate explicitly recorded; its exact archived head was manually retired
after dependency checks and an expected-tip lease. [Receipt](oauth-patch-retirement-1.json).
