# Confirmed encounter unlink display

The encounter wizard must keep its displayed links until the unlink API succeeds.
An HTTP failure, including the ambiguous-identity conflict from PR #455, leaves
every displayed row and raw clinical field intact. A failed request reports that
the user must reload before retrying: a transport failure does not prove that the
server made no change.

The selected display model may be a copy. Match its source UUID, resource type
and business ID to exactly one displayed row, rather than using object identity
or a business ID alone. Capture that row, its list, and its encounter before
opening confirmation. Recheck the captured selection before sending and before
applying a response. A replaced encounter/list/row or changed identity must not
cause a different row to disappear. Permit only one unlink request at a time;
declining or dismissing confirmation sends none.

## Regression evidence

The unchanged seven-case synthetic regression was published first at
`abe826aab2312378be7e0ba9d2cdb30ddaace083`, based on accepted main
`3928db93805bde8cdfe81ccd99fa9b3e8de169b0`. The actual hosted Frontend job
[128382651531](https://github.com/karamhouse/chartrove/actions/runs/42365522581/job/128382651531)
executed 683 of 684 browser tests with five failures and one pre-existing skip.
All five failures are assertions in the new unlink regression: premature or
incorrect row removal, row loss on HTTP rejection, stale confirmation, and
duplicate requests. Compilation and fixture setup succeeded. This demonstrates
incorrect UI state, not a wrong-source database deletion.

The seven cases cover pending/successful requests with a copied model, HTTP 409
with duplicate business IDs across sources, an absent source tuple, a changed
encounter before confirmation, a changed encounter before response, competing
confirmations, and declined confirmation. The implementation must pass those
same assertions without changing their source.

## Retained boundaries

Server authorization and the legacy unlink API are unchanged. The UI does not
gain a source selector and cannot resolve server-side business-ID ambiguity;
that conflict remains an error. Selection and removal are not a new atomic
database transaction. Concurrent imports/deletions and a network failure still
require reconciliation by reloading. No migration or record repair is performed.

This source correction is separate from packaged desktop Epic clinical import,
clinical repeat-update acceptance, and Authentik operational acceptance (#425).
