package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOIDCDiscoveryPKCEAndValidation(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var issuer, nonce, challenge string
	var mutate func(jwt.MapClaims)
	var invalidToken bool
	var wrongSignature bool
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "synthetic-key", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			require.NoError(t, r.ParseForm())
			hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			require.Equal(t, challenge, base64.RawURLEncoding.EncodeToString(hash[:]))
			require.Equal(t, "authorization_code", r.Form.Get("grant_type"))
			require.Equal(t, "https://chartrove.example/web/auth/oidc/callback", r.Form.Get("redirect_uri"))
			claims := jwt.MapClaims{"iss": issuer, "sub": "synthetic-subject", "aud": "synthetic-client", "nonce": nonce, "iat": time.Now().Unix(), "exp": time.Now().Add(3 * time.Minute).Unix(), "email": "duplicate@example.invalid", "role": "admin"}
			if mutate != nil {
				mutate(claims)
			}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "synthetic-key"
			raw, err := token.SignedString(key)
			require.NoError(t, err)
			if wrongSignature {
				raw, err = token.SignedString(otherKey)
				require.NoError(t, err)
			}
			if invalidToken {
				raw = "invalid-token"
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "synthetic-token", "token_type": "Bearer", "id_token": raw})
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	issuer = server.URL
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("auth.oidc.enabled", true)
	cfg.Set("auth.oidc.issuer", issuer)
	cfg.Set("auth.oidc.client_id", "synthetic-client")
	cfg.Set("auth.oidc.redirect_uri", "https://chartrove.example/web/auth/oidc/callback")
	m := NewOIDCManager(cfg)
	m.Client = server.Client()
	start := func() *OIDCStart {
		s, err := m.Start(context.Background(), uuid.Nil, "")
		require.NoError(t, err)
		u, err := url.Parse(s.AuthorizationURL)
		require.NoError(t, err)
		nonce = u.Query().Get("nonce")
		challenge = u.Query().Get("code_challenge")
		require.NotEmpty(t, nonce)
		require.NotEmpty(t, challenge)
		require.Equal(t, "S256", u.Query().Get("code_challenge_method"))
		require.Equal(t, "code", u.Query().Get("response_type"))
		require.Equal(t, s.State, u.Query().Get("state"))
		return s
	}
	t.Run("success and replay", func(t *testing.T) {
		s := start()
		verified, err := m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
		require.NoError(t, err)
		require.Equal(t, "synthetic-subject", verified.Subject)
		require.Equal(t, issuer, verified.Issuer)
		require.True(t, verified.Expires.Before(time.Now().Add(4*time.Minute)))
		_, err = m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
		require.ErrorIs(t, err, ErrOIDC)
	})
	for name, change := range map[string]func(jwt.MapClaims){
		"issuer":                         func(c jwt.MapClaims) { c["iss"] = "https://other.example" },
		"audience":                       func(c jwt.MapClaims) { c["aud"] = "other-client" },
		"nonce":                          func(c jwt.MapClaims) { c["nonce"] = "other-nonce" },
		"expired token":                  func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Minute).Unix() },
		"future token":                   func(c jwt.MapClaims) { c["iat"] = time.Now().Add(time.Hour).Unix() },
		"empty subject":                  func(c jwt.MapClaims) { c["sub"] = "" },
		"multiple audiences without azp": func(c jwt.MapClaims) { c["aud"] = []string{"synthetic-client", "other"} },
		"conflicting azp":                func(c jwt.MapClaims) { c["azp"] = "other" },
		"invalid access token hash":      func(c jwt.MapClaims) { c["at_hash"] = "wrong" },
	} {
		t.Run(name, func(t *testing.T) {
			s := start()
			mutate = change
			_, err := m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
			require.ErrorIs(t, err, ErrOIDC)
			mutate = nil
		})
	}
	t.Run("invalid signature and token", func(t *testing.T) {
		s := start()
		wrongSignature = true
		_, err := m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
		require.ErrorIs(t, err, ErrOIDC)
		wrongSignature = false
		s = start()
		invalidToken = true
		_, err = m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
		require.ErrorIs(t, err, ErrOIDC)
		invalidToken = false
	})
	t.Run("state secret expiry and disable", func(t *testing.T) {
		s := start()
		for _, pair := range [][2]string{{"wrong", s.BrowserSecret}, {s.State, "wrong"}} {
			_, err := m.Complete(context.Background(), pair[0], pair[1], "synthetic-code")
			require.ErrorIs(t, err, ErrOIDC)
		}
		f := m.flows[s.State]
		f.Expires = time.Now().Add(-time.Minute)
		m.flows[s.State] = f
		_, err := m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
		require.ErrorIs(t, err, ErrOIDC)
		s = start()
		cfg.Set("auth.oidc.enabled", false)
		_, err = m.Complete(context.Background(), s.State, s.BrowserSecret, "synthetic-code")
		require.ErrorIs(t, err, ErrOIDC)
	})
	t.Run("insecure issuer or callback rejected", func(t *testing.T) {
		cfg.Set("auth.oidc.enabled", true)
		cfg.Set("auth.oidc.redirect_uri", "http://chartrove.example/callback")
		_, err := m.Start(context.Background(), uuid.Nil, "")
		require.ErrorIs(t, err, ErrOIDC)
	})
}
