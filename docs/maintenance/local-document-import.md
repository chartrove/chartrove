# Local document extraction and reviewed import

Settings → Import a scanned document opens a browser workspace for PNG, JPEG and
PDF files. English recognition runs locally. A PDF's embedded text is used when
available; otherwise each page is rasterized and recognized. The user reviews and
corrects the transcript, selects an owned local source and explicitly confirms
before saving. The original file can be downloaded for comparison before saving.
Neither selecting nor recognizing a document uploads it.

This is a focused replacement for the capture, review and original-document
provenance in [Chartrove #212](https://github.com/karamhouse/chartrove/pull/212),
imported from [Fasten #609](https://github.com/fastenhealth/fasten-onprem/pull/609)
by [TheoAtTechStack](https://github.com/TheoAtTechStack). The historical head
`658dc9751320427cd2d6da18e0b50e5d3fc00c75`, comparison base, commits and discussion
remain preserved. No historical implementation is merged directly. Its shared
temporary-file OCR service, published service port and external terminology calls
are incompatible with the accepted local-only boundary. Automatic inferred
Encounter/Condition/Medication generation is not included: reviewed text is a
document transcript, not a validated clinical assertion.

## Processing and storage boundary

- [Tesseract.js](https://github.com/naptha/tesseract.js) 7.0.0, its locked WebAssembly
  core and the packaged English language data are served from Chartrove's own
  `assets/ocr/` directory. Worker, core and language paths are explicit, including
  deployments under a base path. Language caching is disabled. No CDN, separate
  OCR server, terminology service, analytics endpoint or model API is used.
- [PDF.js](https://github.com/mozilla/pdf.js) 6.5.438 decodes a local byte array.
  Its worker, font, character-map, color-profile and decoder assets are also local.
  This workspace invokes text extraction and page rendering only; it does not
  instantiate the scripting/annotation viewer or navigate PDF links. XFA is disabled.
- One document processes at a time per tab, with a 60-second wall deadline and
  explicit cancellation. The worker lifecycle channel carries only a random job
  identifier and ready/cancel/timeout messages, never documents or credentials.
  Worker termination, PDF destruction, canvas clearing and blob-URL revocation
  release processing state on completion, cancellation or component destruction.
  Current browsers with workers, WebAssembly and BroadcastChannel are required.
- Limits are 8 MiB per original, 10 PDF pages, 128 KiB of reviewed UTF-8 text and
  256 title characters. Image dimensions are checked before browser decoding:
  at most 24 megapixels on input, downscaled to at most 8 megapixels for recognition.
  PDF raster pages are limited to 8 megapixels. These limits bound accepted work;
  browser decoder memory and recognition accuracy depend on the document/device.
  English only is included in this milestone. Password-protected, malformed,
  unreadable or over-limit documents fail without creating records.
- `POST /api/secure/document-import/:sourceId` accepts exactly four multipart
  fields: `file`, `text`, `title`, `reviewed=true`. It streams into bounded memory
  with two concurrent upload slots and a 30-second read deadline. No multipart
  temporary files or standalone plaintext originals are written by this path.
  Reverse proxies must allow the bounded request size and avoid body logging.
  The server checks actual byte signatures for PNG/JPEG/PDF and stores originals
  opaquely; it never runs an OCR/parser process on server-side document contents.

Saving creates one FHIR DocumentReference containing the exact original bytes and
the explicitly reviewed transcript as separate base64 attachments. Its description
is the user-selected title; its date is the import time, not an inferred clinical
event date. A SHA-256 identifier fingerprints the original, and `meta.source`
records `urn:chartrove:document-import:user-reviewed-local-extraction` provenance.
The transcript is not a diagnosis, verified medical terminology, or a substitute
for the original. Editing an extraction before saving never changes the original.

The record uses the existing database and existing encryption configuration:
encrypted SQLite remains encrypted; PostgreSQL retains its existing deployment
storage controls. No encryption key, source credential or unrelated record is
returned. Import SQL logs retain structure/timing with parameter values omitted.
The record and payload-free `import_document` source audit event commit atomically.
A server-derived identity includes the owner/source, original and transcript
fingerprints, title and filename. Retrying the same submission returns the existing
active import without duplicating the record or audit. Conflicting/deleted identities
fail rather than resurrecting a record. Changing reviewed content creates a distinct
import; it does not overwrite history.

New imports also populate the existing DocumentReference text search projection
with the reviewed transcript as a JSON string array. Source permissions → browse
the source → DocumentReference searches title/ID or that decoded transcript using
literal substrings. Search returns only bounded navigation metadata, never the
transcript or attachments; opening a hit rechecks current source permission. Quotes,
backslashes, Unicode and SQL wildcard characters remain literal input. The query
uses the database's existing case-folding behavior (SQLite's built-in LOWER is
primarily ASCII), a five-second deadline and the existing 50-result pagination.
It never scans the original binary data or submits a query to a model/service.
Only this import path's provenance-marked projection is included; historical
documents without that projection remain searchable by title/ID. No existing
clinical rows are silently backfilled, regenerated or rewritten.

## Authorization and compatibility

The server obtains the native account from existing authentication and repeats its
active user/source ownership predicates in the import transaction. Only an owned,
active `fasten` source is eligible. The source picker returns only ID/display, and
neither administrators nor delegates can import into someone else's source. OIDC
authentication uses the same native UUID and grants no extra access. Revoked or
expired OIDC sessions fail the existing middleware before upload parsing.

Imported documents receive no locally authored record marker. Source VIEW/EDIT
delegates may read only through an active explicit source grant, and both levels
remain read-only for this imported content. Revocation/expiry continues to apply to
metadata, search, detail and mutation requests. There is no account-wide sharing.
Existing owner API contracts remain unchanged. Source browsing now uses a new
bounded metadata-only endpoint, so paging does not download a page of attachments;
opening a record rechecks permission before returning its contents.

The new workspace is web-only. Desktop/cloud modes do not advertise it and retain
existing local authentication, encryption, imports and native APIs. This adds no
desktop binding, database migration, infrastructure service, provider credentials
or live configuration. A future desktop decoder/camera flow or additional language
pack needs its own compatibility and resource acceptance.

## Validation and rollback

Synthetic browser regressions exercise the real bundled PNG engine, embedded PDF
text and scanned-PDF recognition; review/no-auto-upload, cancellation, account
changes and original-byte preservation; and existing source access behavior.
Encrypted SQLite and real authenticated HTTP regressions cover ownership, admin
isolation, duplicate/oversized fields, missing review, exact provenance, retry,
atomic audit rollback, read-only delegation and revocation. A separate disposable
PostgreSQL query/transaction fixture covers import, retry, metadata, delegated read,
write denial and revocation; it is not full application-migration acceptance.
Protected Backend, Frontend, Desktop and CodeQL checks remain required on the PR.
No real patient data is used by these tests.

Rollback can revert the replacement commit and restart/rebuild Chartrove. It adds
no tables or migration: already imported DocumentReferences remain ordinary FHIR
records with their original attachments and audit history. Preserve them and the
existing database/key backups; do not drop data or rewrite grants. Remove the
workspace/routes through source rollback if desired. No external service or live
identity-provider change needs undoing.
