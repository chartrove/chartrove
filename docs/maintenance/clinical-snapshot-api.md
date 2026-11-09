# Owner-confirmed clinical snapshot API

Pi owns `task/clinical-snapshot-api`, stacked on #472. This slice adds authenticated storage/reading routes,
**not** native clinical retrieval, a new provider profile or an enabled import UI. The current direct SMART
card remains Patient-only. Production and household end-to-end acceptance remain open.

## Receiving identity and raw fidelity

`POST /api/secure/source/clinical-snapshot` runs under the existing real JWT/session middleware. Its JSON
object requires exactly these case-sensitive keys, once each:

- `expected_owner_id`: UUID of the current receiving local identity, available from authenticated
  `/api/secure/account/me`. A mismatched/stale receiver is forbidden, not silently replaced.
- `confirmed`: literal `true`, asserting the actor's explicit import decision. This is not proof of a
  provider consent grant, legal authority for a relative or a human having clicked a particular UI.
- `origin`: exactly `fhir_base_url` and `display`. Claimed provenance is revalidated by the store, never
  treated as native/provider authorization. No credentials belong here.
- `snapshot_json`: **string** containing the exact raw collection JSON. Do not parse then stringify a
  numeric FHIR object in JavaScript to construct this value. String quoting preserves raw decimal
  lexemes, Unicode, extensions and whitespace through the HTTP boundary.

Input is an owner-supplied collection, not a permission to fetch an arbitrary FHIR URL. No provider or
conversion service is contacted by these endpoints. Unknown/duplicate/case-aliased envelope fields,
invalid UTF-8, extra documents, wrong types, missing confirmation and over-limit collections fail with
fixed payload-free errors. Nested origin keys are also strict. The shared collection validator and
atomic [store contract](clinical-snapshot-persistence.md) remain authoritative for identities/resources.

One expensive import is admitted per route registration. The actual-byte reader caps wire input at
32 MiB plus 8 KiB; the decoded original still has a 16 MiB / 5,000-Observation bound. A 30-second body
read deadline is requested where the transport supports it. Import has the existing two-minute store
budget; no multipart upload or plaintext staging file is created. Busy callers receive a retryable 503.
This is not a full application-wide rate-limit or transport slow-client qualification.

Successful responses contain the scoped source/snapshot IDs, original digest, Patient ID and write
counts. Exact historical acknowledgement is not "already current". No pre-commit success is sent or
published: notification follows a committed new snapshot. Delivery uses the optional nonblocking
`BestEffortPublisher` capability: the production bus either admits the event immediately or drops it with
an unavailable error. A busy dispatcher/stalled SSE client, unsupported legacy publisher or notification
error produces a fixed warning while retaining HTTP success. There is no synchronous-publication fallback,
per-import goroutine, waiting timer or accumulated retry queue. Notification is a best-effort refresh hint,
not a durable delivery receipt; consumers must also load current state. Replay emits no new-commit event.

## Owner-only original history

- `GET /api/secure/source/:sourceId/clinical-snapshots?limit=50&offset=0` returns payload-free metadata.
  Existing database bounds are 1–200 rows and offset 0–1,000,000. Invalid values fail closed.
- `GET /api/secure/source/:sourceId/clinical-snapshots/:snapshotId/original` verifies ownership and original
  digest, then returns exact bytes as `application/fhir+json`, `Content-Disposition: attachment` with a
  UUID-only filename and `X-Content-Type-Options: nosniff`. Two downloads are admitted per registration.

Both routes are `Cache-Control: no-store`. Administrators, foreign identities and EDIT delegates cannot
read another owner's history. Current source sharing does not imply raw-history rights. Database and
decoder error values are not echoed or logged by these handlers. No new read-auditing or retention policy
is implied. Backups remain sensitive encrypted evidence with separate key/configuration custody.

## Synthetic acceptance

Tests use real signed JWT middleware and the real encrypted repository, not manually injected identity
claims. Principals are inserted as disposable UUID-bound fixtures; password/OIDC authentication and live
Authentik operation are not claimed.

- Authorized import, exact original download and payload-free metadata; absent JWT refusal.
- Foreign/admin/delegate denial and stale-recipient refusal; successful exact replay without a new event.
- Updated values retain full original decimal/extension bytes across the string envelope.
- A notification callback queries the committed original; a failed notifier cannot convert a completed
  import to an error or echo its private canary.
- A final audit failure rolls source/originals back with no notification, fixed response and no private logs;
  unchanged valid prerequisites succeed after removing the injected failure.
- Duplicate/unknown/case-aliased keys, malformed/extra JSON, invalid UTF-8, credentials in origin, absent
  confirmation and oversized decoded/wire inputs fail. Actual wire limits are checked without trusting
  Content-Length. A blocked importer proves another expensive projection is not started.
- Encrypted backup of two HTTP-imported versions, boot a separate copy, real-JWT owner download of both
  exact originals, foreign download refusal and an unchanged immutable backup artifact.
- Review correction: a real production bus with a deliberately unconsumed, still-connected listener
  reproduced a committed response blocked beyond two seconds. That unchanged HTTP test now passes while
  the listener remains stalled; a second changed import also succeeds (not 503), with both exact originals
  readable. The test releases/joins its deliberate stall before failure; it does not leave a blocked slot.
  One initial fixture compile error (original getter has two return values) was corrected before reproducing
  the actual failure. A legacy-only publisher control proves the unsafe synchronous fallback is never called.
  Bus controls cover admitted metadata serialization, full/absent dispatcher refusal and no-op support.

Following the review correction, fresh focused HTTP race tests passed (12.949s), complete event-bus race
passed (1.195s), complete web and handler suites passed (62.819s / 52.351s), and stalled/legacy controls under
the desktop dependency graph passed (1.386s). Changed-package vet/diff passed. Original API fixtures and
implementation were developed together; only the stalled-publisher regression is a demonstrated pre-fix
failure, not the whole feature suite. Final-head
hosted validation and source review remain required. This is neither independent peer certification nor
an operator/key-loss, PostgreSQL or cross-platform recovery qualification.
