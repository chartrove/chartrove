package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OwnedResourceUpdate retains the legacy complete-resource replacement contract.
// Omitted presentation fields preserve their current database values; an explicit
// empty/invalid legacy sort date clears it, as before. Source provenance is never
// supplied by the caller or changed by an edit.
type OwnedResourceUpdate struct {
	ResourceRaw json.RawMessage
	SortTitle   *string
	SortDate    *time.Time
	SetSortDate bool
}

type OwnedResourceUpdateRepository interface {
	UpdateOwnedResource(context.Context, uuid.UUID, string, string, OwnedResourceUpdate) error
}

var _ OwnedResourceUpdateRepository = (*GormRepository)(nil)

// UpdateOwnedResource is an owner-only edit, not SDK ingestion or an upsert. It
// preserves existing local-edit eligibility without promoting imported records.
func (gr *GormRepository) UpdateOwnedResource(ctx context.Context, source uuid.UUID, resourceType, resourceID string, edit OwnedResourceUpdate) error {
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return err
	}
	table, err := dm.GetTableNameByResourceType(resourceType)
	if err != nil {
		return fmt.Errorf("%w: invalid resource type", ErrSourceAccessInput)
	}
	model, err := dm.NewFhirResourceModelByType(resourceType)
	if err != nil {
		return err
	}
	if !json.Valid(edit.ResourceRaw) {
		return ErrSourceAccessInput
	}
	if err := model.PopulateAndExtractSearchParameters(edit.ResourceRaw); err != nil {
		return fmt.Errorf("%w: invalid FHIR resource", ErrSourceAccessInput)
	}
	model.SetSortTitle(edit.SortTitle)
	model.SetSortDate(edit.SortDate)
	omit := []string{"id", "user_id", "source_id", "source_resource_type", "source_resource_id", "created_at", "deleted_at", "source_uri", clause.Associations}
	if edit.SortTitle == nil {
		omit = append(omit, "sort_title")
	}
	if !edit.SetSortDate {
		omit = append(omit, "sort_date")
	}
	err = gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// The update is the first transactional statement, acquiring SQLite's
		// writer reservation without a stale read snapshot. Recheck source and
		// native UUID ownership at the write boundary, independent of grants/role.
		ownedSource := tx.Model(&models.SourceCredential{}).Select("1").Where("id = ? AND user_id = ? AND deleted_at IS NULL", source, actor.ID)
		liveActor := tx.Model(&models.User{}).Select("1").Where("id = ? AND username = ? AND deleted_at IS NULL", actor.ID, actor.Username)
		q := tx.Table(table).Where("source_id = ? AND CAST(user_id AS TEXT) = ? AND source_resource_type = ? AND source_resource_id = ? AND deleted_at IS NULL", source, actor.ID.String(), resourceType, resourceID).
			Where("EXISTS (?) AND EXISTS (?)", ownedSource, liveActor)
		res := q.Select("*").Omit(omit...).Updates(model)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrSourceAccessDenied
		}
		return appendSourceAudit(tx, source, actor.ID, actor.ID, "edit_owned", nil, resourceType, resourceID)
	})
	if err == nil && gr.EventBus != nil {
		if publishErr := gr.EventBus.PublishMessage(models.NewEventSourceSync(actor.ID.String(), source.String(), resourceType, resourceID)); publishErr != nil && gr.Logger != nil {
			gr.Logger.Warn("Could not publish owner resource update notification")
		}
	}
	return err
}
