# Existing lab result edits (#437)

Base: `5bbbf50277464362ae814cb9599b06e09dc21f27` (accepted #436).
This is a correction to the inherited Fasten lab editor, not a replacement lab
subsystem or a resolution of historical #186 (manual date entry/display) or #158
(report loading/reference handling). Their original discussions remain intact.

## Reproduction and correction

On the base, four synthetic browser regressions failed: the editor chose the
first same-ID report from another source, matched observations by repeated display
labels, prefilled by labels, and submitted newly generated widget resources in
place of the original records. A value edit could therefore replace an original
ID and omit dates, subject/reference fields, notes and other clinical metadata.

The editor now selects the report and observations by their loaded native
`source_id/resource_type/resource_id` tuple, fetches source-qualified original raw
text, and binds a question only to a unique exact coding-system/code match. The
existing LForms `LOINC` alias is recognized; labels and widget-generated IDs never
select a record. Ambiguous or unsupported mappings stop before saving.

After the real widget has loaded, its initial export is captured once. Save
compares like-typed values against that baseline and submits only changed answers.
The existing locked Microsoft `jsonc-parser` 3.3.1 is declared as a direct
dependency to apply a single JSON scalar edit to the original text. Unrelated
fields, formatting and numeric precision are not reserialized. Quantity edits
change only `valueQuantity.value`; units and quantity metadata stay intact.

The parent checks the complete returned plan against the selected source's loaded
observation tuples before dispatch. The existing owner-only PATCH API remains the
authority; client selectors are not authorization. Existing UUID/source predicates,
local-edit provenance, grant eligibility and revocation enforcement are unchanged.
No schema, server authentication, credentials or deployed configuration change.

## Deliberate limits

- Existing Quantity (without comparator), Integer, Boolean and String values only.
  New answers, clearing, type changes, unit changes/conversions, repeated codes,
  complex/coded values, absent values and malformed/duplicate-property JSON are
  refused. An unsupported supplied observation makes the panel unavailable for
  editing; inspection remains available in the source-specific record view.
- Existing editable numeric values must round-trip through JavaScript without
  mathematical loss. Unrelated original decimals retain their exact bytes. The
  widget's numeric input behavior is unchanged; this is not arbitrary-precision
  entry or a clinical date-editing feature.
- Preservation is relative to the loaded snapshot. The existing owner replacement
  API has no compare-and-swap/version predicate. Reload after edits elsewhere;
  concurrent full-record writes are not merged by this change.
- Multi-observation requests remain independent transactions. A request failure
  can leave a partial save; the UI requires reload before retrying. No batch
  atomicity or deployed operational acceptance is claimed.

## Focused evidence

Permanent regressions cover same-ID cross-source selection, duplicate labels and
codes, original IDs and unrelated fields, unchanged/zero/false values, exact raw
transport, rejection of forged modal targets, and an actual published LForms edit.
The encrypted embedded-backend test sends the frontend-shaped text payload through
the existing native API and checks owner success, delegate/admin/anonymous denial
on the owner route, unchanged other-source records, local provenance and revoked
delegate denial on the separate eligible delegated route.

Original red, green and validation receipts are retained privately under the
existing consolidation evidence root, `lab-edit-preservation-20261106`.
Protected CI and independent review are recorded on the replacement PR. Merge
requires separate owner approval; #425 live acceptance remains separately gated.
