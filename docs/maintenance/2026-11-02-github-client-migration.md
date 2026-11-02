# GitHub Gist client migration and cache boundaries (#330)

Historical #330 adds an unused v93 requirement while the application still
imports v54. This corrected replacement migrates the actual anonymous Gist
client and its tests to github.com/google/go-github/v93 v93.0.0 in both Go
module graphs. Its required go-querystring dependency moves from 1.1.0 to 1.2.0;
all other selected versions, toolchains, database replacements and Wails
contracts remain. Existing checksum lines retain their bytes and order, with
six additions in root and five in desktop. These include exact metadata for
already selected test dependency go-cmp 0.7.0 and querystring's 0.6.0 go.mod;
the selected go-cmp version stays 0.7.0.

The target requires Go 1.26, matching the application minimum. The new
error-returning options constructor is handled explicitly. Anonymous access,
the default HTTP client and API version 2022-11-28 remain; no authentication,
proxy, timeout, provider or dashboard units policy is added. Official release,
README and constructor sources are retained with the package archives and
authenticated module checksums. The upstream comparison contains 1,072 commits
and a capped 300-file response; this is not an exhaustive review of that file list.

## Reproduced cache defects

Local synthetic HTTP and private temporary files reproduce a panic for JSON
`null` and an overwrite outside the dashboard cache when response metadata
supplies ID `../outside`. The normal JSON object case already passes. The same
assertions pass with v54 after rejecting nil objects and unsafe cache IDs,
then pass with v93 after adapting only the import and constructor API.
Dashboard title, source and ID metadata remain intact, and the outside sentinel
retains its original bytes. No production Gist or health record is accessed.
The existing dashboard schema-validation TODO remains unresolved.

The original overlay invocation that entered no tests and an initial assertion
API compilation mistake are preserved as failed diagnostics, not passing
validation. Canonical Go package paths and corrected assertions establish the
actual red baseline. Subsequent v93 constructor compilation diagnostics and the
upstream test checksum failure/correction also remain in the evidence store.

## Validation and preservation

Full handler suites in both Go graphs, desktop host and encrypted embedded
backend tests, both full vet checks and module verifications pass. Upstream SDK
and querystring suites pass after retaining the exact missing test checksum.
Both cache race checks and the server build pass. No timeout or assertion
is relaxed. Fresh native packaging at f22b03676 reuses the qualified 5ad6c590a
frontend; all 22,287 files match the reused and native asset trees. ZIP CRC,
strict signature and ten synthetic WebView checks pass. Binary metadata confirms
v93.0.0 and no v54. Original output contents/stat fields and all xattrs restore
exactly, with zero provenance changes. The ZIP SHA-256 is
1a84852664ad9df8b3048351ef1e9910c7fbfa50a8c482f01023c5caba73bd43.
[Native receipt](native-f22b03676-validation.json). Protected integration is
complete: #370 merged at 3ec2085ec4e9fac8d34f3e050313c4369a1fbd91 after
all three required checks passed.
[Detailed validation](github-client-validation.json).

[Exact historical head recovery](github-client-head-recovery.json) records tip
77a8aa3270cc3744c066a7f45da90df004285432, its verified incremental bundle,
main-reachable prerequisite and all four unique objects proved in original pack
bytes. Original intake history remains. Historical #330 was individually closed
as implemented. Fresh dependency and
expected-tip checks preceded retirement of its exact archived branch and the
merged task branch. [Retirement receipt](github-client-retirement-1.json).
Raw commands, diagnostics, source snapshots, manifests
and archives are retained under
dist/validation/2026-11-01-consolidation/github-client-review/. Original temporary
execution paths remain verbatim after relocation.
