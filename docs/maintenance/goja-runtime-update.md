# Goja runtime dependency qualification

Root and desktop select the same maintained original Goja revision,
`v0.0.0-20261107161106-d6eaa50db8bc`, replacing the October 31 revision.
Chartrove uses it to compile FHIRPath/search programs and extract searchable
fields from imported records. The runtime pool and generated models are unchanged.

The [official comparison](https://github.com/dop251/goja/compare/cf2ff0ec0b3165c13e7f06c1499f006920f86451...d6eaa50db8bc0687ca68b9b4d34a0c65eb3de2a5)
contains runtime correctness changes and Float16 support; this is not described
as a single internal bug fix. Both upstream module requirement files are
byte-identical, including the Go 1.26.0 requirement. Chartrove keeps its Go 1.26.1
minimum and all other selected pins/replacements. The module archive was obtained
with the native Go tool and checksum verification, bound to the exact upstream
commit. This is not publisher-signature or exhaustive upstream-audit attestation.

The added synthetic consumer regression preserves composed/decomposed accents,
CJK and Arabic family names through actual generated Patient search extraction.
It is compatibility coverage, not a newly reproduced application defect or an
upstream conformance suite. Existing resource-isolation, invalid-JSON and FHIR
extractor regressions remain intact. Root/backend and desktop checks must qualify
the changed runtime; matching module metadata alone is not runtime acceptance.

SQLCipher replacements, FHIR libraries, authentication, source authorization,
schema and provider configuration are unchanged. No dependency tidy, production
record migration, provider/account operation, image publication or release is
part of this update. Older checksum entries are retained; selected versions are
established by the two module requirements, not by the mere presence of a sum.
