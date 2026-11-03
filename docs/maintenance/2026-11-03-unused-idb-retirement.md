# Retire unused direct IndexedDB helper

The direct idb 7.1.1 package has no source consumer. Review covers 847 tracked
code/style/workflow/shell files and 859 installed top-level/scoped package
manifests. No application IndexedDB implementation or installed package consumer
was found. The only lock selection is the direct `idb@^7.1.0` leaf.

Remove its single manifest declaration and exact leaf lock block. Every remaining
lock byte and all 72 retained direct versions stay unchanged. This retires #303's
unused upgrade proposal; idb 8 is not implemented. Browser offline health-record
storage (#19) and form-draft persistence (#29) remain separate unimplemented work.
No browser storage, database, record, Go or Wails contract is changed.

The original installed leaf and Yarn integrity metadata are preserved in a native
ACL/xattr/flags archive. Full archive traversal and extracted file hashes pass.
The exact historical head and all five unique objects pass isolated strict pack
verification; prerequisite ancestry and the original intake reference are retained.

Frozen installation, ten Node checks, 523 browser tests (one existing skip),
production and Storybook builds pass. Existing coverage/documentation absence is
preserved. No threshold, production budget or timeout changes.

Exact old native frontend reuse failed and remains a failed diagnostic. The
only difference is the main source map. Runtime JavaScript, runtime maps, index
and all other application assets are identical. The precise source-map difference
is retained in the receipt. A fresh package includes the current source map.

The historical proposal stays open until protected integration. Retiring its
branch requires a new dependency review and exact-tip lease; archive creation
alone is not retirement authority or implementation evidence.

[Consumer review](idb-unused-consumer-review.json),
[installed preservation](idb-installed-preservation.json),
[Git recovery](idb-head-recovery.json), [validation](idb-validation.json).
Raw evidence: `dist/validation/2026-11-01-consolidation/idb-review/`.

Fresh package source `4414fd36a3987be7aa7617f09fdef3045bd5272e` passes ZIP CRC, strict extracted signature
and all ten synthetic native checks. Its 22,289 compiled/reused/native assets
match exactly. Original output content, stat metadata and xattrs restore exactly.
ZIP SHA-256 `0121a204e83e9c640379110613f532015edb02b2b1a075ed66b62efd5aec924b`.
[Native receipt](native-4414fd36a-validation.json).
