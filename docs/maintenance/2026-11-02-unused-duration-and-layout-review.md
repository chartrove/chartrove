# Unused duration dependency and superseded layout proposal

The humanize-duration package has no consumer in any tracked source. The only
match outside lock/evidence is its direct manifest declaration. This bounded
cleanup removes that line and its single leaf lock block, including the unused
installed 3.28.0 package. All 73 retained direct versions and every remaining lock
byte stay unchanged. #244's proposed 3.35.0 upgrade is explicitly not implemented;
it becomes obsolete when this removal integrates. No new runtime module is
selected, and no duration behavior or production source changes.

The original installed leaf and Yarn metadata have native ACL/xattr/flag archives,
checksums and full traversal verification. The exact historical #244 tip and all
four required unique objects are recoverable from a verified incremental bundle
with retained-main prerequisites. [Preservation](humanize-installed-preservation.json),
[Git recovery](humanize-head-recovery.json).

Ordinary frozen installation, all ten Node checks, 423 ordinary browser tests
(one existing skip, exit 0), production and Storybook pass. Every one of 22,287
produced application files matches the current validated native frontend. No new
native run is claimed for identical output. Documentation was absent before and
remains absent; newly generated metadata stays private. Existing build/peer/module
inference warnings are retained without relaxation of tests or thresholds. The
browser launcher sends SIGKILL after its existing two-second shutdown warning;
the ordinary command still exits 0, and its full log is retained.
[Validation receipt](humanize-retirement-validation.json).

## Historical source-filter hover width (#210)

Current main already contains inherited commit
`30b1cb05ffb156eb789fc6fe749310acca54f5ad`, which removes the source-filter
width:100% hover rule and width transitions. Current custom.scss retains only
horizontal-overflow changes on that hover; the template's normal width is 200px.
The full single-file #210 patch targets the same width-expansion mechanism but
also chooses a fixed 280px width. That additional sizing policy is not adopted.
The historical patch is superseded by the inherited removal; this is not a claim
that every navigation/layout issue is implemented or fixed. Broader issue #167
and small-screen facet issue #99 remain open pending their own reproduction.

Both original head and pinned base, complete original description/discussion,
full diff, inherited-fix diff and current style excerpts are retained privately.
All five unique Git objects and main prerequisites pass strict isolated pack
verification. [Exact tips and archive](source-layout-heads-recovery.json).
No CSS is edited for this disposition. #210 remains open until this protected
checkpoint integrates; both source branches require fresh dependency and exact-tip
checks before retirement. Existing tags and release history remain intact.
