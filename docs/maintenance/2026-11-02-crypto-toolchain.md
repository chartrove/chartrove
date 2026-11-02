# CryptoJS tooling and compiler compatibility

Historical #199 changes only the crypto-js transitive lock entry, 4.1.1 to 4.2.0.
The maintained graph still needed that update. Its sole installed package consumer
is @foliojs-fork/pdfkit 0.13.0 under pdfmake and Compodoc; frontend application
source has no crypto-js imports. The PDFKit consumer uses explicit MD5/RC4/AES/
SHA256 operations and WordArray generation; no PBKDF2 calls occur there. Backend
SQLCipher, persisted keys, Wails contracts and application record formats stay
unchanged. This is no database/key migration.

[CryptoJS's upstream advisory](https://github.com/brix/crypto-js/security/advisories/GHSA-xwcq-pm8m-c4vf)
identifies 4.2.0 as patched. Its changed PBKDF2 defaults were reviewed explicitly:
SHA256 and 250,000 iterations match an independently computed Node crypto vector.
This does not claim the application called vulnerable PBKDF2 or had an exploit.

Four fresh synthetic PDFs were generated with 4.1.1, then again with 4.2.0: plain,
PDF 1.4, PDF 1.6 and PDF 1.7 extension 3. Independent qpdf checks and pdftotext
decryption preserve the synthetic text in every case; all three new encrypted
variants also reject an incorrect synthetic password. These fixtures use no
existing documents or health records. Registry metadata/integrities are retained.
Frozen Yarn install, 404 browser tests (one existing skip) and desktop frontend
production build pass. Local Chrome teardown again required forced termination
after successful assertions; required GitHub checks remain mandatory.

The separate #313 proposal changes TypeScript to 7 without advancing Angular.
Installed @angular/compiler-cli 22.2.4 explicitly requires TypeScript >=6.0 <6.1,
consistent with [Angular's compatibility table](https://angular.dev/reference/versions).
Renovate now allows TypeScript updates only within that current peer range.
A focused check matches the bound exactly, accepts 6.0.3 and rejects 7.0.2.
This preserves the chosen toolchain and leaves compatible updates available;
a future compiler/framework migration must update the bound together.

The official Renovate JSON schema validates the config. The full CLI validator
was unavailable: npm's reported latest 44.200.0 tarball returned 404, and metadata
also requires Node 24 rather than the selected Node 26. No engine bypass, alternate
runtime installation, or successful full-validator claim is made. Existing
creation limits, dashboard approval, no automerge, and paused PR labels remain.

After this batch merges, #199 may close as an implemented replacement and #313
as an incompatible standalone proposal explicitly retired for the current compiler.
[Exact TypeScript tip and verified incremental bundle](typescript-pr313-recovery.json)
retain its unique work; all prerequisites are ancestors of retained main. Branch
retirement still needs a fresh open-PR dependency and exact-tip check.

Local evidence is retained under dist/validation/2026-11-01-consolidation/crypto-*
and renovate-*. Original-primary coverage recovery remains unresolved as documented
in the stale-session checkpoint. No claim of complete maintenance acceptance.
