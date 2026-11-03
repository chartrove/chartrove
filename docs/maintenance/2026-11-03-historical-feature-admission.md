# Historical feature source-admission review

These are attributed, unmerged upstream prototypes, not current-main defects or
implemented features. This is a targeted admission review of data/auth/service
boundaries, not an exhaustive security audit of every changed file. All six PRs
and their pinned comparison branches remain live pending individual disposition.
No prototype was executed against records or a production provider.

| PR | Preserved capability | Evidence and required corrected integration |
| --- | --- | --- |
| #209 / upstream #514 | User permissions management | Original description explicitly says permissions are not yet enforced. Adds permission rows/migration and admin editing routes; this is phase-one management, not complete authorization. Requires defined read/manage-sources semantics, enforcement across resource/source/query/export endpoints, denied-access fixtures and encrypted migration/reopen validation. Do not claim delegated permissions implemented. |
| #211 / upstream #594 | Typesense search and RAG | 149 files also bundle obsolete platform modernization. GetSettings returns the full search configuration; the browser client consumes search.api_key and directly lists/deletes conversations without a user filter. Search document-by-ID handler does not pass current-user identity to its indexer lookup. These require corrected credential/user isolation, lifecycle/delete propagation, encrypted-record boundaries and an explicit local/remote model policy before any health-data indexing. Do not revive the old platform stack or enable an external service. |
| #212 / upstream #609 | OCR capture and Encounter assistance | Adds an HTTP OCR sidecar using Tesseract/pdftoppm, a fixed ocr-service URL, unbounded HTTP clients, and shared /tmp/uploads filenames/page prefixes. Needs per-request private file isolation, limits/cancellation, local native packaging/camera permissions, safe optional terminology lookups, and reviewed clinical-data normalization with opt-in save. Preserve its capability; do not treat existing manual Encounter repairs as OCR implementation. |
| #214 / upstream #613 | OIDC sign-in | Login generates state but does not persist it; callback does not validate it. Callback links users by email (without a verified-email/issuer-subject linking policy) and sets every newly provisioned user to admin. App JWT is placed in redirect query parameters. Requires a corrected local callback/PKCE/state flow, durable identity binding, least-privilege provisioning, password/key recovery compatibility and an explicit scope decision; reject this patch as-is without discarding the capability. |
| #215 / upstream #614 | Delegated access | New routes are authenticated but do not call HasAccess for delegated resource reads/edits. Repository reads trust requested owner/source identity; UpdateResourceRaw filters only source/resource identifiers and writes raw JSON without checking delegate rights/current owner or refreshing derived FHIR indexes. Requires ownership/source-policy checks, revocation, authorization-denial fixtures and atomic raw/derived consistency before integration. The original proposal's external-source edit restriction is not established by the inspected write path. |
| #217 / upstream #636 | Reserved registration usernames | The full one-file patch rejects a fixed case-insensitive list, including admin/administrator; it has no tests. Existing roles are determined by actual role/first-user logic, not a username string. This is a registration/name policy, not a fix proving current role escalation. Requires an explicit Chartrove naming policy, creation-only behavior and legacy-account compatibility tests. Do not silently ban legitimate local maintainer names or infer that the feature is implemented. |

On November 3 the maintainer explicitly selected **all six** capabilities above
for Chartrove's implementation scope. This resolves the earlier unspecified
subset, but does not approve the unsafe prototype implementations or settle
credential, indexing, provisioning, naming and migration policies. All six remain
open for corrected integration and capability-specific validation. This review
does not blanket retire them or convert unimplemented capabilities into closed
issues. Generic/type-only modernization overlaps must be separated from feature
work (for example, Chart.js widget type corrections are now integrated by #392).

Recovery: remaining-features-heads-and-bases-20261103.bundle is 1,281,491 bytes,
SHA-256 4d277cdab4ed2944f4e828d60f3cbb5204ae531178beca3344feb85e39bed8f0.
Strict index-pack and original-pack-offset verification cover all 2,586 unique
objects. Six feature heads are advertised; all six exact base OIDs are recorded
in heads-and-bases-recovery.json and are ancestors of protected preserved main.
OIDC/delegation bases are not ancestors of their feature heads. The first archive
admission expected all twelve refs to be advertised; Git excludes base refs under
^main because those commits are already reachable. The corrected proof explicitly
checks main reachability rather than assuming a head contains its pinned base.
All temporary archive refs were removed after verification; no live historical
head/base was retired. The original intake bundle remains intact as well.

Raw original descriptions, complete fetched patch payloads and recovery proofs
remain in the corresponding remaining-feature-review evidence directory.
