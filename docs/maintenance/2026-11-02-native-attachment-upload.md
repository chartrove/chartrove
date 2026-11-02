# Native upload acceptance and attachment correction

After the Mac restart, the exact packaged app from source `6aca492ad` was
observed through the native UI bridge using an explicit disposable data directory.
The existing synthetic smoke account was used; production providers remained
unconfigured. The generated three-resource FHIR fixture opened through the actual
macOS chooser and imported successfully. Explore displayed Synthetic, Picker,
with one Patient, Encounter and Condition. This supplies the missing gesture
acceptance for #168, beyond the automated native smoke checks.

For #101, the encounter's paperclip opened the wizard, New Attachment opened,
and the native chooser selected a generated 78-byte plain-text file. Selecting
Consult Note and Document - Plain Text created an attachment in the wizard.
Save then returned HTTP 500 from `/api/secure/resource/related`, before file
extraction. Persistence was not accepted as successful. The stopped synthetic
profile, fixture hashes, app identity and log are retained in
`dist/validation/2026-11-01-consolidation/native-picker-manual-6aca492ad/`.
[Compact observation receipt](native-picker-6aca492ad-observations.json).

The related-resource API method still used FormData, which WebKit's native custom
scheme drops. The corrected request uses the explicit UTF-8 multipart string
already used by the working Patient import. The endpoint, `file` field,
`related.json` filename, JSON payload, response mapping, database and Wails
contracts remain unchanged. Embedded attachment bytes retain their base64 form.
This changes no existing record and runs no migration.

The first regression attempt had a fixture type error (missing required List
status/mode), retained separately. After correcting the fixture, the regression
failed for the intended object body and missing multipart header. The exact same
test passes after the request correction, including independent multipart
decoding of Unicode text, newline and attachment bytes. All eight focused API
tests and ten Node checks pass. The full ordinary browser suite passes 427 tests with one existing skip and
exit 0; existing framework/fixture warnings remain. Fresh native acceptance is pending;
#101 remains open until persistence is actually observed and the protected PR
passes and integrates. #168 will be individually reconciled with this evidence.

Chooser automation timed out twice; the exact process and existing chooser were
reobserved before further action. No blind selection was replayed. A synthetic
Patient without a birth date also displayed NaN Years Old; that separate defect
remains unresolved. No personal file or real health record was imported.
