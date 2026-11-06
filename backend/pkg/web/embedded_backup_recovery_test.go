package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedBackupRestoresNativeLoginAndLiveFamilyPermissions(t *testing.T) {
	sourceDir := t.TempDir()
	source := filepath.Join(sourceDir, "synthetic.db")
	key := "independently-held-synthetic-key"
	cfg := embeddedConfig(t, source, true)
	cfg.Set("database.type", "sqlite")
	cfg.Set("auth.oidc.enabled", false)
	app := openEmbedded(t, cfg)
	require.NoError(t, app.UnlockDatabase(key))
	request := func(a *EmbeddedApplication, method, path, token, body string, status int) json.RawMessage {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		a.ServeHTTP(w, r)
		require.Equal(t, status, w.Code)
		var response struct{ Data json.RawMessage }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		return response.Data
	}
	credentials := `{"username":"synthetic-backup-owner","password":"independent-native-password","join_mailing_list":false}`
	decodeToken := func(raw json.RawMessage) string {
		var token string
		require.NoError(t, json.Unmarshal(raw, &token))
		return token
	}
	oldToken := decodeToken(request(app, "POST", "/api/auth/signup", "", credentials, 200))
	request(app, "POST", "/api/auth/signup", "", `{"username":"synthetic-backup-delegate","password":"synthetic-delegate-password","join_mailing_list":false}`, 200)
	var sources []struct{ ID string }
	require.NoError(t, json.Unmarshal(request(app, "GET", "/api/secure/source", oldToken, "", 200), &sources))
	require.NotEmpty(t, sources)
	base := "/api/secure/source-access/" + sources[0].ID
	var local struct {
		ID string `json:"source_resource_id"`
	}
	require.NoError(t, json.Unmarshal(request(app, "POST", base+"/resources", oldToken, `{"resource_raw":{"resourceType":"Observation","status":"final","code":{"text":"Synthetic recovery record"}}}`, 200), &local))
	var grant struct{ ID string }
	require.NoError(t, json.Unmarshal(request(app, "POST", base+"/grants", oldToken, `{"delegate_username":"synthetic-backup-delegate","access_level":"VIEW"}`, 200), &grant))
	originalUser, err := app.engine.deviceRepo.GetUserByUsername(context.Background(), "synthetic-backup-owner")
	require.NoError(t, err)
	parent := t.TempDir()
	require.NoError(t, os.Chmod(parent, 0700))
	output := filepath.Join(parent, "new-backup")
	_, err = database.BackupEncryptedSQLite(context.Background(), source, output, key)
	require.NoError(t, err)
	require.NoError(t, app.Close())
	require.NoError(t, os.RemoveAll(sourceDir))
	restoredPath := filepath.Join(t.TempDir(), "recovered.db")
	bytes, err := os.ReadFile(filepath.Join(output, "database.sqlite"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(restoredPath, bytes, 0600))
	recovery := embeddedConfig(t, restoredPath, true)
	recovery.Set("database.type", "sqlite")
	recovery.Set("auth.oidc.enabled", false)
	recovery.Set("auth.registration.enabled", false)
	// A new installation key deliberately rejects copied old JWTs. Neither the
	// signing key nor the database recovery key is recovered from the artifact.
	recovery.Set("jwt.issuer.key", strings.Repeat("independent-new-installation-key", 3))
	restored := openEmbedded(t, recovery)
	require.Error(t, restored.UnlockDatabase("synthetic-wrong-key"))
	require.NoError(t, restored.UnlockDatabase(key))
	request(restored, "GET", base+"/resources/Observation/"+local.ID, oldToken, "", 401)
	request(restored, "POST", "/api/auth/signin", "", `{"username":"synthetic-backup-owner","password":"wrong-password"}`, 401)
	owner := decodeToken(request(restored, "POST", "/api/auth/signin", "", credentials, 200))
	delegate := decodeToken(request(restored, "POST", "/api/auth/signin", "", `{"username":"synthetic-backup-delegate","password":"synthetic-delegate-password"}`, 200))
	found, err := restored.engine.deviceRepo.GetUserByUsername(context.Background(), "synthetic-backup-owner")
	require.NoError(t, err)
	require.Equal(t, originalUser.ID, found.ID)
	detail := base + "/resources/Observation/" + local.ID
	request(restored, "GET", detail, owner, "", 200)
	request(restored, "GET", detail, delegate, "", 200)
	request(restored, "PATCH", detail, delegate, fmt.Sprintf(`{"resource_raw":{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic unauthorized edit"}}}`, local.ID), 404)
	request(restored, "DELETE", base+"/grants/"+grant.ID, owner, "", 200)
	request(restored, "GET", detail, delegate, "", 404)
	require.NoError(t, restored.Close())
	restarted := openEmbedded(t, recovery)
	require.NoError(t, restarted.UnlockDatabase(key))
	request(restarted, "GET", detail, owner, "", 200)
	request(restarted, "GET", detail, delegate, "", 404)
	request(restarted, "POST", "/api/auth/signup", "", `{"username":"synthetic-unwanted-provisioning","password":"synthetic-password"}`, 403)
}
