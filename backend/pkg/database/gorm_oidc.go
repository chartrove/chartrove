package database

import (
	"context"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func oidcAudit(tx *gorm.DB, user, identity uuid.UUID, action string) error {
	return tx.Create(&models.OIDCAudit{ID: uuid.New(), UserID: user, IdentityID: identity, Action: action}).Error
}

func (gr *GormRepository) CompleteOIDC(ctx context.Context, issuer, subject string, linkUser uuid.UUID, passwordHash string, authorizedAt, expires time.Time) (*models.User, *models.OIDCSession, error) {
	if issuer == "" || subject == "" || !expires.After(time.Now()) {
		return nil, nil, ErrOIDCDenied
	}
	var user models.User
	var session models.OIDCSession
	err := gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if linkUser != uuid.Nil {
			// First write reserves SQLite's writer before the authorization reads.
			identity := models.OIDCIdentity{ID: uuid.New(), UserID: linkUser, Issuer: issuer, Subject: subject}
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&identity)
			if result.Error != nil {
				return result.Error
			}
			if passwordHash == "" || authorizedAt.IsZero() || authorizedAt.After(time.Now()) {
				return ErrOIDCDenied
			}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", linkUser).First(&user).Error; err != nil {
				return ErrOIDCDenied
			}
			if user.Password != passwordHash {
				return ErrOIDCDenied
			}
			// A subject already bound to another native ID cannot be reassigned.
			result = tx.Model(&models.OIDCIdentity{}).Where("issuer = ? AND subject = ? AND user_id = ? AND (revoked_at IS NULL OR revoked_at < ?)", issuer, subject, linkUser, authorizedAt.UTC()).Update("revoked_at", nil)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ErrOIDCDenied
			}
		}
		var identity models.OIDCIdentity
		// INSERT SELECT repeats the active binding and native account predicates at
		// the write itself, so revocation cannot race a stale read into a session.
		session = models.OIDCSession{ID: uuid.New(), ExpiresAt: expires.UTC(), CreatedAt: time.Now().UTC()}
		result := tx.Exec(`INSERT INTO oidc_sessions (id, user_id, identity_id, expires_at, created_at)
			SELECT ?, i.user_id, i.id, ?, ? FROM oidc_identities i JOIN users u ON u.id = i.user_id
			WHERE i.issuer = ? AND i.subject = ? AND i.revoked_at IS NULL AND u.deleted_at IS NULL`, session.ID, session.ExpiresAt, session.CreatedAt, issuer, subject)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrOIDCDenied
		}
		if err := tx.Where("issuer = ? AND subject = ? AND revoked_at IS NULL", issuer, subject).First(&identity).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND deleted_at IS NULL", identity.UserID).First(&user).Error; err != nil {
			return err
		}
		session.UserID = user.ID
		session.IdentityID = identity.ID
		if linkUser != uuid.Nil {
			if err := oidcAudit(tx, user.ID, identity.ID, "link"); err != nil {
				return err
			}
		}
		return oidcAudit(tx, user.ID, identity.ID, "login")
	})
	return &user, &session, err
}

func (gr *GormRepository) ValidateOIDCSession(ctx context.Context, id, user uuid.UUID, issuer string) error {
	var count int64
	err := gr.GormClient.WithContext(ctx).Table("oidc_sessions s").Joins("JOIN oidc_identities i ON i.id = s.identity_id AND i.user_id = s.user_id").Joins("JOIN users u ON u.id = s.user_id").Where("s.id = ? AND s.user_id = ? AND s.revoked_at IS NULL AND s.expires_at > ? AND i.revoked_at IS NULL AND i.issuer = ? AND u.deleted_at IS NULL", id, user, time.Now().UTC(), issuer).Count(&count).Error
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrOIDCDenied
	}
	return nil
}

func (gr *GormRepository) RevokeOIDCSession(ctx context.Context, user, id uuid.UUID) error {
	return gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.OIDCSession{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, user).Update("revoked_at", time.Now().UTC())
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrOIDCDenied
		}
		var session models.OIDCSession
		if err := tx.Where("id = ? AND user_id = ?", id, user).First(&session).Error; err != nil {
			return err
		}
		return oidcAudit(tx, user, session.IdentityID, "logout")
	})
}

func (gr *GormRepository) RevokeOIDCIdentity(ctx context.Context, user, id uuid.UUID) error {
	return gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		result := tx.Model(&models.OIDCIdentity{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, user).Update("revoked_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrOIDCDenied
		}
		if err := tx.Model(&models.OIDCSession{}).Where("identity_id = ? AND user_id = ? AND revoked_at IS NULL", id, user).Update("revoked_at", now).Error; err != nil {
			return err
		}
		return oidcAudit(tx, user, id, "unlink")
	})
}

func (gr *GormRepository) ListOIDCIdentities(ctx context.Context, user uuid.UUID) ([]models.OIDCIdentity, error) {
	var rows []models.OIDCIdentity
	err := gr.GormClient.WithContext(ctx).Where("user_id = ?", user).Order("created_at DESC").Find(&rows).Error
	return rows, err
}
func (gr *GormRepository) ListOIDCAudit(ctx context.Context, user uuid.UUID) ([]models.OIDCAudit, error) {
	var rows []models.OIDCAudit
	err := gr.GormClient.WithContext(ctx).Where("user_id = ?", user).Order("created_at DESC").Limit(200).Find(&rows).Error
	return rows, err
}
