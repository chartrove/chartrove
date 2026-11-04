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
Local creation changes only the top-level ID using raw JSON fields, preserving
clinical numeric literals. Ambiguous identity aliases and duplicate top-level
keys are rejected. Shared details also project the stored JSON as text so clients
can preserve decimal precision when displaying and editing it.

Delivery is split into a repository foundation and authenticated API/UI exposure,
each through the existing protected workflow. #209/#215 remain open until the
corrected integration is accepted. OIDC, OCR and Typesense/RAG are separate future
batches; neither their implementation nor any production data operation is part
of this authorization work. Both independent retention worktrees remain held.

## Authenticated API and UI

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

OIDC must bind the authenticated actor to a durable local identity without role
or medical-access escalation. OCR must preserve import provenance. Search/RAG must
enforce current owner/source grants and revocation without exposing service keys
or stale shared results. None of those features is enabled by this integration.
