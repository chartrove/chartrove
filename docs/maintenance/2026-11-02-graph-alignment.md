# Align server graph dependency with the validated desktop (#259)

The root module advances dominikbraun/graph v0.15.0 to v0.23.0, which the
current desktop graph already selects. Only one go.mod version field changes
and two checksum lines are added; all prior checksums remain. The complete
selected root module graph has exactly this one version change. The target has
no runtime module dependencies, and its Go 1.18 minimum is within the existing
Go toolchain. Published module sums and retained archive hash are recorded.

The upstream comparison spans 90 commits and 42 files, including graph storage,
traversal and algorithm changes. This is not a claim of universally identical
behavior. Application code uses directed rooted graphs, vertex/edge insertion,
adjacency/predecessor maps and DFS. Existing repository tests validate the
flattened encounter graph and the NDJSON relationship regression. Both pass
before and after with the same assertions and no skips (package durations
5.358 and 5.933 seconds). All upstream graph subpackages pass their tests.
Both module verifications, database vet and server build also pass.

Backend, frontend and desktop source remain unchanged. The desktop's entire
production import list, including versions, is byte-identical and its module
files are unchanged. The separately qualified native 9add19dd1 artifact remains
applicable; no new native build is claimed. No schema, migration, relationship
rewrite, provider connection or real-record operation is introduced.
[Validation](graph-alignment-validation.json).

The exact current #259 tip d98466fa34975c6ea1bc633e20cad54d9f58fd6d is in a
checksum-recorded incremental bundle. Strict validation and original-pack
offsets prove all four unique objects are included, with main-reachable
prerequisites. The original intake tip remains in history.bundle as well.
[Recovery](graph-alignment-head-recovery.json). The proposal and its branch
remain live pending protected-main checks and integration, individual closure,
and fresh PR dependency/expected-tip checks before retirement.

Raw release/comparison metadata, module package, before/after outputs and
receipts are retained in dist/validation/2026-11-01-consolidation/graph-alignment-review/.
The evidence manifest records hashes; execution-time temporary paths are kept
verbatim after the evidence directory's atomic move to permanent storage.

## Related backlog reconciliation

#182 is individually closed only as a duplicate of still-open #53. Both report
the Cambridge AdministrativeGender m rejection, already present in #53's
April 2025 discussion. Synthetic R4 Patient probes in both current Go module
graphs accept the four standard codes and reject M/m through the same path.
No normalization or defect correction is claimed. The canonical report retains
the unresolved import work. [Receipt and links](administrative-gender-duplicate-182.json).

Read-only review of #194 confirms broader procedure/template requests and an
unwired native timeline note input. That is insufficient to establish the
reporter's exact Save failure; it stays open. Historical #217's reserved-name
policy is also unimplemented and remains open. No broad feature retirement is
inferred from the continuation instruction.
