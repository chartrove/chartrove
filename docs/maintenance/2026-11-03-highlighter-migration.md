# Angular 22 highlighter migration for #308

Update ngx-highlightjs 7.0.1 to 14.0.1 and its required highlight.js selection
11.7.0 to 11.12.0. Seventy-one other direct versions and unrelated lock bytes
remain unchanged. Angular 22 and Node 26 stay pinned. The old unused transitive
line-number package is archived before removal; no template uses its directive.

Sixteen existing JSON-pipe bindings now specify the required JSON language.
Binary keeps its JSON/XML restriction through the new highlightAuto directive.
The application registers the XML language and removes the obsolete line-number
loader option. No data model, raw record, backend, database or Wails contract
changes. This is a corrected replacement for the bare historical version patch.

Eight consumer cases exercise real fallback/Binary DOM: literal zero/false/
Unicode, record refresh, JSON/XML text, and embedded markup remaining inert.
The original cases pass on v7; the bare upgrade fails compilation on the obsolete
loader import. After template migration, six assertions run before v14's
animation-frame DOM writes. The final tests wait for the actual expected text
and highlighted token, with all content/escaping assertions retained. Identical
final test bytes pass all eight cases on both versions. The old comparison uses
archived package overlays and the prior Binary binding; it is a consumer check,
not a canonical old full-release build.

The overlay restores candidate content/stat metadata and Yarn metadata exactly.
macOS changes provenance on the two newly installed candidate package roots;
all non-provenance xattrs match. Both values and the failed strict wrapper receipt
remain retained. No security metadata is rewritten. Historical installed-package
archives remain unchanged.

Official registry SHA1/SHA512 integrity passes. All 31 wrapper and 1,569 engine
published files match their installed counterparts. Seventeen mapped TS sources
match official commit `582d35941b7f4d3f739d8f3aa7ab9c186d970c4d`
after explicit CRLF-to-LF normalization. Three generated bundle indices are
separately qualified under archive integrity. Initial path/line-ending verifier
failures remain retained; raw byte equality to Git sources is not claimed.
The full upstream suite was not run; actual consumers are the tested boundary.

Frozen install, ten Node checks, 529 browser tests (one existing skip), production
and Storybook builds pass. Budgets, coverage thresholds and original absent
coverage/documentation state are unchanged. Fresh native packaging and synthetic
GUI validation remain pending before protected integration.

[Validation and retained diagnostics](highlight-validation.json),
[installed preservation](highlight-installed-preservation.json),
[exact Git recovery](highlight-head-recovery.json).
Raw evidence: `dist/validation/2026-11-01-consolidation/highlight-review/`.
