# Timeline private note availability

The timeline's “Write a private note” field had no model binding, submission
handler or persistence path. Its unsupported state was only a tooltip, so it
invited entering text that could not be saved. The enclosing empty-action form
also offered ordinary browser submission without a save handler.

The timeline now visibly says that private notes are unavailable and provides
no editable field or form for them. The existing attachment/document entry
button retains its original wizard action and has an accessible label. This
does not implement note storage or change any existing clinical records.

Four frozen real-template browser regressions check the absence of the editable
field and form, visible availability text, and the preserved attachment action.
The original template produced three semantic failures and one passing control.
The form assertion checks rendered DOM, not a native keyboard or navigation
exercise. Full protected validation is recorded separately for the final source.

The broader notes and custom clinical-entry request remains open in
[Chartrove #194](https://github.com/karamhouse/chartrove/issues/194), imported
with yiohar's authorship and discussion from
[Fasten #637](https://github.com/fastenhealth/fasten-onprem/issues/637).
This correction does not claim to fulfill those requested capabilities.
Authentication, source authorization, persistence, encryption, provenance and
provider behavior are unchanged. No new notes schema, API or data migration is
introduced.
