# Retire unused upstream Chromatic publishing CLI

Chromatic is referenced only by a frontend development dependency and the retired
upstream CI upload workflow. There is no current application import, package script
or active workflow using it. Chartrove's validation-only CI does not operate the
upstream Chromatic project. Remove the unused publishing CLI rather than restoring
that external publishing operation. Local Storybook and its build scripts remain.

This deletes one package-manifest line and the exact five-line Chromatic 6.19.8
lock block. The runtime manifest and all 75 retained direct installed versions
are unchanged. The frozen install passes; no broad lock refresh, runtime package
update, engine bypass or new dependency is included.

The entire pre-retirement installed dependency archive remains checksum verified.
All twelve installed Chromatic regular files match its bytes; thirteen AppleDouble
metadata records are retained in that archive, not mistaken for installed files.
[Installed-package recovery receipt](chromatic-installed-preservation.json).
The complete current diffs of #239 and #299 affect only manifest/lock files.
Their exact rebased heads are preserved in a checksum-recorded bundle, with native
bundle/prerequisite checks and isolated strict pack/object verification.
[Proposal recovery receipt](chromatic-heads-recovery.json).

Five bridge tests, production frontend and the normal pinned Yarn Storybook build
with Compodoc all exit 0. The original documentation.json state is restored.
All 22,287 production application asset files match those packaged and natively
tested at d19dca35a (#347). The older preserved frontend tree additionally has two
Finder .DS_Store files, which remain preserved; they are recorded separately rather
than silently treating the complete trees as identical. No application source,
Go graph, Wails contract or embedded application asset changed, so the previous
ten-check native package validation remains applicable. Required fresh protected
Backend/Frontend/Desktop checks still gate integration.
[Validation receipt](chromatic-validation.json), [native receipt](native-d19dca35a-validation.json).

After protected integration, retire #239 (6.24.1) and #299 (18.13.6) with the explicit
reason that their unused publishing tool was removed. Those proposed upgrades are
not implemented. Keep the original attributed proposals and recoverable Git tips;
verify current PR dependencies and exact remote tips before branch retirement.
No Chromatic project was configured or published, and no cloud publishing
workflow was enabled. Original-primary coverage recovery and the pending
product/data dispositions remain unresolved.
