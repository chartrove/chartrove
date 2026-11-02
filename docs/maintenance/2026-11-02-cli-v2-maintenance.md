# Server CLI v2 maintenance and unused v3 proposal (#274 / #331)

The root module selects urfave/cli v2.27.7 in place of v2.11.2, with its required
md2man v2.0.7 and smetrics 20240521 versions. Only these three go.mod version
fields and six added checksum lines change; all earlier checksums remain.
The selected module graph also advances optional BurntSushi/toml 1.1.0 to 1.5.0
through CLI's minimum requirement. Neither executable imports that optional
package. Backend application source, SQLCipher/provider replacements, Wails
contracts and desktop module files remain unchanged.

Upstream's [support policy](https://cli.urfave.org/) retains v2 for security/bug
fixes, recommending v3 for new development. The reviewed v2 comparison spans
425 upstream commits and 133 files; this is a long-standing maintenance update,
not a claim that all upstream behavior is identical. Existing start, migrate,
version and help commands keep their names and flags. Before/after compiled
executable probes cover nine help/version/rejection cases, plus four alias and
DEBUG environment cases per version. Invalid flags, invalid booleans and missing
configurations reject before any server or database starts. No migration was run.
Help now follows declaration order, adds explicit help flags and omits unused
argument placeholders. Raw outputs and normalized differences are retained.

Both root and desktop module verification pass. The server executable builds;
application CLI/config tests, CLI vet and upstream CLI tests pass. The entire
production desktop import list, including versions, is byte-identical before
and after and excludes CLI, md2man, smetrics and optional TOML. No frontend source
or lockfile changed. The separately qualified native 9add19dd1 application and
its ten native checks therefore remain the applicable desktop evidence; this
batch does not claim a redundant new native build.
[Validation receipt](cli-v2-validation.json).

## Disposition of the v3 proposal

The actual #331 patch keeps the v2 requirement at 2.27.7, adds v3.14.0 and edits
no Go imports or application code. Every CLI import remains /v2. Thus it does
not migrate the command and adds an unused module. Upstream's
[migration guide](https://cli.urfave.org/migrate-v2-to-v3/) requires explicit
changes to imports, App/Command, callbacks, Context and environment flag sources.
The v2 maintenance target is adopted; the unused v3 addition is explicitly
retired, not represented as an implemented v3 migration. A narrow Renovate
allowedVersions <3 rule keeps v2 bug/security updates available and prevents
recreation of this unused addition. A future v3 adoption needs its own tested
command migration and an intentional policy update. Current official Renovate
JSON schema validation passes using installed Ajv and formats support.

## Preservation and limits

Exact current heads a5cff25a3ad25f31c33490ac917e14ccfeb33917 (#274) and
cf7169db6b9f32acba83133f9325587eb9a1066d (#331) are in separate verified bundles.
Each has strict pack validation, main-reachable prerequisites and all four
unique objects independently proved present in original pack bytes.
[Current v2 recovery](cli-v2-head-recovery.json),
[current v3 recovery](cli-v3-head-recovery.json). Original intake tips are also
retained in history.bundle. Existing install archives, native outputs, coverage,
local metadata, sparse configuration and all original tags remain preserved.

Two diagnostic preconditions were corrected without changing application source:
the initial desktop dependency query used its module directory rather than the
existing executable package; the first graph assertion omitted optional TOML's
selected version. Both errors and corrected evidence remain. All executable
probes ran in private task directories and no real health records were accessed.

Raw evidence is dist/validation/2026-11-01-consolidation/cli-review/ and the two
cli-v2-review/cli-v3-review directories; the manifest records retained file
hashes. Original command paths identify their execution-time temporary location,
which was subsequently moved atomically to permanent evidence. Protected-main
checks and integration remain pending. Neither historical PR is closed or its
live head retired before that integration and a fresh dependency check.
