package database

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
)

var ErrSourceAccessDenied = errors.New("source access denied")
var ErrSourceAccessInput = errors.New("invalid source access request")

// SourceAccessRepository keeps explicit delegation separate from owner-only APIs,
// source credential access, account summaries and administrative user management.
type SourceAccessRepository interface {
	SetSourceGrant(context.Context, uuid.UUID, string, models.SourceAccessLevel, *time.Time) (*models.SourceGrant, error)
	RevokeSourceGrant(context.Context, uuid.UUID, uuid.UUID) error
	ListSourceGrants(context.Context, uuid.UUID) ([]models.SourceGrant, error)
	ListSourceAccessAudit(context.Context, uuid.UUID) ([]models.SourceAccessAudit, error)
	ListSharedSources(context.Context) ([]models.SharedSource, error)
	GetSharedSourceSummary(context.Context, uuid.UUID) (*models.SharedSourceSummary, error)
	ListSharedResources(context.Context, uuid.UUID, string, int) ([]models.ResourceBase, error)
	GetSharedResource(context.Context, uuid.UUID, string, string) (*models.SharedResource, error)
	CreateLocalResource(context.Context, uuid.UUID, json.RawMessage) (*models.ResourceBase, error)
	UpdateSharedResource(context.Context, uuid.UUID, string, string, json.RawMessage) error
}

var _ SourceAccessRepository = (*GormRepository)(nil)
