# Explicit source permissions

The maintainer selected explicit per-source VIEW/EDIT delegation on November 4,
2026. Owners retain their existing own-record APIs. Administrator status manages
accounts; it does not confer medical-record access. There are no account-wide
medical permissions or implicit grants.

Historical references: imported #209 / upstream fastenhealth/fasten-onprem#514
by dwradcliffe (`cf4e493599dd0f7ce4716169ad713a2f360fc751`) and #215 / upstream
#614 by TheoAtTechStack (`e7c9632e54e04ff2701f1cd36a51d39dd7388bb4`). Their
permission-management and VIEW/EDIT capabilities inform this corrected design.
The old account-wide model and unchecked delegated handlers are not reused.
Original discussions, exact refs and the existing recovery bundle are preserved.

## Repository boundary

`SourceAccessRepository` is separate from the existing owner-only repository
operations. It resolves the authenticated actor itself. A source ID locates data;
it does not prove authorization. Each clinical SELECT/UPDATE repeats both the
source's actual owner and the actor's currently active grant in SQL. Unknown
levels, revoked grants, expired grants and deleted sources deny access. Owners
alone can grant, change, revoke and inspect grants/audit events for their source.
Delegates cannot grant onward access or manage source credentials, sync or users.

Shared-source responses use an allowlisted DTO, never `SourceCredential` or
`User`. List/detail operations do not hydrate associations from other sources.
Source-specific summaries count authorized rows only; they do not merge patients
across owners. Existing account summaries, IPS, queries and graphs stay own-only.

EDIT additionally requires server-recorded local provenance. Platform type alone
is insufficient: even a Fasten source can contain uploaded bundles. The explicit
owner-only local creation operation assigns a fresh resource ID in the owner's
Fasten source and records provenance transactionally. It never promotes an
existing record. Historical records, imported bundles and provider records remain
read-only to delegates. SDK ingestion removes provenance when replacing a record,
including when an owner edits via an older ingestion-based API. Broader manual
form integration requires using this explicit creation boundary; no backfill
guesses which historical records were local.

Delegated updates preserve the database identity/owner/source, reject changed
FHIR identity, rebuild derived FHIR fields and clear removed values. They update
exactly one existing authorized row, never upsert. No source sync, association
editing, deletion or credential mutation is delegated. Grant/change/revoke and
local creation/edit append identifier-only audit events in the same transaction;
failure to record an event rolls back the mutation. Audit does not store raw FHIR,
tokens, passwords or encryption keys. Revocation is checked afresh on subsequent
statements/requests; already delivered data cannot be recalled.

## Validation and delivery

The synthetic encrypted SQLite regression exercises three users, an administrator
without a grant, two sources, malformed historical cross-owner rows, VIEW/EDIT,
onward-grant denial, changed identity, revocation, expiration, import replacement,
derived search fields, audit failure rollback and encrypted reopen. The additive
migration has a frozen schema and does not rewrite clinical tables or credentials.
Offset-bearing expirations are normalized to UTC without changing caller values.
An orchestrated second SQLite writer reproduces and guards against ingestion's
read-to-write snapshot contention. Failed imports retain provenance, while repeated
association imports remain idempotent inside the replacement transaction.
Delegated updates perform the fully authorized UPDATE as their first transaction
statement, preserving identity/presentation columns without a preliminary read.
Revocation likewise writes its scoped grant row before reading audit metadata;
an unrelated concurrent writer cannot invalidate a prior read snapshot.
Local creation changes only the top-level ID using raw JSON fields, preserving
clinical numeric literals. Ambiguous identity aliases and duplicate top-level
keys are rejected. Shared details also project the stored JSON as text so clients
can preserve decimal precision when displaying and editing it.

Delivery was split into repository foundation #413 and authenticated API/UI #414,
each accepted through the protected workflow. Historical #209/#215 were closed
after that corrected integration. Later OIDC, browser OCR and retrieval milestones
retain this source-level authorization model. No production data operation is
implied by source acceptance. Both independent retention worktrees remain held.

## Authenticated API and UI

Repository foundation accepted through #413 at
`f98caab9aeefd8fb1d12fd7b1ae2a9f207eb6910`; all seven required checks passed at
head `b7d60685ef16236b85794ff5fc6c984492d14f39`. #414 subsequently delivered the
following API/UI through its own protected acceptance on that main baseline.

Settings links to **Source permissions** (`/source-permissions`). Owners select
one source and an existing username, choose VIEW or EDIT, optionally set an
expiration, and can explicitly revoke the grant. The owner sees retained grants
and the latest 200 audit events. Shared sources have a separate browser with
source-specific counts, paginated resources and details. Only server-confirmed
editable local records expose the advanced JSON editor. Pending read requests
are cancelled on selection changes; a denial clears the displayed record/editor.
These controls supplement, rather than substitute for, repository authorization.

All new routes are under the existing authenticated `/api/secure/source-access`:

| Method/path | Capability |
| --- | --- |
| GET `/shared` | Active source grants for the actor; metadata only |
| GET/POST `/:sourceId/grants` | Owner-only list or explicit grant/change/reactivation |
| DELETE `/:sourceId/grants/:grantId` | Owner-only revocation, retaining history |
| GET `/:sourceId/audit` | Owner-only latest 200 identifier-only events |
| GET `/:sourceId/summary` | Authorized per-source resource counts |
| GET `/:sourceId/resources/:resourceType?page=0` | Authorized list, 50 per page |
| GET `/:sourceId/resources/:resourceType/:resourceId` | Authorized detail and edit eligibility |
| POST `/:sourceId/resources` | Owner-only explicit creation of a new locally managed record in their Fasten source |
| PATCH `/:sourceId/resources/:resourceType/:resourceId` | EDIT plus local provenance; fixed identity, atomic raw/derived/audit update |

