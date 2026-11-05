package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrOwnedResourceAmbiguous = errors.New("resource ID is ambiguous; select its source")

type OwnedResourceMutationRepository interface {
	OwnedResourceUpdateRepository
	ResolveOwnedResource(context.Context, string, string, *uuid.UUID) (*models.ResourceBase, error)
	DeleteOwnedResource(context.Context, uuid.UUID, string, string) error
}

var _ OwnedResourceMutationRepository = (*GormRepository)(nil)

// ResolveOwnedResource never guesses between sources. A supplied source narrows
// selection but is not evidence of ownership; grants/admin roles are irrelevant.
// Return only target metadata, never clinical content or source credentials.
func (gr *GormRepository) ResolveOwnedResource(ctx context.Context, resourceType, resourceID string, source *uuid.UUID) (*models.ResourceBase, error) {
	actor, err := gr.sourceAccessActor(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSourceAccessDenied
	}
	if err != nil {
		return nil, err
	}
	table, err := dm.GetTableNameByResourceType(resourceType)
	if err != nil || resourceID == "" || (source != nil && *source == uuid.Nil) {
		return nil, fmt.Errorf("%w: invalid resource target", ErrSourceAccessInput)
	}
	db := gr.GormClient.WithContext(ctx)
	// Historical clinical source columns can be text while credential IDs are
	// PostgreSQL UUIDs. Compare canonical values without migrating clinical data.
	ownedSources := db.Model(&models.SourceCredential{}).Select("CAST(id AS TEXT)").Where("user_id = ? AND deleted_at IS NULL", actor.ID)
	liveActor := db.Model(&models.User{}).Select("1").Where("id = ? AND username = ? AND deleted_at IS NULL", actor.ID, actor.Username)
	q := db.Table(table).Select("id", "user_id", "source_id", "source_resource_type", "source_resource_id").
		Where("CAST(user_id AS TEXT) = ? AND source_resource_type = ? AND source_resource_id = ? AND deleted_at IS NULL", actor.ID.String(), resourceType, resourceID).
		Where("CAST(source_id AS TEXT) IN (?) AND EXISTS (?)", ownedSources, liveActor)
	if source != nil {
		q = q.Where("source_id = ?", *source)
	}
	var rows []models.ResourceBase
	if err := q.Limit(2).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrSourceAccessDenied
	}
	if len(rows) != 1 {
		return nil, ErrOwnedResourceAmbiguous
	}
	return &rows[0], nil
}

func (gr *GormRepository) DeleteOwnedResource(ctx context.Context, source uuid.UUID, resourceType, resourceID string) error {
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return err
	}
	table, err := dm.GetTableNameByResourceType(resourceType)
	if err != nil {
		return fmt.Errorf("%w: invalid resource type", ErrSourceAccessInput)
	}
	err = gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First transactional statement: acquire the writer reservation and
		// recheck current native/source/record ownership at the deletion boundary.
		ownedSource := tx.Model(&models.SourceCredential{}).Select("1").Where("id = ? AND user_id = ? AND deleted_at IS NULL", source, actor.ID)
		liveActor := tx.Model(&models.User{}).Select("1").Where("id = ? AND username = ? AND deleted_at IS NULL", actor.ID, actor.Username)
		res := tx.Table(table).Where("source_id = ? AND CAST(user_id AS TEXT) = ? AND source_resource_type = ? AND source_resource_id = ? AND deleted_at IS NULL", source, actor.ID.String(), resourceType, resourceID).
			Where("EXISTS (?) AND EXISTS (?)", ownedSource, liveActor).Delete(&models.ResourceBase{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrSourceAccessDenied
		}
		if err := tx.Where("owner_user_id = ? AND source_id = ? AND resource_type = ? AND resource_id = ?", actor.ID, source, resourceType, resourceID).Delete(&models.LocalResource{}).Error; err != nil {
			return err
		}
		return appendSourceAudit(tx, source, actor.ID, actor.ID, "delete_owned", nil, resourceType, resourceID)
	})
	if err == nil && gr.EventBus != nil {
		if publishErr := gr.EventBus.PublishMessage(models.NewEventSourceSync(actor.ID.String(), source.String(), resourceType, resourceID)); publishErr != nil && gr.Logger != nil {
			gr.Logger.Warn("Could not publish owner resource deletion notification")
		}
	}
	return err
}
