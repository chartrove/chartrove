# Retire unused direct UUID dependency

The direct UUID v9 dependency has no source/script imports. Chartrove's existing
uuidV4 helper uses Web Crypto directly and remains byte-identical. Only Compodoc
and http-auth request an npm UUID dependency; both use the separate v8.3.2 range.

Remove one unused runtime-manifest declaration and its exact five-line v9 lock
block. Do not replace the existing helper, change identifier semantics, or remove
transitive UUID v8. The frozen install passes. All 74 retained direct versions are
identical. After normal Yarn hoisting, both known tooling consumers still select
v8.3.2 with matching module-entry bytes, now from the shared root install.
[Validation and consumer resolution](uuid-validation.json).

The installed v9's 76 regular files match the preserved full dependency archive;
its metadata records are retained there too. Exact current #230/#314 tips and
complete diffs are preserved in a checksum-recorded bundle, with native
prerequisite and isolated strict pack/object verification.
[Installed preservation](uuid-installed-preservation.json), [proposal recovery](uuid-heads-recovery.json).

Bridge tests, production frontend and normal pinned Yarn Storybook/Compodoc builds
exit 0. All 22,287 generated application asset files match the fd5194396 native
package; the two extra Finder metadata files in its preserved frontend tree remain
preserved and are explicitly recorded. Original documentation.json state is
restored. The application's code, Go graph, data contracts and Wails bindings stay
unchanged, so the current ten-check native package remains applicable. Fresh
Backend/Frontend/Desktop protection checks still gate integration.

After integration, retire #230 (v9 patch) and #314 (v14 migration) because their
unused direct dependency was removed; those upgrades are not implemented. UUID v8
remains a tooling dependency, and no claim is made to upgrade or retire its
consumers. Recheck exact tips/dependencies before branch retirement.

The accompanying [reference issue review](2026-11-02-reference-issue-review.json)
classifies #46's informational styling scratchpad for retirement after integration.
#193 remains open because its discussion includes an unresolved custom-certificate
request; its recipe alone does not justify closing it. No new deployment or
external service is configured. Original-primary coverage recovery remains pending.
