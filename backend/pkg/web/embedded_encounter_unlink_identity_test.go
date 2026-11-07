package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedEncounterUnlinkRejectsAmbiguousSourceIdentity(t *testing.T) {
	blocked := &blockedTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv("TMPDIR", t.TempDir())
	cfg := embeddedConfig(t, filepath.Join(t.TempDir(), "synthetic-unlink.db"), true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	cfg.Set("auth.registration.enabled", true)
	cfg.Set("jwt.session.duration", "1h")
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase("synthetic-unlink-identity-key"))
	repo := app.engine.deviceRepo.(*database.GormRepository)
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w
	}
	login := func(path, name string) string {
		t.Helper()
		w := request("POST", path, "", `{"username":"`+name+`","password":"synthetic-password","join_mailing_list":false}`)
		require.Equal(t, http.StatusOK, w.Code)
		var result struct{ Data string }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.NotEmpty(t, result.Data)
		return result.Data
	}
	owner := login("/api/auth/signup", "synthetic-unlink-owner")
	login("/api/auth/signup", "synthetic-unlink-other-admin")
	// This changes only the disposable fixture account, then issues a fresh login.
	require.NoError(t, repo.GormClient.Model(&models.User{}).Where("username = ?", "synthetic-unlink-other-admin").Update("role", pkg.UserRoleAdmin).Error)
	otherAdmin := login("/api/auth/signin", "synthetic-unlink-other-admin")
	var sources []string
	for i, patient := range []string{"synthetic-unlink-patient-A", "synthetic-unlink-patient-B"} {
		bundle := fmt.Sprintf(`{"resourceType":"Bundle","type":"collection","entry":[{"resource":{"resourceType":"Patient","id":%q,"name":[{"family":"Synthetic","given":["Unlink"]}],"gender":"unknown"}},{"resource":{"resourceType":"Encounter","id":"synthetic-shared-encounter","status":"finished","class":{"system":"http://terminology.hl7.org/CodeSystem/v3-ActCode","code":"AMB"},"period":{"start":"2020-01-02T12:00:00Z"},"subject":{"reference":"Patient/%s"}}},{"resource":{"resourceType":"Observation","id":"synthetic-shared-observation","status":"final","code":{"text":"Synthetic unlink fixture"},"subject":{"reference":"Patient/%s"},"encounter":{"reference":"Encounter/synthetic-shared-encounter"},"valueInteger":%d,"note":[{"text":"Retain café صفر"}]}}]}`, patient, patient, patient, i)
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		file, err := form.CreateFormFile("file", "synthetic-unlink-bundle.json")
		require.NoError(t, err)
		_, err = io.WriteString(file, bundle)
		require.NoError(t, err)
		require.NoError(t, form.Close())
		r := httptest.NewRequest("POST", "/api/secure/source/manual", &body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("Authorization", "Bearer "+owner)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
		var result struct{ Source struct{ ID string } }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.NotEmpty(t, result.Source.ID)
		sources = append(sources, result.Source.ID)
	}
	require.NotEqual(t, sources[0], sources[1])
	associations := func() []models.RelatedResource {
		t.Helper()
		var rows []models.RelatedResource
		require.NoError(t, repo.GormClient.Order("resource_base_user_id, resource_base_source_id, resource_base_source_resource_type, resource_base_source_resource_id, related_resource_user_id, related_resource_source_id, related_resource_source_resource_type, related_resource_source_resource_id").Find(&rows).Error)
		return rows
	}
	rawRecords := func() map[string]string {
		t.Helper()
		rows := map[string]string{}
		for _, source := range sources {
			for _, kind := range []string{"Encounter", "Observation"} {
				id := "synthetic-shared-encounter"
				if kind == "Observation" {
					id = "synthetic-shared-observation"
				}
				w := request("GET", "/api/secure/source-access/"+source+"/resources/"+kind+"/"+id, owner, "")
				require.Equal(t, http.StatusOK, w.Code)
				var result struct{ Data models.SharedResource }
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
				require.Equal(t, source, result.Data.SourceID.String())
				require.NotEmpty(t, result.Data.ResourceRaw)
				rows[source+"/"+kind] = string(result.Data.ResourceRaw)
			}
		}
		return rows
	}
	before := associations()
	for _, source := range sources {
		found := false
		for _, row := range before {
			direct := row.ResourceBaseSourceResourceType == "Encounter" && row.ResourceBaseSourceResourceID == "synthetic-shared-encounter" &&
				row.RelatedResourceSourceResourceType == "Observation" && row.RelatedResourceSourceResourceID == "synthetic-shared-observation"
			reverse := row.ResourceBaseSourceResourceType == "Observation" && row.ResourceBaseSourceResourceID == "synthetic-shared-observation" &&
				row.RelatedResourceSourceResourceType == "Encounter" && row.RelatedResourceSourceResourceID == "synthetic-shared-encounter"
			if row.ResourceBaseSourceID.String() == source && row.RelatedResourceSourceID.String() == source && (direct || reverse) {
				found = true
			}
		}
		require.True(t, found, "fixture must contain the actual encounter/result link in each source")
	}
	beforeRaw := rawRecords()
	const path = "/api/secure/encounter/synthetic-shared-encounter/related/Observation/synthetic-shared-observation"
	require.Equal(t, http.StatusUnauthorized, request("DELETE", path, "", "").Code)
	require.GreaterOrEqual(t, request("DELETE", path, otherAdmin, "").Code, 400, "administrative roles do not grant medical access")
	require.Equal(t, before, associations(), "denied actors must not alter source relationships")
	w := request("DELETE", path, owner, "")
	after := associations()
	t.Logf("synthetic ambiguous unlink: HTTP=%d associations before=%d after=%d", w.Code, len(before), len(after))
	if !reflect.DeepEqual(before, after) {
		t.Error("source-ambiguous unlink changed associations instead of preserving every source")
	}
	if w.Code != http.StatusConflict {
		t.Errorf("source-ambiguous unlink: want HTTP 409, got %d", w.Code)
	}
	require.Equal(t, beforeRaw, rawRecords(), "raw records and native source provenance must stay intact")
	require.Zero(t, blocked.requests, "synthetic unlink must remain local")
}
