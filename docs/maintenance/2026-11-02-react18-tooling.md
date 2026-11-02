# React 18 patch for Storybook tooling

This bounded replacement covers #282's React/DOM 18.3.1 target and the necessary
scheduler 0.23.2 dependency. Only those three current lock blocks change; the
manifest and all 72 other direct installed versions stay unchanged. The historical
patch's incidental loose-envify registry-host change is omitted. React 19 and its
separate proposal remain unimplemented. The [first-party upgrade guide](https://react.dev/blog/2024/04/25/react-19-upgrade-guide)
describes 18.3 as the deprecation-diagnostic step before a React 19 migration.

The three official registry artifacts pass integrity verification. Installed
files match every reviewed package member; no installation hooks are introduced.
Synthetic server rendering passes in both development and production. An initial
diagnostic incorrectly assumed the production ReactDOM.version export equals its
package version. The exact official production CJS/UMD artifacts instead embed
`18.3.1-next-f1338f8080-20240426`, while both package versions are 18.3.1 with the
same publisher gitHead. The initial assertion failure remains retained. The final
diagnostic checks that exact reviewed literal, both package versions and the same
rendered output; no application assertion was changed to accommodate it.

Frozen install, all seven Node bridge/framework checks, ordinary browser tests
(423 passes, one existing skip, exit 0), production and Storybook pass. A nonfatal
webpack stream-close warning followed the browser pass summary; its complete log
is retained. No timeout, test exclusion, coverage threshold or browser flag was
relaxed. Original documentation metadata is restored. Every one of 22,287 produced
application files matches the current native frontend; only prior Finder metadata
is absent. Production Go/Angular code, database contracts and Wails bindings are
unchanged. No new native smoke run is claimed for this identical application
output; the latest validated package remains the #357 runtime.

[Validation, artifact hashes and observed version-label detail](react18-validation.json).
[Original installed React/DOM/scheduler and metadata archive](react18-installed-preservation.json).
[Exact historical tip and isolated unique-object proof](react18-head-recovery.json).
The source proposal stays open until this replacement passes all three protected
checks and integrates; no later React 19 feature is inferred from this patch.
