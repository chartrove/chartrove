package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestOIDCHTTPNativeAuthorizationAndLocalRecovery(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var issuer, nonce, challenge string
	subject := "synthetic-delegate-subject"
	idp := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "synthetic", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			require.NoError(t, r.ParseForm())
			hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			require.Equal(t, challenge, base64.RawURLEncoding.EncodeToString(hash[:]))
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"iss": issuer, "sub": subject, "aud": "synthetic", "nonce": nonce, "exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix(), "email": "duplicate@example.invalid", "role": "admin", "groups": []string{"administrators"}})
			token.Header["kid"] = "synthetic"
			raw, err := token.SignedString(key)
			require.NoError(t, err)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "synthetic", "token_type": "Bearer", "id_token": raw})
		default:
			w.WriteHeader(404)
		}
	}))
	defer idp.Close()
	issuer = idp.URL
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-oidc-http.db"))
	cfg.Set("database.encryption.enabled", false)
	cfg.Set("auth.oidc.enabled", true)
	cfg.Set("auth.oidc.issuer", issuer)
	cfg.Set("auth.oidc.client_id", "synthetic")
	cfg.Set("auth.oidc.redirect_uri", "https://chartrove.example/web/auth/oidc/callback")
	db, err := database.NewRepository(cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	owner := &models.User{Username: "synthetic-local-admin", Password: "synthetic-password", Email: "duplicate@example.invalid", Role: pkg.UserRoleAdmin}
	delegate := &models.User{Username: "synthetic-local-delegate", Password: "synthetic-password", Email: owner.Email, Role: pkg.UserRoleUser}
	for _, u := range []*models.User{owner, delegate} {
		require.NoError(t, db.CreateUser(context.Background(), u))
	}
	r := gin.New()
	r.Use(middleware.RepositoryMiddleware(db), middleware.ConfigMiddleware(cfg), middleware.LoggerMiddleware(logrus.NewEntry(logrus.New())))
	api := r.Group("/api")
	api.POST("/auth/signin", AuthSignin)
	secure := api.Group("/secure").Use(middleware.RequireAuth())
	secure.GET("/account/me", GetCurrentUser)
	secure.GET("/source", ListSource)
	secure.GET("/summary", GetSummary)
	secure.GET("/resource/fhir/:sourceId/:resourceId", GetResourceFhir)
	secure.POST("/users", CreateUser)
	RegisterSourceAccessRoutes(secure)
	m := auth.NewOIDCManager(cfg)
	m.Client = idp.Client()
	RegisterOIDCRoutes(api, secure, m)
	var browserCookie *http.Cookie
	request := func(method, path, token string, body interface{}, status int) json.RawMessage {
		t.Helper()
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(method, "/api"+path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		if path == "/auth/oidc/complete" && browserCookie != nil {
			req.AddCookie(browserCookie)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if strings.Contains(path, "/search/") {
			require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		}
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == oidcCookieName {
				require.True(t, cookie.HttpOnly)
				require.True(t, cookie.Secure)
				require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
				require.Equal(t, "/api/auth/oidc/complete", cookie.Path)
				if cookie.MaxAge < 0 {
					browserCookie = nil
				} else {
					browserCookie = cookie
					require.NotContains(t, w.Body.String(), cookie.Value)
				}
			}
		}
		require.Equal(t, status, w.Code, w.Body.String())
		var result struct {
			Data json.RawMessage `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		return result.Data
	}
	login := func(username string) string {
		var token string
		require.NoError(t, json.Unmarshal(request("POST", "/auth/signin", "", map[string]string{"username": username, "password": "synthetic-password"}, 200), &token))
		return token
	}
	localOwner, localDelegate := login(owner.Username), login(delegate.Username)
	crossSite := httptest.NewRequest("POST", "/api/auth/oidc/start", nil)
	crossSite.Header.Set("Origin", "https://other.example")
	denied := httptest.NewRecorder()
	r.ServeHTTP(denied, crossSite)
	require.Equal(t, 403, denied.Code)
	require.Empty(t, denied.Result().Cookies())
	flow := func(token, password string) auth.OIDCStart {
		var s auth.OIDCStart
		path := "/auth/oidc/start"
		body := map[string]string{}
		if token != "" {
			path = "/secure/auth/oidc/link"
			body["password"] = password
		}
		require.NoError(t, json.Unmarshal(request("POST", path, token, body, 200), &s))
		require.Empty(t, s.BrowserSecret)
		require.NotNil(t, browserCookie)
		u, err := url.Parse(s.AuthorizationURL)
		require.NoError(t, err)
		nonce = u.Query().Get("nonce")
		challenge = u.Query().Get("code_challenge")
		return s
	}
	complete := func(s auth.OIDCStart, status int) string {
		var token string
		data := request("POST", "/auth/oidc/complete", "", map[string]string{"state": s.State, "code": "synthetic-code"}, status)
		if status == 200 {
			require.NoError(t, json.Unmarshal(data, &token))
		}
		return token
	}
	request("POST", "/secure/auth/oidc/link", localDelegate, map[string]string{"password": "wrong"}, 401)
	complete(flow("", ""), 401) // Unknown issuer/subject never matches duplicate email.
	oidcToken := complete(flow(localDelegate, "synthetic-password"), 200)
	request("POST", "/auth/oidc/complete", "", map[string]string{"state": "replayed", "code": "synthetic-code"}, 401)
	var profile struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	require.NoError(t, json.Unmarshal(request("GET", "/secure/account/me", oidcToken, nil, 200), &profile))
	require.Equal(t, delegate.ID.String(), profile.ID)
	require.Equal(t, "user", profile.Role)
	request("POST", "/secure/users", oidcToken, map[string]string{"username": "synthetic-escalation", "password": "synthetic-password", "role": "admin"}, 401)
	// A different authenticated native user cannot claim an existing binding.
	complete(flow(localOwner, "synthetic-password"), 401)
	ownerCtx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	sources, err := db.GetSources(ownerCtx)
	require.NoError(t, err)
	source := sources[0].ID
	access := db.(database.SourceAccessRepository)
	resource, err := access.CreateLocalResource(ownerCtx, source, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic private observation"}}`))
	require.NoError(t, err)
	path := "/secure/source-access/" + source.String() + "/resources/Observation/" + resource.SourceResourceID
	searchPath := "/secure/source-access/" + source.String() + "/search/Observation"
	request("POST", searchPath, oidcToken, map[string]interface{}{"text": "Synthetic", "page": 0}, 404)
	request("GET", path, oidcToken, nil, 404)
	grant, err := access.SetSourceGrant(ownerCtx, source, delegate.Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	request("GET", path, oidcToken, nil, 200)
	hits := request("POST", searchPath, oidcToken, map[string]interface{}{"text": resource.SourceResourceID, "page": 0}, 200)
	require.Contains(t, string(hits), resource.SourceResourceID)
	require.NotContains(t, string(hits), "resource_raw")
	request("POST", searchPath, oidcToken, map[string]interface{}{"text": "Synthetic", "page": -1}, 400)
	request("POST", searchPath, oidcToken, map[string]interface{}{"text": strings.Repeat("s", 5000), "page": 0}, 400)
	request("POST", searchPath, oidcToken, map[string]interface{}{"text": "Synthetic", "page": 0, "owner_user_id": owner.ID}, 400)
	// Delegation never opens the existing owner-only detail API.
	request("GET", "/secure/resource/fhir/"+source.String()+"/"+resource.SourceResourceID, oidcToken, nil, 500)
	request("PATCH", path, oidcToken, map[string]interface{}{"resource_raw": map[string]string{"resourceType": "Observation", "id": resource.SourceResourceID}}, 404)
	grant, err = access.SetSourceGrant(ownerCtx, source, delegate.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	request("PATCH", path, oidcToken, map[string]interface{}{"resource_raw": map[string]interface{}{"resourceType": "Observation", "id": resource.SourceResourceID, "status": "amended", "code": map[string]string{"text": "Synthetic delegated edit"}}}, 200)
	var ownSources []models.SourceCredential
	require.NoError(t, json.Unmarshal(request("GET", "/secure/source", oidcToken, nil, 200), &ownSources))
	for _, s := range ownSources {
		require.Equal(t, delegate.ID, s.UserID)
	}
	require.NoError(t, access.RevokeSourceGrant(ownerCtx, source, grant.ID))
	request("POST", searchPath, oidcToken, map[string]interface{}{"text": "Synthetic", "page": 0}, 404)
	request("GET", path, oidcToken, nil, 404)
	request("POST", "/secure/auth/oidc/logout", oidcToken, map[string]string{}, 200)
	request("GET", "/secure/account/me", oidcToken, nil, 401)
	oidcToken = complete(flow("", ""), 200)
	var identities []models.OIDCIdentity
	require.NoError(t, json.Unmarshal(request("GET", "/secure/auth/oidc/identities", oidcToken, nil, 200), &identities))
	require.Len(t, identities, 1)
	revoke := "/secure/auth/oidc/identities/" + identities[0].ID.String() + "/revoke"
	request("POST", revoke, localOwner, map[string]string{"password": "synthetic-password"}, 404)
	request("POST", revoke, localDelegate, map[string]string{"password": "synthetic-password"}, 200)
	request("GET", "/secure/account/me", oidcToken, nil, 401)
	complete(flow("", ""), 401)
	cfg.Set("auth.oidc.enabled", false)
	idp.Close()
	request("GET", "/secure/account/me", localOwner, nil, 200)
	require.NotEmpty(t, login(owner.Username))
	require.NotEmpty(t, login(delegate.Username))
	request("POST", "/auth/oidc/start", "", nil, 503)
}
