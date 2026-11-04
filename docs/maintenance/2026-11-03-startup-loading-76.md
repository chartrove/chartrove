# Startup loading feedback (#76)

Issue [#76](https://github.com/karamhouse/chartrove/issues/76), imported from
Fasten #271, requests a loading/splash page or logo while the SPA is loading.
The complete report has no further description or discussion. Current web,
desktop and legacy cloud source entry templates all have empty app-root elements.

The existing Chartrove SVG and a plain Loading Chartrove status now appear inside
app-root, styled inline so this shell does not depend on the application CSS or
Angular bundles. It uses a fixed white background and dark text, no animation,
percentages or record details. JavaScript-disabled pages instead explain that
JavaScript is required. Ordinary Angular client bootstrap replaces root content;
no route overlay, authentication change or persistent loading flag is introduced.
Existing styles remain scoped to the disposable shell. The logo has empty alt
text because the adjacent status already names Chartrove.

All three index templates receive the same small shell. Their complete existing
script attributes/content/order and link contracts are verified unchanged,
including the Wails runtime/bridge, base-path calculation and LForms scripts.
Editing the retained cloud template does not enable cloud authentication,
services or publishing workflows. No dependency or toolchain changes.

An isolated system Chrome profile verifies all 24 combinations: exact old/new
source HTML, three entry variants, 1200x800/320x480 viewports and JavaScript on/off.
The twelve old controls confirm the missing shell; twelve corrected controls
verify status/JavaScript guidance, loaded existing logo, viewport bounds and no
horizontal overflow. The small-screen screenshot was visually inspected. The JavaScript-disabled
style was moved to head noscript for valid metadata placement; the identical
24 cases pass again on final source bytes, with a byte-identical screenshot. These
are pre-bootstrap source-shell checks, not full compiled application startup,
boot failure recovery, clinical accessibility certification or packaged native
acceptance. Inherited external CDN/native runtime endpoints are blocked or absent
on the private synthetic server; their code is preserved, not activated.

The browser/profile/output scratch is disposed and compact logs, screenshot and
receipts remain. Free disk is about23GiB; the 50GiB full-workflow reserve and live
Angular cache=false remain unchanged. No full local build or storage cleanup is
run. Existing protected Backend/Frontend/Desktop validation is required before
integration. [Compact validation](startup-loading-76-validation.json).
Raw evidence: dist/validation/2026-11-01-consolidation/
post-dependency-issue-review/issue76/. Individual issue disposition remains
pending protected integration and a fresh full-content check.
