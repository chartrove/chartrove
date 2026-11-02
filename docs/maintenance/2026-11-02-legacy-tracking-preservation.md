# Standalone recovery for cached legacy history

The 17 nonsymbolic legacy-fork tracking refs and symbolic HEAD are unchanged from
intake. Their complete history is now preserved in a private full Git bundle with
no retained-main prerequisites. Strict isolated index-pack verification succeeds
without alternates, and all captured tips are present. The original local config,
packed refs, loose tracking refs and their reflogs are separately preserved using
native tar ACL/xattr/flag handling, checksum verification and full traversal. The
archived config bytes match the recorded pre-branch-change hash. Raw metadata and
the full bundle remain local with restrictive permissions; only the receipt is
tracked. [Paths, exact tips and checksums](legacy-tracking-preservation.json).

No cached ref or remote configuration is deleted by this preservation step.
The configured URL remains the old karamhouse/fasten-onprem fork. These aliases
are distinct from current origin heads and are not counted as already retired.
All 30 tag OIDs and existing release history remain intact.

[Unique-work metadata review](legacy-tracking-unique-review.json) identifies one
old README patch, one temporary ARM Dockerfile workaround, five manual-input
prototype commits and 29 export-summary commits. No old implementation is revived
or declared implemented from those subjects alone. The export/narrative lineage
still needs reconciliation with the pending summary/text_field_extract decision;
those exact live origin tips are retained. The older dependency snapshots have
current integrated/replaced dispositions elsewhere in the consolidation ledger.
Further alias decommission requires explicit purpose/dependency review, with this
complete recovery archive available; it does not authorize dropping unique work
or modifying the purchased application's data.
