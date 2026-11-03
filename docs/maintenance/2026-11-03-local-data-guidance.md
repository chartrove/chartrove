# Local data guidance reconciliation

#88 requests secure-default guidance instead of shared beta example credentials.
The runtime already rejects empty/historical signing keys before storage (commit
6ac39837606de0b6012b40db76f0f1e3efb5bc56); current startup instructions contain
no shared beta login. The new guide separates desktop/server defaults, required
keys and explicit synthetic-test options. config.yaml leaves the rejected shared
key empty for operator configuration. All other resolved settings match through
the actual Go config loader. Existing keys and records are not read or rotated.
Stale PostgreSQL/Angular/Node prose is corrected and frozen install examples match
the contributor command. This does not certify an arbitrary production deployment.

#193 preserves an attributed community Tailscale/Docktail recipe for an upstream
image. Proposed disposition is reference retirement, with discussion and custom-
certificate comments retained, rather than a new integration claim. The guide
explains the existing HTTPS toggle and desktop boundary; no proxy, trust-root,
container publication or custom-certificate API is implemented.

Four corrected observations pass in each graph. Actual startup rejects old/new
example keys before database setup; private environment keys are accepted by
configuration validation and explicit HTTPS false is honored. Negative controls
use an unsupported database type and explicitly disable encryption only inside
the harness to prevent standby. The positive control never starts a server.
Qualified outbound HTTP attempts are zero. Thirty-eight local links and four
single-H1 guides pass. Application/Go/frontend inputs remain at qualified Goja
source 4d1067a61; no additional native/UI check is claimed. Protected checks and
individual issue disposition remain pending.
[Validation](local-data-guidance-validation.json), [guide](../LOCAL_DATA_SAFETY.md).

The initial positive control called server Start with encryption enabled but no
recovery key. Standby bypassed its unsupported database type, and desktop probe
PID 49782 briefly listened on *:8080. The exact owned child was terminated with
SIGTERM; independent ps confirms absence. Its open-file snapshot had zero database
file matches. No family/purchased records were selected or supplied. Source
standby skips repository creation and hosted registration. Both initial runs are
failed diagnostics, excluded from qualification. Child logger buffers do not
reconstruct incoming request history; no zero-exposure claim is made. Socket,
file, process and source evidence remain, with an explicit warning against
rerunning the initial harness. A missing Node YAML parser preflight is also
retained; comparison uses the actual Go loader without installing a package.

Raw evidence: `dist/validation/2026-11-01-consolidation/local-safety-review/`.
