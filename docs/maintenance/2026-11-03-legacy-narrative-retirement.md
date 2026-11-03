# Retire the legacy narrative repair prototypes

The maintainer selected: “Retire legacy repair; preserve archives and records.”
This retires these prototypes from the maintenance baseline; it does not claim
their feature was implemented or authorize a record repair/migration.

| Historical branch | Exact tip | Disposition |
| --- | --- | --- |
| summary | 62a89d799969b1de8ac208543e89a7da85345a91 | Retire the historical generated narrative/Markdown experiment; preserve all unique work. |
| text_field_extract | 38ea0562ea29cbcfaf2460e7fa04f28d0ec0d63e | Retire its unfinished extraction/reprocessing experiment; preserve all unique work. |

summary is an ancestor of text_field_extract. Their union contains two commits
and 77 objects not reachable from current main. Source review confirms the old
reprocessing loop computes a new model but saves the original temporary wrapper;
it also ignores extraction failures. Neither patch is a reviewed migration for
the current model layout or toolchain. Do not run or revive that repair during
cleanup. Current runtime, attribution, schema and stored records are unchanged.
Existing unrelated extraction/indexing issues are not closed as fixed.

A new small Git bundle preserves both exact tips. Its prerequisite is reachable
from protected main. Strict pack verification, original pack offsets and the
complete unique-object set prove all 77 objects are in the archive itself;
alternates are used only for reachable prerequisites. SHA-256 is
799532b01c07c2b994022f15ae1a6d4162dabfda02a4ca7a962c51c2eff1513c
(19,681 bytes). The original intake bundle also still lists both tips.
[Recovery receipt](legacy-narrative-recovery.json).

Both live origin tips match the archive, and no open PR uses either as a head or
base at review. Their deletion remains pending publication of this disposition,
fresh exact-tip/dependency checks and lease-protected deletion. All 18 cached
legacy refs stay unchanged for their separate alias review. Existing bundles,
tags, sparse settings, local metadata, unknown work and real records are retained.

Raw evidence: dist/validation/2026-11-01-consolidation/legacy-narrative-retirement/.
