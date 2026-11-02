# Server signing-key guard - issue 52

The standalone server previously used the historical shared signing key unless
an operator replaced it. Reject empty, whitespace-only and historical default
values before any database initialization or network startup. Existing private
values and the desktop's generated per-installation key remain unchanged. No key
rotation, database migration, default storage change or production-data inspection.
This guard does not assess operator-selected key entropy or detect copied keys.

Three synthetic startup regressions fail before the fix with the unrelated
unsupported-database error, showing startup reached database initialization.
They pass with the guard. The environment-override regression passes and reaches
an explicitly unsupported synthetic database type; no database can open in any
of these tests. Full root web/config tests pass; changed-package vet and backend
build pass. Desktop-graph startup and encrypted embedded-backend tests pass in
17.597 seconds within the unchanged 120-second limit; desktop host tests pass.

Remove the CLI debug dump of all configuration (which can contain credentials)
and AuthService signup-response console logging (which contains the token).
These are direct logging removals, with no API or token-format changes.
Remaining frontend dependency advisories are not resolved by this patch.
Required GitHub CI remains mandatory before integration.
