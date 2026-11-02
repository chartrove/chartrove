# Minimal replacements for historical proposals

[Exact reviewed heads and complete native Git inventories](2026-11-02-minimal-pr-review.json)
cover both proposals. Original attribution and discussions remain in the imported
GitHub PRs and archived source; the verified history bundle retains their work.

PR #197, imported from Fasten upstream #166, contains 3,092 changed paths. Native
Git enumeration confirms exactly three outside node_modules: the footer template
and new root package/package-lock manifests adding an npm package named docker.
GitHub's capped 3,000-file API alone is insufficient for that scope review.

The only applicable application change removes Bootstrap's fixed-bottom class
from the current footer. It now stays in normal document flow instead of covering
content at the viewport bottom. Existing Chartrove identity, Fasten attribution,
external-link handling and responsive footer styles are preserved. Historical
formatting, generated node_modules, and unrelated root Docker npm dependency
are explicitly retired; they are not represented as implemented application work.
No global layout redesign or new feature is included.

PR #216, imported from upstream #625, changes only docker-compose-prod.yml:
remove the obsolete top-level version field and add its missing final newline.
The current file is updated accordingly. Docker Compose config produces identical
resolved JSON before/after with explicit synthetic IP/port/hostname and an empty
env file; the obsolete-field warning disappears. This validation starts no
services, pulls no images, and reads no mounted database/cache/certificate data.
The retained file still references the historical upstream image; this patch does
not introduce a Chartrove deployment or authorize using it for family records.

The existing footer browser check passes with coverage redirected to private task
output. Production desktop frontend build passes with existing bundle/CommonJS
warnings. All required GitHub checks must independently pass before merge. This
small template/config change does not add tests that merely mirror those edits.
The prior native package and ten smoke checks validate c957b1480, before this
footer edit; no later-source packaged-UI acceptance is inferred from that artifact.

Close #197/#216 only after this minimal replacement merges, explicitly citing both
implemented and retired scope. Their four remote head/base refs still require
fresh dependency checks, archive verification and exact-tip leases before deletion.
