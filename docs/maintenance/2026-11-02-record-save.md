# Medical record save lifecycle - issue 35

The wizard already disabled Save while uploading, but a second form submission
could still send a duplicate write. Its initial encounter data was an empty object,
which passed required validation and then caused serialization to throw. It also
logged the complete clinical encounter to the browser console.

Guard pending submissions, start the required encounter control at null, disable
Save for an invalid form, and show an accessible saving indicator. Remove the
automatic encounter dump. Keep the API payload and backend/data contracts intact.

Four synthetic browser regressions cover initial invalid submission, duplicate
form events and saving status, failure/retry preserving unsaved form contents,
and absence of the clinical payload log. Three fail before the fix; all four pass
afterward. The full browser suite passes 394 tests with one existing skip and
process exit 0, keeping existing timeouts and assertions; production desktop_prod
build passes with existing 9.92 MB budget/CommonJS/browser warnings. No live
health records or external provider services are used. Required GitHub CI gates
integration. This does not claim every wizard validation or native file-picker
flow has been verified.
