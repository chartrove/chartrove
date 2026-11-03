# Devcontainer Node feature v2 replacement for #318

Change only the Node feature key from `:1` to `:2`. Existing Node 26.13.0,
Go 1.27.3, Chrome feature, base image and other configuration stay unchanged.
JSONC parsing and a semantic comparison confirm that single key change. The
excluded configuration file was edited through its Git index blob without
materializing it: sparse patterns and the original skip-worktree flag remain
identical.

The published feature is 2.1.1. Its OCI manifest digest and both config/layer
digests pass independent SHA-256 checks. All four extracted feature files match
their official GitHub Git blob IDs, including the unmodified installer. The
[first-party metadata](https://github.com/devcontainers/features/blob/main/src/node/devcontainer-feature.json)
allows a specific Node version and lists the feature's package-manager defaults.

An isolated linux/arm64 image on the existing Debian base installs the exact
published feature with its real option defaults and Node 26.13.0. Node 26.13.0,
npm 11.20.0 and the repository's unchanged pinned `npm exec` Corepack 0.35.0 / Yarn
1.22.22 command pass. The initial harness required global Corepack, which is not
provided with this Node version. That assertion failed after Node/npm succeeded;
the qualified harness uses the actual Makefile command. The original context,
failure and logs remain intact; no installer/default was changed to make it pass.

Only published feature files, a Dockerfile and the current frontend manifest
entered the image context. No checkout, credentials, records or volumes were
mounted. The uniquely labeled image remains local; no registry push occurred.
The full Go/Chrome project devcontainer and amd64 installation were not rebuilt.
This is scoped Node-feature acceptance, not a claim about those separate paths.

Application source, Go graphs and emitted assets are unchanged. Existing #388
frontend validation and source-identified native package remain applicable;
all three required protected checks still gate this configuration integration.
The exact historical proposal and all four unique objects are archived and
strictly verified, with prerequisite ancestry and the original intake reference.

[Validation and platform limits](node-feature-validation.json),
[Git recovery](node-feature-head-recovery.json).
Raw evidence: `dist/validation/2026-11-01-consolidation/node-feature-review/`.
