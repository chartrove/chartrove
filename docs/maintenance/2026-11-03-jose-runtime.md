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
Protected PR #385 merged as 6419a63e029d592542c14e341b1468d294af6baf after
all three required checks passed (run 41590962910). #305 is closed as implemented;
#246 is closed as superseded by v6. Both historical branches were already absent
when independently checked for retirement; their actor is unknown. Exact tips
remain archived. The integrated task head was deleted after ancestry, exact-tip
and open-PR dependency checks. [Retirement receipt](jose-retirement-1.json).

Fresh native packaging at 50cabfe5696e7275a3c6b83c71fe62c9292d1e71 passes
ZIP CRC, strict signature and all ten synthetic smoke checks. All 22,287 frontend
asset files match the qualified build; original output contents/stat/xattrs are
restored exactly with zero provenance changes. The actual app signs in to the
known synthetic profile, renders the dashboard and shows desktop-smoke with
Administrator in its profile menu. Sign Out returns to the empty sign-in form.
Both raw synthetic records and the seed database bytes remain unchanged. The
read-only SQLite verifier created an empty WAL and a 32 KiB SHM sidecar in the
synthetic seed at verification time; whole-directory unchanged wording was too
broad. Both sidecars and the original receipts are retained, with the correction
recorded separately. No real/purchased data was involved. Subsequent checks query
copies of stopped profiles and use immutable access for checkpointed seeds.
[Metadata correction](synthetic-seed-sidecar-correction.json). Tab completed
deferred initial/sign-in rendering; no source or asset workaround was applied.
The app is closed and ps confirms absence; no Cua call follows closure.
[Native receipt](native-50cabfe56-validation.json).

This batch also records #384/#252 integration and the individually retired #56
upstream FAQ follow-up. Stale pending markers for previously integrated
#101/#168/#137/#378/#379 are corrected using current issue states and existing
published evidence. No additional feature or migration is implied.

Raw evidence: dist/validation/2026-11-01-consolidation/jose-review/.
