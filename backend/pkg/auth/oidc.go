package auth

// Generic discovery follows the original Fasten #613 / Chartrove #214 proposal
// by TheoAtTechStack. Account binding, browser proof and sessions are rewritten.
import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var ErrOIDC = errors.New("OIDC authentication could not be completed")

type OIDCFlow struct {
	Secret          [32]byte
	Nonce, Verifier string
	LinkUser        uuid.UUID
	PasswordHash    string
	AuthorizedAt    time.Time
	Expires         time.Time
	Provider        *oidc.Provider
	OAuth           oauth2.Config
}

type OIDCStart struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
	BrowserSecret    string `json:"-"`
}

type OIDCVerified struct {
	Issuer, Subject string
	LinkUser        uuid.UUID
	PasswordHash    string
	AuthorizedAt    time.Time
	Expires         time.Time
}

// Transactions are bounded, browser-bound, single-use and deliberately ephemeral.
// A restart invalidates pending logins, without depending on the IdP for recovery.
type OIDCManager struct {
	Config         config.Interface
	Client         *http.Client
	mu             sync.Mutex
	flows          map[string]OIDCFlow
	discoverySlots chan struct{}
}

func NewOIDCManager(cfg config.Interface) *OIDCManager {
	return &OIDCManager{Config: cfg, Client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || req.URL.Scheme != "https" {
			return ErrOIDC
		}
		return nil
	}}, flows: make(map[string]OIDCFlow), discoverySlots: make(chan struct{}, 8)}
}

func secureOIDCURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.Fragment == "" && u.RawQuery == ""
}

func (m *OIDCManager) Enabled() bool { return m.Config.GetBool("auth.oidc.enabled") }

func (m *OIDCManager) Start(ctx context.Context, linkUser uuid.UUID, passwordHash string) (*OIDCStart, error) {
	authorizedAt := time.Now().UTC()
	issuer := m.Config.GetString("auth.oidc.issuer")
	redirect := m.Config.GetString("auth.oidc.redirect_uri")
	if !m.Enabled() || !secureOIDCURL(issuer) || !secureOIDCURL(redirect) || strings.TrimSpace(m.Config.GetString("auth.oidc.client_id")) == "" {
		return nil, ErrOIDC
	}
	select {
	case m.discoverySlots <- struct{}{}:
		defer func() { <-m.discoverySlots }()
	default:
		return nil, ErrOIDC
	}
	// Discovery is lazy: an unavailable provider cannot prevent local startup/login.
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctx = oidc.ClientContext(ctx, m.Client)
	p, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, ErrOIDC
	}
	var metadata struct {
		JWKS string `json:"jwks_uri"`
	}
	if p.Claims(&metadata) != nil || !secureOIDCURL(metadata.JWKS) || !secureOIDCURL(p.Endpoint().AuthURL) || !secureOIDCURL(p.Endpoint().TokenURL) {
		return nil, ErrOIDC
	}
	oauth := oauth2.Config{ClientID: m.Config.GetString("auth.oidc.client_id"), ClientSecret: m.Config.GetString("auth.oidc.client_secret"), RedirectURL: redirect, Endpoint: p.Endpoint(), Scopes: []string{oidc.ScopeOpenID}}
	state, secret, nonce, verifier := oauth2.GenerateVerifier(), oauth2.GenerateVerifier(), oauth2.GenerateVerifier(), oauth2.GenerateVerifier()
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, flow := range m.flows {
		if !flow.Expires.After(time.Now()) {
			delete(m.flows, key)
		}
	}
	if len(m.flows) >= 256 {
		return nil, ErrOIDC
	}
	m.flows[state] = OIDCFlow{Secret: sha256.Sum256([]byte(secret)), Nonce: nonce, Verifier: verifier, LinkUser: linkUser, PasswordHash: passwordHash, AuthorizedAt: authorizedAt, Expires: time.Now().Add(5 * time.Minute), Provider: p, OAuth: oauth}
	return &OIDCStart{State: state, BrowserSecret: secret, AuthorizationURL: oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oidc.Nonce(nonce), oauth2.SetAuthURLParam("prompt", "select_account"))}, nil
}

func (m *OIDCManager) Complete(ctx context.Context, state, secret, code string) (*OIDCVerified, error) {
	if !m.Enabled() || len(state) > 128 || len(secret) > 128 || code == "" || len(code) > 8192 {
		return nil, ErrOIDC
	}
	proof := sha256.Sum256([]byte(secret))
	m.mu.Lock()
	flow, ok := m.flows[state]
	if !ok || subtle.ConstantTimeCompare(proof[:], flow.Secret[:]) != 1 {
		m.mu.Unlock()
		return nil, ErrOIDC
	}
	delete(m.flows, state)
	m.mu.Unlock()
	if !flow.Expires.After(time.Now()) || flow.OAuth.ClientID != m.Config.GetString("auth.oidc.client_id") || flow.OAuth.RedirectURL != m.Config.GetString("auth.oidc.redirect_uri") {
		return nil, ErrOIDC
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ctx = oidc.ClientContext(ctx, m.Client)
	token, err := flow.OAuth.Exchange(ctx, code, oauth2.VerifierOption(flow.Verifier))
	if err != nil {
		return nil, ErrOIDC
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrOIDC
	}
	id, err := flow.Provider.VerifierContext(oidc.ClientContext(context.Background(), m.Client), &oidc.Config{ClientID: flow.OAuth.ClientID, SupportedSigningAlgs: []string{oidc.RS256, oidc.RS384, oidc.RS512, oidc.ES256, oidc.ES384, oidc.ES512, oidc.PS256, oidc.PS384, oidc.PS512, oidc.EdDSA}}).Verify(ctx, raw)
	if err != nil || id.Subject == "" || len(id.Subject) > 255 || id.Issuer != m.Config.GetString("auth.oidc.issuer") || id.Nonce != flow.Nonce || id.IssuedAt.IsZero() || id.IssuedAt.After(time.Now().Add(30*time.Second)) {
		return nil, ErrOIDC
	}
	var claims struct {
		AuthorizedParty string `json:"azp"`
	}
	if id.Claims(&claims) != nil || (len(id.Audience) > 1 && claims.AuthorizedParty != flow.OAuth.ClientID) || (claims.AuthorizedParty != "" && claims.AuthorizedParty != flow.OAuth.ClientID) {
		return nil, ErrOIDC
	}
	if id.AccessTokenHash != "" && id.VerifyAccessToken(token.AccessToken) != nil {
		return nil, ErrOIDC
	}
	lifetime, err := config.UserSessionDuration(m.Config)
	if err != nil {
		return nil, ErrOIDC
	}
	expires := time.Now().Add(lifetime)
	if id.Expiry.Before(expires) {
		expires = id.Expiry
	}
	return &OIDCVerified{Issuer: id.Issuer, Subject: id.Subject, LinkUser: flow.LinkUser, PasswordHash: flow.PasswordHash, AuthorizedAt: flow.AuthorizedAt, Expires: expires}, nil
}

func JwtGenerateOIDCSession(user models.User, key string, expires time.Time, id string) (string, error) {
	return generateToken(user, key, expires, id, "oidc")
}
