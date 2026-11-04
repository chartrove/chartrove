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

Delivery is split into a repository foundation and authenticated API/UI exposure,
each through the existing protected workflow. #209/#215 remain open until the
corrected integration is accepted. OIDC, OCR and Typesense/RAG are separate future
batches; neither their implementation nor any production data operation is part
of this authorization work. Both independent retention worktrees remain held.
