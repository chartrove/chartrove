# Chart.js 4 maintenance

Replacement for Renovate #238 updates only the chart.js lock selector from
4.4.2 to 4.6.1 and its @kurkle/color leaf from 0.3.2 to 0.3.4. package.json and
every other lock byte remain unchanged. Frozen install proves only those two
installed package versions change; all 117 Chart.js and eight color published
files match their registry archives. SHA-1 and SHA-512 integrity checks pass.
Both packages retain their existing licenses and attribution.

The installed ng2-charts 4.1.1 peer range explicitly accepts Chart.js ^4.0.0.
This does not implement #307's separate wrapper/API migration. Official release
and migration references: https://github.com/chartjs/Chart.js/releases/tag/v4.6.1
and https://www.chartjs.org/docs/latest/migration/v4-migration.html.

The direct update reproduced TS2416 in grouped-bar and simple-line widgets:
their typed options overrode an unspecialized generic base. Supplying the existing
bar/line type parameters fixes compilation, matching the doughnut widgets'
established pattern. These are erased TypeScript annotations; runtime options,
query processing and raw FHIR data remain unchanged. No casts, compiler flags,
test assertions or budgets were weakened.

The identical four populated chart consumer cases pass on old and corrected new
versions (spec SHA-256 2780b3cf01946e7774413c8bd69445bc93e278a5bb75d533a8685604b358bb73).
They exercise actual canvas creation, parsed dataset geometry, chart updates and
disposal for donut, dual gauges, line and grouped bars. The initial incompatible
upgrade's compile failure remains preserved. Local Node: 14 passed, two Linux
filesystem-dependent skips. Required full validation runs in protected CI.

Host free storage was approximately 42 GiB, below the unchanged 50 GiB full-run
admission reserve. No local full browser suite, production/Storybook compilation
or new packaged native app was run. Focused browser scratch was disposed only
after its producer exited. CI must validate full browser/production and native Go
build/test behavior before integration. Packaged native chart UI and a new
Storybook build remain unverified; prior compiled artifacts are not presented as
having incorporated this runtime dependency update.

The exact historical PR tip has a strict verified Git bundle with all four unique
objects and a main-reachable prerequisite. Original installed packages and Yarn
metadata have a native archive whose extracted payload hashes all match. An
initial tar invocation treated the scoped package's leading @ as archive syntax
and failed; it changed no source files. Its partial archive/error are retained;
the qualified retry used explicit ./ paths and passed.

Raw intake, patches, archives, integrity checks, installed-version delta and
old/failure/new logs are in dist/validation/2026-11-01-consolidation/chartjs-review/.
The directory was moved intact; relocation.json explains original receipt paths.
Recovery: [historical tip](chartjs-head-recovery.json),
[installed state](chartjs-installed-preservation.json). #238 remains open until
the protected replacement merges and a fresh full-content/head guard passes.
