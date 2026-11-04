package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func oidcJSON(c *gin.Context, status int, data interface{}) {
	c.Header("Cache-Control", "no-store")
	c.JSON(status, gin.H{"success": status == http.StatusOK, "data": data})
}

const oidcCookieName = "__Secure-chartrove_oidc"

func oidcBrowserCookie(c *gin.Context, secret string, maxAge int) {
	cfg := c.MustGet(pkg.ContextKeyTypeConfig).(config.Interface)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oidcCookieName, secret, maxAge, strings.TrimRight(cfg.GetString("web.listen.basepath"), "/")+"/api/auth/oidc/complete", "", true, true)
}

func oidcOrigin(c *gin.Context) bool {
	if origin := c.GetHeader("Origin"); origin != "" {
		cfg := c.MustGet(pkg.ContextKeyTypeConfig).(config.Interface)
		u, err := url.Parse(cfg.GetString("auth.oidc.redirect_uri"))
		if err != nil || origin != u.Scheme+"://"+u.Host {
			oidcJSON(c, 403, nil)
			return false
		}
	}
	return true
}

func oidcBody(c *gin.Context, dst interface{}) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if d.Decode(dst) != nil || d.Decode(new(interface{})) != io.EOF {
		oidcJSON(c, 400, nil)
		return false
	}
	return true
}

func oidcRepos(c *gin.Context) (database.DatabaseRepository, database.OIDCRepository, bool) {
	r := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)
	o, ok := r.(database.OIDCRepository)
	if !ok {
		oidcJSON(c, 503, nil)
	}
	return r, o, ok
}

func oidcPasswordUser(c *gin.Context, r database.DatabaseRepository) (*models.User, bool) {
	var body struct {
		Password string `json:"password"`
	}
	if !oidcBody(c, &body) {
		return nil, false
	}
	u, err := r.GetCurrentUser(c)
	if err != nil || u == nil || u.DeletedAt != nil || len(body.Password) > 1024 || u.CheckPassword(body.Password) != nil {
		oidcJSON(c, 401, nil)
		return nil, false
	}
	return u, true
}

// All account association is explicit, with local password proof before redirect.
// The callback accepts only a code and browser-bound transaction proof, never an
// externally supplied account ID, email, token, owner ID or role.
func RegisterOIDCRoutes(api, secure gin.IRoutes, mgr *auth.OIDCManager) {
	api.GET("/auth/oidc", func(c *gin.Context) { oidcJSON(c, 200, gin.H{"enabled": mgr.Enabled()}) })
	api.POST("/auth/oidc/start", func(c *gin.Context) {
		if !oidcOrigin(c) {
			return
		}
		if _, _, ok := oidcRepos(c); !ok {
			return
		}
		start, err := mgr.Start(c.Request.Context(), uuid.Nil, "")
		if err != nil {
			oidcJSON(c, 503, nil)
			return
		}
		oidcBrowserCookie(c, start.BrowserSecret, 300)
		oidcJSON(c, 200, start)
	})
	secure.POST("/auth/oidc/link", func(c *gin.Context) {
		if !oidcOrigin(c) {
			return
		}
		r, _, ok := oidcRepos(c)
		if !ok {
			return
		}
		user, ok := oidcPasswordUser(c, r)
		if !ok {
			return
		}
		start, err := mgr.Start(c.Request.Context(), user.ID, user.Password)
		if err != nil {
			oidcJSON(c, 503, nil)
			return
		}
		oidcBrowserCookie(c, start.BrowserSecret, 300)
		oidcJSON(c, 200, start)
	})
	api.POST("/auth/oidc/complete", func(c *gin.Context) {
		if !oidcOrigin(c) {
			return
		}
		_, repo, ok := oidcRepos(c)
		if !ok {
			return
		}
		var body struct {
			State string `json:"state"`
			Code  string `json:"code"`
		}
		if !oidcBody(c, &body) {
			return
		}
		secret, err := c.Cookie(oidcCookieName)
		if err != nil {
			oidcJSON(c, 401, nil)
			return
		}
		oidcBrowserCookie(c, "", -1)
		verified, err := mgr.Complete(c.Request.Context(), body.State, secret, body.Code)
		if err != nil {
			oidcJSON(c, 401, nil)
			return
		}
		user, session, err := repo.CompleteOIDC(c, verified.Issuer, verified.Subject, verified.LinkUser, verified.PasswordHash, verified.AuthorizedAt, verified.Expires)
		if err != nil {
			status := 500
			if errors.Is(err, database.ErrOIDCDenied) {
				status = 401
			}
			oidcJSON(c, status, nil)
			return
		}
		cfg := c.MustGet(pkg.ContextKeyTypeConfig).(config.Interface)
		token, err := auth.JwtGenerateOIDCSession(*user, cfg.GetString("jwt.issuer.key"), session.ExpiresAt, session.ID.String())
		if err != nil {
			_ = repo.RevokeOIDCSession(c, user.ID, session.ID)
			oidcJSON(c, 500, nil)
			return
		}
		oidcJSON(c, 200, token)
	})
	secure.GET("/auth/oidc/identities", func(c *gin.Context) {
		r, repo, ok := oidcRepos(c)
		if !ok {
			return
		}
		u, err := r.GetCurrentUser(c)
		if err != nil {
			oidcJSON(c, 500, nil)
			return
		}
		rows, err := repo.ListOIDCIdentities(c, u.ID)
		if err != nil {
			oidcJSON(c, 500, nil)
			return
		}
		oidcJSON(c, 200, rows)
	})
	secure.GET("/auth/oidc/audit", func(c *gin.Context) {
		r, repo, ok := oidcRepos(c)
		if !ok {
			return
		}
		u, err := r.GetCurrentUser(c)
		if err != nil {
			oidcJSON(c, 500, nil)
			return
		}
		rows, err := repo.ListOIDCAudit(c, u.ID)
		if err != nil {
			oidcJSON(c, 500, nil)
			return
		}
		oidcJSON(c, 200, rows)
	})
	secure.POST("/auth/oidc/identities/:identityID/revoke", func(c *gin.Context) {
		r, repo, ok := oidcRepos(c)
		if !ok {
			return
		}
		u, ok := oidcPasswordUser(c, r)
		if !ok {
			return
		}
		id, err := uuid.Parse(c.Param("identityID"))
		if err != nil {
			oidcJSON(c, 400, nil)
			return
		}
		if err = repo.RevokeOIDCIdentity(c, u.ID, id); err != nil {
			status := 500
			if errors.Is(err, database.ErrOIDCDenied) {
				status = 404
			}
			oidcJSON(c, status, nil)
			return
		}
		oidcJSON(c, 200, nil)
	})
	secure.POST("/auth/oidc/logout", func(c *gin.Context) {
		r, repo, ok := oidcRepos(c)
		if !ok {
			return
		}
		u, err := r.GetCurrentUser(c)
		if err != nil {
			oidcJSON(c, 500, nil)
			return
		}
		cfg := c.MustGet(pkg.ContextKeyTypeConfig).(config.Interface)
		claims, err := auth.JwtValidateFastenToken(cfg.GetString("jwt.issuer.key"), c.GetString(pkg.ContextKeyTypeAuthToken))
		if err != nil {
			oidcJSON(c, 401, nil)
			return
		}
		if claims.TokenType == "oidc" {
			id, err := uuid.Parse(claims.ID)
			if err != nil {
				oidcJSON(c, 401, nil)
				return
			}
			if err = repo.RevokeOIDCSession(c, u.ID, id); err != nil {
				oidcJSON(c, 500, nil)
				return
			}
		}
		oidcJSON(c, 200, nil)
	})
}
