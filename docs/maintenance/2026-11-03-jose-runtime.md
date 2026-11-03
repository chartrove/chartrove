# Browser JWT decoding dependency

Replace the current direct jose 4.11.2 with 6.2.15 under ^6.0.0, implementing
#305 and superseding the older v4 patch #246. The application uses only decodeJwt
at two AuthService call sites. No application method, Go signature verification,
database, Wails contract, stored record or provider integration changes.

Thirteen real AuthService cases pass on both installed versions with identical
test bytes. They cover legacy Go claims, access-token metadata, optional claims,
UTF-8 text, future/expired/exact-boundary/missing expiry, absent/malformed tokens,
the existing case-sensitive admin role and logout/state publication. Time is
fixed and local storage is restored. The synthetic tokens exercise decoding,
not signature validation; backend authentication remains responsible for trust.
The initial baseline compile diagnostic concerned expected-value typing for
claims broader than the inherited UI class. It is excluded; qualified exact
wire-value assertions use unknown without changing the model or runtime.

Canonical frozen installation, ten existing Node checks, the full 513-test
browser suite (one existing skip), desktop production and Storybook pass.
The other 72 direct versions and unrelated manifest/lock bytes are unchanged.
Coverage and generated-documentation absence remain intact.
[Validation](jose-validation.json).

The npm archive passes its registry SHA-1/SHA-512. All 80 installed package files
match the verified archive. Of these, 79 runtime/declaration/license/README files
match official tag v6.2.15 exactly. package.json removes only the tag's imports,
scripts and devDependencies; every remaining field is unchanged. The original
strict all-file comparison failed on that metadata difference, and its comparison
and exact diff remain preserved alongside the qualified result. The full upstream
cryptographic suite was not run; the validated application boundary is decoding.

Both historical heads are preserved in verified checksum-recorded bundles, with
all four unique objects for #246 and five for #305 proved present in archive
bytes. Prerequisites remain reachable from main; earlier intake tips are linked.
[v4 recovery](jose4-head-recovery.json), [v6 recovery](jose6-head-recovery.json).
Both proposals and branches stay live until protected integration and fresh
dependency checks. Native packaging and UI acceptance remain pending.

This batch also records #384/#252 integration and the individually retired #56
upstream FAQ follow-up. Stale pending markers for previously integrated
#101/#168/#137/#378/#379 are corrected using current issue states and existing
published evidence. No additional feature or migration is implied.

Raw evidence: dist/validation/2026-11-01-consolidation/jose-review/.
