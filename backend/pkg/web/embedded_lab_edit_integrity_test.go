package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedLabEditTextPreservesFieldsProvenanceAndAuthorization(t *testing.T) {
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-lab-edit.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-lab-edit-key"))
	request := func(method, path, token, body string, status int) json.RawMessage {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		require.Equal(t, status, w.Code)
		var response struct{ Data json.RawMessage }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		return response.Data
	}
	token := func(raw json.RawMessage) string {
		var tkn string
		require.NoError(t, json.Unmarshal(raw, &tkn))
		return tkn
	}
	owner := token(request("POST", "/api/auth/signup", "", `{"username":"synthetic-lab-owner","password":"synthetic-password","join_mailing_list":false}`, 200))
	delegate := token(request("POST", "/api/auth/signup", "", `{"username":"synthetic-lab-delegate","password":"synthetic-password","join_mailing_list":false}`, 200))
	request("POST", "/api/secure/users", owner, `{"username":"synthetic-lab-admin","password":"synthetic-password","role":"admin"}`, 200)
	admin := token(request("POST", "/api/auth/signin", "", `{"username":"synthetic-lab-admin","password":"synthetic-password"}`, 200))
	r := app.engine.deviceRepo.(*database.GormRepository)
	u, err := r.GetUserByUsername(context.Background(), "synthetic-lab-owner")
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username)
	ctx = context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, u.ID)
	sources, err := r.GetSources(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, sources)
	source := sources[0].ID
	original := json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic result","coding":[{"system":"http://loinc.org","code":"synthetic-code"}]},"subject":{"reference":"Patient/synthetic"},"effectiveDateTime":"2020-01-02T03:04:05+05:30","note":[{"text":"Retain café صفر"}],"extension":[{"url":"urn:synthetic:exact","valueDecimal":0.1234567890123456789}],"valueQuantity":{"value":2.000,"unit":"mg/dL","system":"http://unitsofmeasure.org","code":"mg/dL","extension":[{"url":"urn:synthetic:quantity","valueString":"keep"}]}}`)
	local, err := r.CreateLocalResource(ctx, source, original)
	require.NoError(t, err)
	otherSource := &models.SourceCredential{Patient: "synthetic-another-lab-source"}
	require.NoError(t, r.CreateSource(ctx, otherSource))
	otherRaw := json.RawMessage(fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"final","code":{"text":"Synthetic other source"},"valueQuantity":{"value":1}}`, local.SourceResourceID))
	_, err = r.UpsertRawResource(ctx, otherSource, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: local.SourceResourceID, ResourceRaw: otherRaw})
	require.NoError(t, err)
	base := "/api/secure/source-access/" + source.String()
	var grant struct{ ID string }
	require.NoError(t, json.Unmarshal(request("POST", base+"/grants", owner, `{"delegate_username":"synthetic-lab-delegate","access_level":"EDIT"}`, 200), &grant))
	detail := base + "/resources/Observation/" + local.SourceResourceID
	var before models.SharedResource
	require.NoError(t, json.Unmarshal(request("GET", detail, owner, "", 200), &before))
	require.Contains(t, before.ResourceRawText, `"value":2.000`)
	// Same exact-text transport emitted by the client: only the numeric value
	// changes; no generated widget ID, source credentials or ownership fields.
	edited := strings.Replace(before.ResourceRawText, `"value":2.000`, `"value":3.5`, 1)
	body := fmt.Sprintf(`{"source_id":%q,"resource_raw":%s}`, source, edited)
	legacy := "/api/secure/resource/fhir/Observation/" + local.SourceResourceID
	request("PATCH", legacy, "", body, 401)
	request("PATCH", legacy, delegate, body, 404) // EDIT grant does not authorize owner-only transport.
	request("PATCH", legacy, admin, body, 404)    // Admin status is not medical access.
	request("PATCH", legacy, owner, body, 200)
	var after models.SharedResource
	require.NoError(t, json.Unmarshal(request("GET", detail, delegate, "", 200), &after))
	require.True(t, after.Editable, "owner value edit must retain explicit local-edit provenance")
	require.Equal(t, before.ID, after.ID)
	require.Equal(t, before.SourceID, after.SourceID)
	require.Equal(t, before.SourceUri, after.SourceUri)
	require.Equal(t, before.SortDate, after.SortDate)
	require.Contains(t, after.ResourceRawText, "0.1234567890123456789")
	var oldFields, newFields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(before.ResourceRaw, &oldFields))
	require.NoError(t, json.Unmarshal(after.ResourceRaw, &newFields))
	for field, want := range oldFields {
		if field != "valueQuantity" {
			require.JSONEq(t, string(want), string(newFields[field]), field)
		}
	}
	var oldQuantity, newQuantity map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(oldFields["valueQuantity"], &oldQuantity))
	require.NoError(t, json.Unmarshal(newFields["valueQuantity"], &newQuantity))
	for field, want := range oldQuantity {
		if field != "value" {
			require.JSONEq(t, string(want), string(newQuantity[field]), field)
		}
	}
	require.Equal(t, "3.5", string(newQuantity["value"]))
	other, err := r.GetSharedResource(ctx, otherSource.ID, "Observation", local.SourceResourceID)
	require.NoError(t, err)
	require.JSONEq(t, string(otherRaw), string(other.ResourceRaw))
	require.False(t, other.Editable, "SDK resource must not become locally editable")
	request("PATCH", detail, delegate, `{"resource_raw":`+edited+`}`, 200) // Existing eligible delegated route remains valid.
	request("DELETE", base+"/grants/"+grant.ID, owner, "", 200)
	request("PATCH", detail, delegate, `{"resource_raw":`+edited+`}`, 404)
	request("GET", detail, delegate, "", 404)
	request("GET", detail, admin, "", 404)
}
