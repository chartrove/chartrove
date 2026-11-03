# PDF dependency compatibility

This replacement implements #267's Maroto 2.4.4 in both Go module graphs. Its
required PDF provider changes from jung-kurt/gofpdf to phpdave11/gofpdf 1.4.3;
pdfcpu becomes 0.16.2. Required barcode, go-tree, TIFF and runewidth dependencies
and their metadata are updated together. The obsolete explicit requirement
for the old PDF engine is removed only after both import inventories
prove it unused; historical checksums and source history remain preserved.

The library requires Go 1.26.1. Both declared module minimums explicitly increase
from 1.26.0 to 1.26.1; the actual selected Go 1.27.3 toolchain and CI pin stay
unchanged. This is the required patch-level floor for this library, not an
implicit toolchain switch. FHIR models, ORM family, exact SQLCipher source,
provider mirrors, Wails and authentication contracts remain unchanged. Complete
MVS differences (14 per graph) and imported-module inventories are retained;
imported changes are restricted to the reviewed PDF family.

## Consumer checks and test corrections

The first isolated PDF suite failed 15 structural snapshots. Full recursive
comparison identifies the new all-border representation: string "1" becomes
numeric bitmask 15, matching the published library's old/new Full constants.
Those exact properties are updated in 15 snapshot files. No assertion is removed
and no snapshot is blindly regenerated.

One prior patient fixture supplied a JSON object array where the existing derived
Name field expects strings. It rendered an empty name; the new provider allocates
zero height to that empty text instead of the old reserved row height. The
fixture now uses ["John Doe"], and its snapshot explicitly requires that text.
Both old and new renderers retain the original populated row geometry. This is
a fixture correction, not a claimed production patient-name repair.

Thirty structural cases pass without failures/skips on the old renderer, new
root graph and new desktop graph, using the corrected fixture and explicit border
representation transform. Private test configurations isolate their snapshots;
the original checkout snapshots remain byte-identical during those probes.
An initial desktop probe used the root module's macOS path casing, so Go ignored
its source overlay and retained the malformed fixture. The corrected overlay
uses the exact import directory from the desktop graph and passes the same cases.
Original failed controls and their logs remain preserved.

An independent synthetic consumer calls the actual PDF renderer outside the
source tree on all three graphs. It includes the embedded Fasten logo, patient
and source identity, dates, a zero-valued observation and one or 120 allergy rows.
Poppler extracts identical normalized logical text and identical token counts
across graphs, with one and seven A4 pages respectively. The first-page render
was visually reviewed for the small corrected document: logo and table contents
are present and readable. The PDF provider changes wrapping; layout-preserving
text extraction and image bytes are not asserted identical. Initial attempts to
compare layout text byte-for-byte and to match a phrase split across table lines
failed and remain qualified diagnostics. All source content, including the final
synthetic row and zero value, remains in the normalized comparison.

Compiled focused-test binaries are removed after exit with checksummed disposal
receipts. Compact PDFs, PNGs, text, helper source, logs, input hashes and recovery
evidence remain retained under
dist/validation/2026-11-01-consolidation/maroto-review/.
Relocation receipts preserve the original command paths without rewriting logs.
No real health records, purchased-app storage or backup data were opened.

Applied graph verification, vet, IPS/PDF, desktop and encrypted embedded results
are recorded in the [compact validation receipt](maroto-validation.json). Protected Backend, Frontend and
Desktop checks remain required before integration. Full local frontend/Storybook
and native packaging are not repeated for this Go-only batch. The unchanged
50 GiB admission remains in place; observed capacity is near that floor. No
storage cleanup or cache/backup pruning is performed. No new native UI/package
acceptance is claimed; the last qualified local package predates later dependency
updates.

## Provenance and original proposal

The exact #267 tip ac49aa46139a5e951aaac6cdd2a2d6d458b0ea4f has one unique
Renovate-authored commit changing only go.mod/go.sum. Its verified Git bundle is
2,472 bytes with SHA-256
e804aa6e86d299338d72514c6a2b836ae145746ea5722fdbe73863ead29b553f.
Strict pack validation proves all four unique objects occur in the original
archive bytes; prerequisites remain reachable from main.
[Recovery receipt](maroto-head-recovery.json).
The original PR and branch remain live until protected replacement integration,
fresh individual content/head checks and guarded ownership/consumer retirement.
