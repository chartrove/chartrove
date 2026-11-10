# Encryption setup response cache boundary

## Ownership and draft scope

Pi owns #506 in the sole checkout on `task/encryption-response-cache`, from protected
#505 merge `e0e6bbb289c591efffbfbf91317030f16e98714a`, separately main-qualified by
Validation42747314980. Validated two-file Go product checkpoint:
`6f89d613232cf2121d20930a2a32e1bf42816b31`; protected matching-head/exact-main acceptance
remain pending. #505 accepted a failed-control receipt, not a #479/#504 maximum-race fix.

An actual registered STANDBY `AppEngine.Setup` GET `/api/encryption-key` diagnostic on
clean cd9c59ea (Go source identical to accepted e0e) returned200 with no response
`Cache-Control: no-store`, exit1. The real handler generated a random encryption key;
its body was never printed/persisted/exported. No network/socket, repository, existing
database, CLI key/setup POST or real record was used. This proves missing HTTP cache
prohibition, not an observed browser/proxy cache disclosure or credential theft.

## Concrete change

Set `Cache-Control: no-store` at entry to the existing encryption-key Generate, Set
and Validate handlers, including error returns. No global cache setting or header
middleware. Random generation, key acceptance, response bodies/statuses, route
exposure, unlock/validation/reinitialize behavior and authorization are unchanged.

This instructs compliant HTTP caches not to store these responses. It does not
retroactively clear cached material, revoke/rotate keys, wipe memory, undo clipboard
copies, guarantee browser/OS storage behavior or provide TLS, Origin/CSRF/Host safety,
key-custody, independent recovery or production-readiness acceptance.

## Synthetic controls and retained compatibility reds

- Tracked actual STANDBY routes genuinely failed2.184s on ten cache-header cases:
  root/reverse-prefix GET success, Set invalid/missing and Validate invalid/missing.
  Status controls still matched. Invalid POST cases never opened a database or
  triggered setup/reinitialize; generated bodies are not dumped by assertions.
- Three header lines yielded initial focused race8.038s. A separate real SQLCipher
  fixture then checked correct/wrong validation, wrong/correct synchronous unlock,
  no-store, unchanged locked state during validation/wrong-key rejection, and key
  generation route absence after unlock. No passwords, accounts, provider or network.
  Isolated new cache/lifecycle race54.992s passed, two source hashes unchanged.
- Expanded encryption/unlock race530.567s failed unchanged
  `TestEmbeddedEncryptionUnlockIsSynchronousAndRejectsWrongKey` at
  `embedded_test.go:112`: bootstrap with closed public registration500, test330.73s.
  A hash-verified accepted-e0e handler-source overlay ran that exact legacy test once:
  race600.215s/test595.00s, same bootstrap500. This compatibility red is reproduced
  with pre-header source; no full legacy-unlock race acceptance or host/cost/timing
  cause is established. Original15-second/cost14 registration and15-minute package
  budgets remain. No further candidate legacy race retry or host experiment.
- Initial expanded ordinary60.859s also failed legacy bootstrap500/test30.57s.
  Accepted-e0e ordinary control35.393s passed; then one unchanged current ordinary
  comparison27.360s passed the complete selected encryption/unlock group. Retain
  all three results, not a rewritten green first invocation. No test deadline,
  fixture, registration policy, cryptography, scheduler/priority/foreground/capture
  setting or assertion changed. Changed web/handler vet and diff checks pass.

Complete serialized handler/web ordinary passed566.432/280.445s, actual Go-test0/
two source hashes0/unchanged15-minute package budgets. Go/configuration stayed frozen
through capture; documentation separate. Owned wrapper34083 observed absent. All six
native packages compiled/race-built with **no tests run**; actual cmd2.198/smoke1.313s
race controls passed. These are bounded compile/smoke scopes, not full local native-core,
packaged runtime/GUI or older-OS acceptance; macOS27.2-object/11.0-link-target warnings
retained. Protected matching head/exact-main, live browser/native GUI caching and
independently operated recovery remain unclaimed. The full legacy race red above stays distinct from focused cache
controls, ordinary results and hosted checks. No historical cache/log cleanup or
broader token/credential/privacy certificate follows from this directive.
