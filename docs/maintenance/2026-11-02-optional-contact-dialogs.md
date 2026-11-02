# Optional address and contact edit-dialog repair

Issue #105 reports an address `.line` null/undefined exception. On current main
`03dbba999`, both medical-record edit dialogs reproduce that error when a resource
has an empty address array. Practitioner conversion also turns an absent address
into an empty array. Optional contact arrays cause separate `.find`/`.findIndex`
exceptions during opening and submission; empty practitioner names/qualifications
have the same first-element dereference defect.

Guard the optional first element before reading its fields, and handle absent
contacts during form initialization. Initialize an empty contacts array only when
submitting an edit so newly entered phone/fax/email can be saved. Opening the
synthetic resources leaves their original input JSON unchanged. Populated address
fields and unrelated contact entries survive the tested phone edit. No backend,
database schema, migration, FHIR storage format or Wails contract changes.

Eleven additional real-component regressions cover absent/empty address arrays,
empty address objects, adding contacts to a resource without them, populated
address/contact preservation, and optional practitioner name/qualifications.
Before repair, the focused suite has six failures and seven passes, including
`.line`, `.find`, `.findIndex`, and `.displayName` errors. After repair all thirteen
focused assertions pass. Both focused command processes required task-scoped
termination after the existing local Chrome teardown hang; their exit is -15,
not a successful CLI exit.

The full browser suite then exits 0 with 415 passes and one existing skip, using
the retained debug-logging wrapper and a fresh private coverage directory. No
assertion, timeout or threshold changed. The launcher still logs teardown warnings;
existing Angular template diagnostics also remain. Five desktop bridge tests pass.
[Exact validation receipt](contact-dialog-validation.json). Protected Backend/Frontend/Desktop checks and packaged native build, ZIP/signature
verification and ten synthetic native checks passed before #347 merged.
[Native receipt](native-d19dca35a-validation.json).

This resolves the reproduced optional-address/contact edit crash. It does not
claim arbitrary malformed FHIR arrays are supported, complete practitioner
address parsing (#27), or preservation of additional addresses during an edit;
those broader behaviors were not changed. All fixtures are synthetic. Original
primary coverage recovery remains unresolved as recorded in the consolidation
ledger; these runs wrote only new report directories.
