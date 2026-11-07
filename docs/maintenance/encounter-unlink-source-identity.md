# Encounter unlink source identity

FHIR business IDs can repeat in separate imports. The legacy encounter-unlink
route accepts resource type and ID without a source selector. Both endpoints now
use the existing native-owner resolver before looking up or removing associations.
It rejects multiple live owned matches, ignores deleted sources and other owners,
and does not treat administrator roles or source grants as ownership. Ambiguity
returns HTTP 409 with a fixed message; no association mutation is attempted.
Unambiguous requests keep their existing direct/shared-neighbor behavior.

A synthetic encrypted embedded regression on main `7c4562f97f4b` imported two
sources with repeated Encounter/Observation IDs. The original route returned HTTP
200 instead of rejecting ambiguity. All six associations and raw records stayed
unchanged in that execution: it did not demonstrate cross-source deletion. The
regression preserves those stronger checks alongside the conflict requirement.
An earlier local compiler failure caused by a missing explicit SDK path is
retained separately and supplies no application-defect evidence.

This correction adds no schema, source-selector API, frontend matching heuristic,
delegated unlink permission, credential operation or historical data repair.
Selection and association removal retain their existing separate transaction
boundaries; concurrent imports/deletions are not newly claimed atomic. The UI
cannot yet disambiguate this legacy route. Operational deployment, provider
connectivity and #425 recovery/authentication acceptance remain separate.
