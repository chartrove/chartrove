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
		user, err := databaseRepo.GetUserByUsername(c, claims.Subject)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Authentication service unavailable"})
			return
		}
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid account"})
			return
		}

		if claims.TokenType == "access" {
			token, err := databaseRepo.GetAccessTokenByTokenIDAndUsername(c, claims.ID, claims.Subject)
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

		// Set context for both regular and access tokens
		c.Set(pkg.ContextKeyTypeAuthToken, tokenString)
		c.Set(pkg.ContextKeyTypeAuthUsername, claims.Subject)
		c.Next()
	}
}
