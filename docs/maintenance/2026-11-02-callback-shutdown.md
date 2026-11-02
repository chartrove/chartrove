# SMART callback listener shutdown

PR #334 CI failed TestCancelAndCloseReleasePendingListener/true with
"callback port leaked". Repeating the existing cancellation/close test 200 times
locally reproduces failures on unchanged main. The listener is bound before
HTTP Serve runs in another goroutine. If cancellation wins before Serve registers
it with http.Server, Server.Close can return while the bound socket is still open.

Retain ownership of the net.Listener and close it synchronously before
Server.Close. Preserve the existing grace period for an accepted callback response,
state/PKCE validation, cancellation, token isolation and all test timeouts.
The existing cancellation test then passes 200 repetitions. The full SMART suite
passes under the race detector, all desktop host tests pass, and SMART vet passes.
No test or status-check requirement is weakened. Raw before/after and CI failure
logs remain in dist/validation/2026-11-01-consolidation.

This repairs an observed callback lifecycle race; it does not introduce a provider
registration or prove production EHR interoperability. Required CI gates merging.
