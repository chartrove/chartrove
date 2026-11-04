# Reserved registration usernames (#217)

Historical [PR #217](https://github.com/karamhouse/chartrove/pull/217), originally
[Fasten #636](https://github.com/fastenhealth/fasten-onprem/pull/636) by
Vangalla, Rohith, is classified **implement and integrate with a corrected
Chartrove policy**. Its exact commit `bba5c43f954346ef170c2f6abbf57cc09abb59aa`
and original discussion remain preserved. The existing six-feature bundle still
verifies and matches SHA-256
`4d277cdab4ed2944f4e828d60f3cbb5204ae531178beca3344feb85e39bed8f0`.
No prototype history is rewritten or discarded.

## Policy and behavior

New account creation reserves exact names api, contact, fasten, chartrove, help,
info, login, mail, noreply, postmaster, root, security, support, system, webmaster
and www. Matching trims surrounding whitespace and ignores case, without changing
accepted usernames. `admin` and `administrator` remain available for legitimate
local maintainers, retaining the compatibility regression established by #406.
This narrower policy was selected autonomously after presenting an optional
policy question; it is not represented as a maintainer response.

The shared predicate guards both the HTTP registration boundary and repository
creation. Rejection returns a clear HTTP 400 before account lookup, hashing or
creation. Direct repository callers reject before hashing or any database write,
so no user, default settings or manual-source credential can be left behind.
Role assignment still follows the existing first-account policy. A username does
not confer administrator privileges.

The predicate is absent from sign-in and existing-account reads. Existing reserved
names, IDs, passwords, roles and source ownership survive encrypted cold reopen.
There is no renaming, backfill, schema migration, key rotation or patient operation.
This integrates the naming feature; it makes no claim of fixing a demonstrated
privilege escalation. The upstream authorship and substantive discussion remain
on the historical PR and in the existing recovery material.

## Validation and acceptance

Corrected baseline controls have 20 leaf failures and 19 passes: all 17 direct
reserved-name creations and three HTTP cases fail on old source. Root and desktop
graphs then each pass 49 focused cases, plus the updated seven-case sign-in suite
which explicitly permits a legacy reserved name and checks token subject/role.
Actual repository fixtures verify denied-creation side effects, accepted literal
names, password checks, encrypted headers and legacy cold reopening. Synthetic
private temporary databases are the only data used.

An initial test scaffold referenced the wrong settings model and failed to build;
that log is retained and excluded from behavior proof. The corrected baseline and
final logs and source hashes are in [the receipt](reserved-usernames-validation.json).
Protected CI remains mandatory and pending at this source checkpoint. The existing
50 GiB reserve prevents local full builds; no new native package is claimed.

The independent issue batch also dispositioned #99 as a superseded facet-layout
report: neither current `/sources` route renders the old filter component.
#160's provider-search feature and #167's separate navigation report remain open.
#194's private notes are not implemented by the existing unbound timeline input;
its custom report/template requests also remain unresolved. No closure is claimed.

Protected acceptance: #411 merged at
`fa2044a5127e6f15ef2d628a955260fd543a4300` after all seven checks passed for
`c777a1bd0f71a2c937f1ed9edb4fd5c3b3f3ad70` in run41810661747. Historical #217
is individually closed as integrated through the documented policy variant.
The original admin-name ban is not misrepresented as adopted. The integrated
local task branch is retired; expected-tip atomic remote retirement is recorded
separately. Both independent retention owners remain held.