Request bodies reject unknown envelope fields and trailing JSON, and are limited
to 1 MiB. Grants accept `delegate_username`, `access_level` and optional RFC3339
`expires_at`; clinical writes accept `resource_raw`. Source/owner identities are
resolved in the repository, never supplied through request-body ownership fields.
Foreign, missing or unauthorized records all return 404 without revealing their
existence. This also leaves valid sign-in sessions intact on permission denial.
Invalid input returns 400; internal errors do not expose database details.
Clinical details include the exact stored `resource_raw_text`. The browser displays
and submits that text, using JSON parsing only for validation, so editing never
round-trips clinical numbers through JavaScript's numeric representation.

Focused HTTP tests cover owner/delegate/admin separation, successful reads and
edits, denied writes/onward grants, revocation, changed identity, malformed
envelopes and body limits. Browser regressions cover read-only rendering without
HTML execution, stale-response cancellation, denied-write clearing, source-scoped
grant/revoke requests, and owner credential projection. URL construction uses the
existing endpoint helper for both web and desktop environments.

## Remaining feature dependencies

Historical/manual import records are deliberately not backfilled as local. Existing
manual forms and graph/association editing can adopt the explicit provenance and
source authorization boundaries in separately reviewed changes. They must not
derive permissions from platform labels or widen account summaries. The new local
creation API/UI already supports new eligible records without those migrations.

OIDC #415 binds authentication to native identities; browser OCR #418 preserves
import provenance; #416/#420 add source-authorized metadata/reviewed-transcript
search. Native session identity was hardened separately in #423. Full-text/RAG
remains unresolved in #419, and operational Authentik/recovery acceptance remains
approval-gated in #425. These capabilities preserve current owner/source grants,
revocation and the separation between administrative and medical access.

## Source-scoped metadata search

The source browser can find records by their existing title or FHIR ID within
one selected source and resource type. It queries the existing encrypted database;
there is no external index, model, new credential or external clinical-record store.
This narrow retrieval step follows the useful pagination and source/type/resource
identity concepts in groundmuffin's Fasten #594 / Chartrove #211 proposal. The
historical branch, follow-on authorship and discussion remain preserved; its
Typesense/browser-key and conversational paths are not integrated.

`POST /api/secure/source-access/:sourceId/search/:resourceType` accepts only
`{"text":"title or ID","page":0}`. Text stays out of the request URL and SQL
trace parameters; query structure, timing and row counts remain available for
diagnostics. Responses use `Cache-Control: no-store`. The body is
limited to 4 KiB, text to 128 Unicode characters, and pages to 50 hits with a five-second
database deadline and stable
resource-ID order. Percent signs, underscores and the escape character match
literally. Case matching uses the database's native lower-case rules. Only
allowlisted FHIR table names enter SQL; text and permission values are parameters.

Each query repeats the current native actor, actual source owner, active grant,
expiration and deletion predicates in the clinical-table statement. Administrative
roles do not grant access. Hits contain only source/type/record IDs, existing title
and date. For DocumentReference, #420 also searches the decoded reviewed-transcript
projection from new local document imports. Arbitrary raw clinical JSON and binary
attachments are not searched or returned in hits. Some records have
no derived title and can be found by ID. Opening a hit uses the existing detail
API, which checks permissions again; a hit never carries an enduring permission.
Search adds no write permission and changes none of the existing list, detail,
summary or mutation contracts. Source changes and denials clear search results and
record/editor state. The original source counts remain unfiltered.

The common shared-resource predicate compares canonical owner values as text on
both sides, supporting PostgreSQL's UUID source owners and legacy text FHIR owner
columns without rewriting clinical tables. This preserves the same exact owner
and grant checks for browsing, searching and editing, including native UUID FHIR
owner columns. SQLite isolation remains unchanged.

Focused regressions cover encrypted SQLite multi-user isolation, administrator
denial, VIEW/EDIT reads, expiration, revocation before detail access, metadata-only
responses, literal matching, Unicode, pagination and deleted records. A disposable
PostgreSQL regression exercises the same source authorization and literal-match
semantics. Browser tests cover request bodies, source scoping, pagination and stale
response cancellation. Typesense, whole-record full-text retrieval and RAG remain
separate unresolved features.

## Inline document text

Read-only DocumentReference details can preview a single inline `text/plain`
attachment as UTF-8, including reviewed transcripts from local document imports.
The attachment title and text render literally with line breaks, without HTML
execution or automatic URL fetching. The exact stored FHIR text remains available
under **FHIR record as JSON**. This display does not certify review/provenance,
interpret clinical statements, or change eligibility for editing.

Preview work is bounded to eight content entries, one plain-text attachment,
128 KiB decoded bytes and the corresponding base64 input bound. Unsupported
charsets, multiple text attachments, missing inline data, oversized or malformed
encoding fall back to the original FHIR record; no partial transcript is silently
selected or truncated. No additional API, model, destination, dependency or
persistent browser storage is introduced. The existing authenticated detail
request and live source permissions remain authoritative. Preview state clears
with source/type/page/detail changes, denied access and component teardown; late
responses remain cancelled by the existing source/detail lifecycle.

Focused browser coverage includes UTF-8/literal markup, byte limits, remote and
malformed attachments, raw-record preservation, denied-write behavior and stale
response/permission clearing. Backend/Desktop and all security checks remain
required through the protected workflow. No live clinical acceptance is inferred
from these synthetic fixtures.
