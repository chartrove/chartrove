package database

import (
	"context"
	"errors"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"time"
)

var ErrOIDCDenied = errors.New("OIDC authentication denied")

// Separate capability keeps legacy repositories and owner-only data APIs intact.
type OIDCRepository interface {
	CompleteOIDC(context.Context, string, string, uuid.UUID, string, time.Time, time.Time) (*models.User, *models.OIDCSession, error)
	ValidateOIDCSession(context.Context, uuid.UUID, uuid.UUID, string) error
	RevokeOIDCSession(context.Context, uuid.UUID, uuid.UUID) error
	RevokeOIDCIdentity(context.Context, uuid.UUID, uuid.UUID) error
	ListOIDCIdentities(context.Context, uuid.UUID) ([]models.OIDCIdentity, error)
	ListOIDCAudit(context.Context, uuid.UUID) ([]models.OIDCAudit, error)
}

var _ OIDCRepository = (*GormRepository)(nil)
