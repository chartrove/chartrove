package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		appConfig := c.MustGet(pkg.ContextKeyTypeConfig).(config.Interface)

		authHeader := c.GetHeader("Authorization")
		authHeaderParts := strings.Fields(authHeader)

		if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "request does not contain a valid token"})
			c.Abort()
			return
		}

		tokenString := authHeaderParts[1]

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "request does not contain an access token"})
			c.Abort()
			return
		}

		claims, err := auth.JwtValidateFastenToken(appConfig.GetString("jwt.issuer.key"), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
			c.Abort()
			return
		}

		if claims.Subject == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid account"})
			return
		}
		databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)
		user, err := databaseRepo.GetUserByUsername(c.Request.Context(), claims.Subject)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Authentication service unavailable"})
			return
		}
		if err != nil || user == nil || user.DeletedAt != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid account"})
			return
		}
		// Bind sessions to an immutable account, never to a reusable username.
		// Old ordinary JWTs have no immutable identity and require a fresh local
		// login. Account timestamps cannot prove continuity: historical creation
		// requests could supply them. Legacy API/OIDC tokens have independent
		// UUID-bound database checks below and retain those revocation contracts.
		invalidIdentity := user.ID == uuid.Nil
		if claims.NativeUserID != "" {
			invalidIdentity = invalidIdentity || claims.NativeUserID != user.ID.String()
		} else if claims.TokenType == "" {
			invalidIdentity = true
		}
		if invalidIdentity {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid account session"})
			return
		}

		if claims.TokenType == "access" {
			token, err := databaseRepo.GetAccessTokenByTokenIDAndUsername(c.Request.Context(), claims.ID, claims.Subject)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Authentication service unavailable"})
				return
			}
			if err != nil || token == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid access token"})
				c.Abort()
				return
			}
		}
		if claims.TokenType == "oidc" {
			repo, ok := databaseRepo.(database.OIDCRepository)
			id, parseErr := uuid.Parse(claims.ID)
			if !ok || parseErr != nil || !appConfig.GetBool("auth.oidc.enabled") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid OIDC session"})
				return
			}
			err := repo.ValidateOIDCSession(c.Request.Context(), id, user.ID, appConfig.GetString("auth.oidc.issuer"))
			if err != nil {
				status := http.StatusInternalServerError
				if errors.Is(err, database.ErrOIDCDenied) {
					status = http.StatusUnauthorized
				}
				c.AbortWithStatusJSON(status, gin.H{"success": false, "error": "OIDC session unavailable"})
				return
			}
		}

		// Set context for both regular and access tokens
		c.Set(pkg.ContextKeyTypeAuthToken, tokenString)
		c.Set(pkg.ContextKeyTypeAuthUsername, claims.Subject)
		c.Set(pkg.ContextKeyTypeAuthUserID, user.ID)
		c.Next()
	}
}
