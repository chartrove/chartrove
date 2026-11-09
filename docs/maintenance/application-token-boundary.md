# Application HTTP authority boundary

## Scope and ownership

Pi owns #485, branch `task/application-token-boundary`, in the canonical single checkout.
This is a bounded interceptor/registered-provider correction, not a global credential audit or live
Lighthouse/provider compatibility certification. No real credentials, records, external endpoint,
production identity, deployment, other repository or dependency refresh was used. Routine trusted
project synthetic tests are not an OS-sandbox/security certificate; the prior restrictive macOS
sandbox limitation remains.

## Source boundary and corrections

The accepted interceptor inserted the local application JWT into optional external Lighthouse
requests and interpreted their 401/403 as local session invalidation. It also overwrote an explicit
external Authorization header and treated API-prefixed sibling paths as application API membership.
A local signed session is not a remote service credential. This does not establish that a real service
received, validated, abused or retained any installation token.

Only the configured `fasten_api_endpoint_base` authority and path namespace now receive automatic
application JWT insertion and local authentication-error handling. Scheme/host (including port)
match explicitly, userinfo is rejected and the normalized pathname matches the API root or a slash
boundary underneath it. Reverse-proxy prefixes remain supported. Explicitly configured remote cloud
application APIs remain application authorities; unrelated local or remote origins do not become one.

External requests retain their original method, body, headers/explicit provider credentials and
error. The interceptor does not even read a native JWT for them. Optional Lighthouse calls remain
present but no longer receive unrelated local session authority. If a remote deployment requires
that credential forwarding, it needs a separately specified scoped integration contract; no replacement
credential or live availability is invented. Existing local protected-API 401/403 handling and the
`registration_closed` signup exception remain. The legacy local auth-error-to-next-message behavior
is unchanged, not newly certified as a uniform transport error contract.

Native custom-scheme URLs serialize `origin` as `null`, which is neither a usable URL base nor proof
of matching authority. Resolve against actual browser protocol/host, and compare protocol/host/path
instead. `APPLICATION_HTTP_LOCATION` defaults to the real browser location and permits explicit
synthetic-location provider tests without global patching. AppModule's existing explicit interceptor
`deps` list now includes it. Same-serialized-origin foreign hosts/protocols never gain a token.

The inspected legacy C-CDA path already has an explicit external-sharing warning/confirmation;
its separate converter origin was not receiving an automatically added token in these fixtures.
It is not mislabeled as automatic/unconsented conversion, nor qualified as offline conversion.

## Evidence

- Accepted-source classifier genuinely failed **5/13** focused tests: Lighthouse JWT insertion,
  explicit external credential substitution, remote 401/403 logout/error swallowing and API-prefixed
  sibling classification. Other-origin/converter and local positive controls passed.
- Adding only the injectable location seam while retaining old classification exposed **four more**
  invalid-base native URL cases: **9/18 failures**. This synthetic native location models URL
  semantics, not an observed rendered clinical window. Correcting the base must not introduce
  null-origin equality as a new trust decision.
- Initial correction passed 25 focused interceptor/registered-provider tests. Stronger coupled
  custom-URI provider tests caught an omission in the **new draft seam**: the existing explicit DI
  dependency list still passed two arguments, ignoring the override. Two absolute-native assertions
  failed in 41 tests while relative positives passed. This is not mislabeled as a pre-existing
  production DI failure. Updating that list made all **41** unchanged assertions pass.
- Controls cover both registered HttpClient entry points, external 401/403 propagation and session
  preservation, explicit provider authorization, local credentials/invalidation/signup exception,
  relative/absolute native URLs, other custom hosts/protocols, reverse-proxy roots, sibling and dot
  segments, foreign authority and userinfo.

Feature `da4e0a7b44` was normally integrated with accepted native #484 at clean checkpoint
`b9167cb50bf7ed5a3d18e922ff50f431783237aa`. Complete existing frontend validation passed all
Node checks, **772 browser passes/one skip**, desktop production build and Storybook. Its retained
receipt records four zero exits and scratch removal. Existing icon/form/builder/Browserslist
warnings are not suppressed. Complete serialized native races also passed: command 1.533s, host
22.688s, encrypted composed shell 26.864s, SMART 6.317s, fixture 1.253s, smoke 1.197s. Diff checks
passed; backend production source is unchanged from the accepted parents.

A clean full b916 package was built, extracted and strictly ad-hoc-signature verified, preserving the
prior exact-68e ZIP with byte/hash checks. The actual packaged backup binary passed a separate-process
synthetic encrypted WAL restore (3.75s test / 4.166s package). Normal installed Cua in standard mode
launched the exact extracted app, with all windows hidden, into a fresh explicitly unencrypted smoke
directory. Independent readback found all ten native checks true in a 339-byte mode-0600 receipt;
the named driver lifecycle ended. This custom smoke page does **not** exercise the changed normal
Angular interceptor; its evidence stays separate from the coupled synthetic-URI Angular tests.

The current ZIP is an exact-b916 inspection candidate, not a later receipt-only source commit,
protected-main/current release, notarized build, older-OS or independently operated recovery result.
Private logs, test data, keys, archives, source/hash receipts and the earlier negative-only native
calibration remain private. Native completion does not independently attest process exit. Actual
visual/interactive receiving-profile/clinical workflows still require legitimate capture consent.
