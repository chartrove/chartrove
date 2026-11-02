# Original primary coverage recovery

The user granted Full Disk Access specifically to recover the original generated
frontend coverage. The pre-task local APFS snapshot
`com.apple.TimeMachine.2026-11-01-183642.local` was mounted read-only at a private
task mount. Only the authorized repository's frontend/coverage subtree was read.
No health-record or purchased-application storage was inspected.

The recovered tree is retained at
`dist/validation/2026-11-01-consolidation/primary-coverage-before-maintenance`:
349 regular files, 129 directories and 4,974,406 file bytes. All relative paths
and every file's SHA-256 match that exact snapshot. The original index differs
from the separately retained mixed post-maintenance index; the coffinfish archive
was not substituted for this recovery.

Before recovery, native tar preserved the snapshot subtree with ACLs, extended
attributes and flags. The zstd archive passed integrity testing and complete
native tar traversal. Its SHA-256 is
`3c1cfa47426c6035db280ae99b0011aa5a8c75f8a34e7b7f0bd9240a2cd8251b`.
[Receipt and archive paths](primary-coverage-recovery.json).

`tmutil restore` rejected the privately mounted source as outside its recognized
snapshot layout (exit 255); that exact failure is retained. It created no target.
Native `ditto --rsrc --extattr --acl` then copied to a new staging directory;
content verification preceded its atomic rename to the agreed recovery target.
A subsequent metadata audit found that ditto had not retained 350 birth times.
The documented macOS getattrlist/setattrlist APIs restored exact nanosecond
creation times from the read-only source, scoped only to those recovered copies.
Final checks match all contents, creation times, mtimes, permissions, ownership,
flags and ACLs. Original and final metadata manifests are checksum-recorded.

macOS assigned different `com.apple.provenance` values to 478 recovered entries.
These OS security attributes were not rewritten. Exact original values remain
in the source archive and the snapshot metadata manifest; destination values and
all differences are retained separately. The recovery is a verified copy with
this explicit provenance difference, not an assertion of identical inode identity.
The private mount was unmounted afterward; neither Time Machine snapshot was
deleted or modified.

frontend/coverage remains absent. Ordinary Karma runs now create private output
through #344, so future validation does not overwrite this recovered artifact.
The mixed reports, all failed-operation logs and prior recovery manifests remain
retained. This closes the original-primary coverage content-recovery gap; it does
not close unresolved feature requests or authorize legacy narrative migrations.
