# Encrypted SQLite backup and isolated recovery

This operator-only tool creates a **full-instance encrypted database backup**.
It is not a user/admin HTTP capability, an IPS export, an automatic restore,
or authorization to open existing family records. The #425 operational exercise
remains separately approved. Tests use only disposable synthetic databases.

## Consistency and scope

Chartrove uses WAL. A synthetic actual-constructor test demonstrated that copying
only the live main file omitted a committed row. The existing pinned driver exposes
the [SQLite online-backup API](https://www.sqlite.org/backup.html); the command uses
that API, including committed WAL state, rather than copying database/WAL files.
SQLite obtains a consistent committed state; the exact transaction cutoff can
advance while incremental copying runs. Receipt completion time is not a guaranteed
point-in-time start cutoff. Uncommitted transactions are excluded.

All database tables are copied: clinical raw data and indexes, inline FHIR
attachments/reviewed transcripts, native users/password hashes/UUIDs, source
credentials, explicit grants, local provenance, OIDC bindings/session revocations
and audit history. The artifact is therefore highly sensitive even without the key.
This is not a patient-scoped export. Remote attachment URLs remain references;
the command does not fetch remote objects or capture external files/services.

It does **not** include the encryption recovery key, JWT signing key, IdP client
configuration/secrets, TLS material, external attachments or installation settings.
Preserve those through separately approved custody. A database backup alone is not
a complete installation recovery kit. The original Fasten storage format remains
unchanged: SQLCipher legacy 3, HMAC off, KDF 4000, page size 1024, through the existing
tested driver. This change does not rekey, migrate or upgrade encryption.

## Explicit command

Build the command from the accepted revision with the existing Go/CGO toolchain:

```sh
go build -o /absolute/private/bin/fasten-backup ./backend/cmd/fasten-backup
```

An operator must approve the exact source, fresh output directory, key-file custody
and artifact before any persistent-data use. For the approved synthetic exercise:

```sh
/absolute/private/bin/fasten-backup \
  --source "$APPROVED_SYNTHETIC_DATABASE" \
  --output "$NEW_APPROVED_BACKUP_DIRECTORY" \
  --key-file "$APPROVED_PRIVATE_KEY_FILE" \
  --timeout 2m
```

All paths must be absolute. The output must not exist and its immediate parent
must be a private directory (0700). Source/key final components must be regular
files, not symlinks. The key file must be private (0600/0400), contain 1–4096 exact
UTF-8 key bytes and contain no NUL. Whitespace is significant; the command does not
strip a newline. Supply the existing key exactly, without adding a newline. Never
put a key in command arguments, command substitution, a report or a Git file.
It loads neither `config.yaml` nor `FASTEN_*` variables and has no default data path.

The source connection is read-only and runs no migrations/jobs. Normal SQLite WAL
read coordination still applies; this is not an immutable/raw-filesystem read.
The destination uses the same key/format, private directory/file permissions,
incremental copying, a maximum five-minute timeout, cancellation, and bounded lock
waits. It finalizes a standalone file, closes/reopens it for integrity verification,
hashes and syncs it, and writes/syncs a receipt. Existing outputs are never replaced.
Failures attempt removal of only this call's newly created known files, not a
recursive cleanup. A kill/power loss can leave partial output: retain for inspection
and use a new output path, never mistake it for a completed backup.

Success requires exit zero plus `database.sqlite` and `receipt.json` in the new
directory. The JSON stdout/receipt contain format, UTC completion, byte length and
SHA-256; no key, source path, account or clinical contents. Verify the ciphertext
hash against an independently retained trusted receipt after transfer. SHA-256
detects accidental change; it is not a signature/authentication or a forensic
erasure guarantee. Files are 0600 and the output directory is 0700. Do not include
sidecars or copy plaintext intermediates.

## Isolated restore contract

1. Check the source revision/tool and accepted application artifact, receipt format,
   ciphertext byte length and SHA-256. Preserve the original backup and receipt.
2. With the destination stopped, copy only the verified `database.sqlite` into a
   **new approved disposable private recovery directory**. Never replace an open
   database or combine it with WAL/SHM files from another installation/snapshot.
3. Supply the independently held database key. Use the same accepted application
   revision and explicitly isolated SQLite configuration/profile. Wrong-key unlock
   must fail without making an account or enabling protected data access. Startup
   normally runs migrations: do not use an unreviewed older/newer binary on recovery.
4. Without Authentik, sign in with the independently held native password. Verify
   stable native UUIDs, raw clinical bytes/dates/decimals, attachment hashes, source
   provenance, active VIEW/eligible-local EDIT, admin/unrelated denial, expiration,
   revocation, deletion and audit history. Restart and repeat the access checks.
5. Review the snapshot cutoff before enabling any service/provider/OIDC path.
   Deletions/revocations committed **before** backup are preserved. A historic
   snapshot cannot know later revocations/deletions and can restore earlier policy;
   reconcile those explicitly while isolated. Old backups and deleted free pages
   are not purged by this command. Define RPO/RTO and retention with the operator.

The native regression uses a deliberately new installation JWT key, so copied old
JWTs fail and fresh native login succeeds; restored grants remain revocable across
restart. Keeping an existing signing key has its existing bounded-session contract.
No password reset, account migration, production restore or key rotation is added.

## Verified source boundary and PostgreSQL

The [recovery regression](../../backend/pkg/database/sqlite_backup_test.go) compares
every application table after deleting its original synthetic installation,
checks inline attachment bytes, exact clinical numbers/dates and authorization,
then restarts twice. [Boundary tests](../../backend/pkg/database/sqlite_backup_boundary_test.go)
cover uncommitted writes, lock deadlines/cancellation, private files and rejection
of plaintext, bad keys, symlinks, missing input, existing output and non-SQLite input.
[CLI tests](../../backend/cmd/fasten-backup/main_test.go) use the actual command
entry point; [embedded recovery](../../backend/pkg/web/embedded_backup_recovery_test.go)
checks native HTTP login/unlock and live family permission enforcement.
Protected Backend/Frontend/Desktop/CodeQL checks remain required. These source tests
do not certify an installed artifact, storage power-loss behavior, off-host retention,
custody or deployed restore. Reuse the existing #425 approval packet and matrix.

**PostgreSQL recovery is UNVERIFIED and separate.** This command never connects to
PostgreSQL and must not be used on its files or substituted for its recovery tools.
An operator-approved PostgreSQL contract must pin server/tool versions, choose
logical dump/restore versus physical base backup plus WAL/PITR, protect transport
and backup storage/roles/secrets, and prove restoration into a new isolated server
with the same identities/attachments/grants/provenance/deletion checks. Do not run
an existing database constructor as a backup operation; it performs migrations.

Rollback removes the new command/capability through protected review; it changes
no schema or existing data. Keep verified backups and the current key/custody plan.
Never replace this mechanism with an uncoordinated copy of the live main file.
