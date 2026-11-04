// Frozen additive schema; no existing users or clinical rows are rewritten.
package _20261104210000

import (
	"github.com/google/uuid"
	"time"
)

func (OIDCIdentity) TableName() string { return "oidc_identities" }
func (OIDCSession) TableName() string  { return "oidc_sessions" }
func (OIDCAudit) TableName() string    { return "oidc_audits" }

// OIDCIdentity binds a verified issuer/subject pair to an immutable native ID.
// Email, provider roles and group claims never participate in this binding.
type OIDCIdentity struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `json:"-" gorm:"type:uuid;not null;index"`
	Issuer    string     `json:"issuer" gorm:"not null;uniqueIndex:oidc_subject"`
	Subject   string     `json:"-" gorm:"not null;uniqueIndex:oidc_subject"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// Only native session identifiers are persisted; no provider tokens or secrets.
type OIDCSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	IdentityID uuid.UUID `gorm:"type:uuid;not null;index"`
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

type OIDCAudit struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `json:"-" gorm:"type:uuid;not null;index"`
	IdentityID uuid.UUID `json:"identity_id" gorm:"type:uuid;not null"`
	Action     string    `json:"action"`
	CreatedAt  time.Time `json:"created_at"`
}
