# Practitioner name parsing and payload privacy (#229)

Main at `25b6b1dc1` resolves parse-full-name 1.2.6. The bounded replacement for #229
selects 1.2.7 within the unchanged manifest range. Only that selector's version,
archive URL and integrity change. All other lock bytes and 72 other direct
versions are unchanged; frozen installation through the existing pinned Makefile
target passes. The published SHA-512 matches the archive, and all eight published
files match the installed package. There are no runtime dependencies or install
hooks. Existing LForms, documentation-tool and engine warnings remain.

The runtime patch replaces nickname extraction with a delimiter scanner. Its
source was reviewed directly from the published package. The old upstream
GitHub comparison returns HTTP 404; that failure is retained rather than treated
as a reviewed upstream commit. Sixteen synthetic cases cover ordinary names,
titles/suffixes, comma order, six nickname delimiters, apostrophes, hyphens,
Unicode, a name particle, one-word and absent input. Their complete outputs match
before/after and match the installed 1.2.7 package. This is a compatibility sample,
not proof that every possible name parses identically.

Three practitioner form paths call the parser; another page retains an unused
import. These consumers keep the original entered name as text/display name.
No existing record is reparsed or migrated. Initial dependency-only validation
passes 427 browser tests with one existing skip and exit 0; ten Node checks pass.
The existing Chrome two-second shutdown warning is retained without changing
launcher flags or timeouts. Fresh native packaging and acceptance are pending.
[Validation receipt](name-parser-validation.json).

The initially attempted bare Corepack command did not start because Node 26
does not bundle it. The existing documented Makefile invocation was then used;
no global toolchain change was needed. Both attempts are recorded.

The current historical head `6a1e3d7124ff634bf145c068af6e4ffd17b2550a` differs
from the original intake tip. Its exact one-commit/six-line patch is retained in a
new checksum-recorded bundle. All four unique objects are proved present in the
original archive bytes, with strict pack validation and main-reachable
prerequisites. The earlier intake tip remains in its existing archive. Installed
1.2.6 and Yarn integrity metadata are also retained in a native archive.
[Head recovery](name-parser-head-recovery.json). The historical PR/branch stay
open until protected integration, fresh validation and dependency checks finish.

## Practitioner creation privacy

The same consumer review found full outgoing FHIR, server-response and raw-error
console logging in the practitioner creation page. These three payload logs are
removed. Two synthetic canary regressions fail on those original logs and pass
unchanged afterward. The tests exercise the real component and HTTP service,
verify the submitted original name and success navigation, and preserve a failed
form for a successful retry. The existing initialization test also passes.
Final ordinary browser validation after this correction passes 429 tests with
one existing skip and exit 0. The earlier Chrome shutdown warning is retained.
[Privacy regression and full-suite receipt](practitioner-privacy-validation.json).
This change does not alter practitioner formatting or serialization.
