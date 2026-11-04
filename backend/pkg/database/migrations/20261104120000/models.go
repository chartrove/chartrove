// Frozen schema for the additive per-source permissions migration.
package _20261104120000

import (
	"time"

	"github.com/google/uuid"
)

type SourceAccessLevel string

const (
	SourceAccessView SourceAccessLevel = "VIEW"
	SourceAccessEdit SourceAccessLevel = "EDIT"
)

// SourceGrant grants access to exactly one owner's source. Administrative roles
// are deliberately absent. Revocation retains the grant and its audit history.
type SourceGrant struct {
	ID             uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey"`
	OwnerUserID    uuid.UUID         `json:"owner_user_id" gorm:"type:uuid;not null;uniqueIndex:source_delegate"`
	SourceID       uuid.UUID         `json:"source_id" gorm:"type:uuid;not null;uniqueIndex:source_delegate"`
	DelegateUserID uuid.UUID         `json:"delegate_user_id" gorm:"type:uuid;not null;uniqueIndex:source_delegate"`
	AccessLevel    SourceAccessLevel `json:"access_level" gorm:"not null"`
	ExpiresAt      *time.Time        `json:"expires_at"`
	RevokedAt      *time.Time        `json:"revoked_at"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// LocalResource records provenance only for the server's explicit local creation
// operation. Neither a client flag nor the source's platform proves provenance.
type LocalResource struct {
	OwnerUserID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	SourceID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ResourceType string    `gorm:"primaryKey"`
	ResourceID   string    `gorm:"primaryKey"`
}

// SourceAccessAudit intentionally contains identifiers, never clinical payloads,
// credentials or encryption material. Repository operations only append events.
type SourceAccessAudit struct {
	ID           uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey"`
	SourceID     uuid.UUID         `json:"source_id" gorm:"type:uuid;not null;index"`
	OwnerUserID  uuid.UUID         `json:"owner_user_id" gorm:"type:uuid;not null"`
	ActorUserID  uuid.UUID         `json:"actor_user_id" gorm:"type:uuid;not null"`
	GrantID      *uuid.UUID        `json:"grant_id,omitempty" gorm:"type:uuid"`
	Action       string            `json:"action"`
	AccessLevel  SourceAccessLevel `json:"access_level,omitempty"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	ResourceType string            `json:"resource_type,omitempty"`
	ResourceID   string            `json:"resource_id,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}
