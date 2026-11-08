# Bounded native clinical snapshot reader

This is an internal source foundation for the packaged Apple Silicon desktop
clinical-import milestone in [the existing SMART tracking record](../../desktop/DIRECT_SMART.md).
It does not register a host bridge, HTTP import route, UI action or provider
profile. The existing public SMART operation remains Patient-only. Clinical
persistence, owner-authorized import and repeat-update behavior remain unfinished;
this reader is not provider or operational acceptance.

## Retrieval contract

The manager retains granted scopes privately alongside its existing in-memory
access token. Existing token-scope validation remains unchanged, including the
OAuth rule that an omitted response scope denotes the requested scope. A clinical
read requires `patient/Observation.read` (SMART v1) or `patient/Observation.rs`
(v2). V2 `.r` alone permits read, not search. Requested-but-withheld permission
cannot enable the reader. See the official
[SMART scope semantics](https://hl7.org/fhir/smart-app-launch/STU2.2/scopes-and-launch-context.html).

The reader fetches the token-bound Patient using the existing method, then starts
an Observation search with that exact Patient ID. Continuation query parameters
remain opaque, as specified by [FHIR paging](https://hl7.org/fhir/R4/http.html#paging).
Bearer-bearing continuations must remain HTTPS on the pinned origin and exact
Observation path. Redirects remain disabled. Other continuation shapes fail
without output, rather than forwarding credentials or guessing a compatible URL.

A read has a two-minute deadline capped by token expiry, and follows at most 64
pages. Each page is at most 2 MiB. Total response-resource bytes and the final
collection are bounded by 16 MiB, with at most 5,000 Observations. Cancellation,
disconnection, expiry, duplicate resources, malformed pages, cycles, transport
failure or any exceeded bound produces one constant error and no partial result.
No tokens, opaque paging state or provider error bodies enter returned snapshots.
There is no token refresh, retry loop or new background producer.

## Identity and provenance

The shared, dependency-free parser accepts one Patient and at least one
Observation. Every Observation must reference that exact Patient using its
relative reference or the pinned FHIR base. It validates literal FHIR IDs,
unique type/ID pairs, canonical identity field spelling and duplicate JSON keys
at identity/framing boundaries. It preserves raw resource JSON, number precision,
partial dates, Unicode, resource order and unrelated clinical fields.

Missing full URLs derive from the pinned base, type and literal ID. Reported full
URLs must equal that URI or be a syntactically valid UUID URN; duplicate provenance
URIs fail. No normalization merges distinct IDs or rewrites clinical content.
OperationOutcome warnings, included resources and unsupported resource types
fail closed. This bounded subset can reject otherwise valid FHIR server responses;
it does not promise a complete clinical chart or a universal FHIR client.

This parser establishes payload consistency, not internal account authorization.
The subsequent persistence layer must independently enforce native UUID ownership,
active source identity, revocation/deletion, stable resource identity and atomic
repeat updates. No existing source credentials, encryption, schema, delegation,
local-edit permissions, Patient import or source-association behavior changes here.

## Validation and rollback

The test-first hosted run at `e1d5c8f7ed11f622e91d40ef7fe86329e5ca608c`
compiled both modules. The four shared-parser positive cases failed against the
closed scaffold; the native positive/page/cancellation tests failed because the
reader was unavailable. Denial controls remained closed. These are missing-feature
proofs, not demonstrations of a pre-existing clinical-data vulnerability.

Frozen synthetic tests exercise exact values and Patient identity, opaque paging,
actual scope grants, cancellation and foreign-origin rejection. Additional tests
exercise page/byte/count bounds, malformed/duplicate envelopes, expired or removed
connections and no partial success. Required hosted validation and independent
source review apply to the final head; original failures remain retained evidence.
No real records, provider connections or local heavy builds are used for this step.

The foundation is not exposed to callers outside the internal manager. Reverting
its source and tests restores the prior Patient-only surface with no data or
configuration migration. Packaged native clinical import, persistence and repeat
updates require subsequent source work and separately approved live acceptance.
