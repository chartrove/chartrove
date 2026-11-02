# Do not log identity-provider token responses

The existing frontend IdpCallback logs its decoded token-response object. Remove
only that payload argument from the existing phase log; forwarding to the local
backend, returned/local-storage token and callback-state consumption stay intact.
No identity provider is configured, enabled or migrated by this change.

A regression exercises the real OAuth SDK callback/grant request and Response JSON
parsing, with synthetic provider/local-backend transport responses. It verifies
forwarding/storage and captures actual console arguments. Before repair those
contracts pass but both synthetic access/identity secrets appear in the log (one
failed test, exit 1). After repair both focused tests pass, exit 0. The full
ordinary browser command passes 423 tests with one existing skip, exit 0; five
bridge tests pass. [Exact validation](idp-log-validation.json).

Earlier fixture-load/response-metadata errors are retained separately and are not
counted as the product regression. The test explicitly provides HttpClient for the
legacy shared token and supplies the Response URL that fetch normally supplies to
the SDK. Its touched local-storage keys are restored after execution. No external
provider, real credentials, health database or family/purchased-app storage is used.
This is not production-provider or backend identity-validation acceptance.

The existing generic phase log remains. No authentication exchange, storage/data
schema, dependency, Wails or registration contract changes. The full macOS package build, ZIP CRC, strict extracted signature and all ten
native smoke checks pass for source 4756cf99e. Original output contents, stat
metadata and all xattrs restore exactly. [Native receipt](native-4756cf99e-validation.json).
Required Backend/Frontend/Desktop checks passed before protected PR #353 merged
as 4283ef5fd57c133c8f4d414c5dcddce4e96d3117.
Original-primary coverage recovery remains unresolved; these reports use new
private output directories.
