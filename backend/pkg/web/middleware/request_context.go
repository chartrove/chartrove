package middleware

import (
	"context"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthenticatedRequestContext preserves the request lifetime and copies only
// server-authenticated account values. It never retains pooled *gin.Context:
// database/sql may still inspect a context after the HTTP handler has returned.
// Missing/invalid Gin identity values mask any inherited account context.
func AuthenticatedRequestContext(c *gin.Context) context.Context {
	id := uuid.Nil
	if value, ok := c.Get(pkg.ContextKeyTypeAuthUserID); ok {
		if account, ok := value.(uuid.UUID); ok {
			id = account
		}
	}
	ctx := context.WithValue(c.Request.Context(), pkg.ContextKeyTypeAuthUsername, c.GetString(pkg.ContextKeyTypeAuthUsername))
	return context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, id)
}
