# Chartrove frontend

This is the Angular 22 desktop/server frontend. Use Node 26.13.0 and the pinned
Yarn 1.22.22 commands in [CONTRIBUTING](../CONTRIBUTING.md), including synthetic
storage and dependency-install instructions. Fasten attribution and licenses are
retained in the repository and application.

From this directory, use the installed CLI:

```sh
node node_modules/@angular/cli/bin/ng.js test --watch=false --browsers=ChromeHeadlessCI
node node_modules/@angular/cli/bin/ng.js build --configuration desktop_prod --output-path=/path/to/new/build-directory
```

Coverage goes to a fresh private temporary directory, printed by Karma. Set
CHARTROVE_COVERAGE_DIR explicitly only for a chosen new report destination.
Run Storybook's documentation build through the pinned package manager so its
Compodoc subprocess can also find Yarn:

```sh
npm exec --yes --package=corepack@0.35.0 -- corepack yarn run build-storybook --output-dir=/path/to/new/storybook-directory
```

Select a new output directory when retaining an existing build.

The obsolete TSLint/Protractor targets and starter test are [preserved and retired](../docs/retired-frontend-tooling/README.md).
There is currently no configured frontend lint or full browser end-to-end command;
those retired commands are not advertised as working checks. Active browser unit,
production compilation and native integration checks remain mandatory.
