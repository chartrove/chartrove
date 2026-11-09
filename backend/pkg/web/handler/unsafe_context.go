package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// unsafeDiagnosticOwner is a query scope, NOT caller authentication. Only the
// explicit unsafe-mode route may select an owner without proving its password.
// Copy immutable values, never Gin, a user pointer, roles or an application JWT.
type unsafeDiagnosticOwner struct {
	context.Context
	id       uuid.UUID
	username string
}

func (c unsafeDiagnosticOwner) Value(key any) any {
	switch key {
	case pkg.ContextKeyTypeAuthUsername:
		return c.username
	case pkg.ContextKeyTypeAuthUserID:
		return c.id
	case pkg.ContextKeyTypeAuthToken:
		return nil
	default:
		return c.Context.Value(key)
	}
}

func resolveUnsafeDiagnosticOwner(parent context.Context, repo database.DatabaseRepository, username string) (context.Context, error) {
	// The lookup is explicitly named; inherited account/token values are not its
	// authority. The repository's normal active-user lookup excludes deletion.
	lookup := unsafeDiagnosticOwner{Context: parent}
	if err := lookup.Err(); err != nil {
		return nil, err
	}
	user, err := repo.GetUserByUsername(lookup, username)
	if err != nil {
		return nil, err
	}
	if user == nil || user.ID == uuid.Nil || user.DeletedAt != nil || user.Username != username {
		return nil, database.ErrSourceAccessDenied
	}
	return unsafeDiagnosticOwner{Context: parent, id: user.ID, username: user.Username}, nil
}

// Call only after the handler's independent unsafe-mode guard. Do not put the
// selected owner into Gin authentication claims or reuse RequireAuth's adapter.
func unsafeOwnerForRequest(c *gin.Context, repo database.DatabaseRepository) (context.Context, bool) {
	ctx, err := resolveUnsafeDiagnosticOwner(c.Request.Context(), repo, c.Param("username"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "diagnostic owner unavailable"})
		return nil, false
	}
	return ctx, true
}

// A provider may have rotated its refresh credential before the caller aborts.
// Preserve the existing deferred persistence intent, with an independent bound
// and the same immutable selected owner. This is synchronous cleanup, not a new
// detached worker/job or a general-purpose cancellation bypass.
func unsafeCredentialCleanupContext(selected context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(selected), 10*time.Second)
}
