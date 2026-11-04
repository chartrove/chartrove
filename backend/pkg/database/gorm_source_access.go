package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (gr *GormRepository) sourceAccessActor(ctx context.Context) (*models.User, error) {
	actor, err := gr.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if actor.DeletedAt != nil {
		return nil, ErrSourceAccessDenied
	}
	return actor, nil
}

func activeSourceGrant(db *gorm.DB, actor uuid.UUID, source uuid.UUID, edit bool) *gorm.DB {
	q := db.Model(&models.SourceGrant{}).Select("1").Where("source_id = ? AND delegate_user_id = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > ?)", source, actor, time.Now().UTC())
	q = q.Where("owner_user_id = (SELECT user_id FROM source_credentials WHERE id = ? AND deleted_at IS NULL)", source)
	if edit {
		return q.Where("access_level = ?", models.SourceAccessEdit)
	}
	return q.Where("access_level IN ?", []models.SourceAccessLevel{models.SourceAccessView, models.SourceAccessEdit})
}

func (gr *GormRepository) ownedAccessSource(ctx context.Context, source uuid.UUID) (*models.SourceCredential, error) {
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	var s models.SourceCredential
	err = gr.GormClient.WithContext(ctx).Select("id", "user_id", "display", "platform_type").Where("id = ? AND user_id = ? AND deleted_at IS NULL", source, actor.ID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSourceAccessDenied
	}
	return &s, err
}

func appendSourceAudit(db *gorm.DB, source, owner, actor uuid.UUID, action string, grant *models.SourceGrant, resourceType, resourceID string) error {
	e := models.SourceAccessAudit{ID: uuid.New(), SourceID: source, OwnerUserID: owner, ActorUserID: actor, Action: action, ResourceType: resourceType, ResourceID: resourceID}
	if grant != nil {
		e.GrantID = &grant.ID
		e.AccessLevel = grant.AccessLevel
		e.ExpiresAt = grant.ExpiresAt
	}
	return db.Create(&e).Error
}

func (gr *GormRepository) SetSourceGrant(ctx context.Context, source uuid.UUID, username string, level models.SourceAccessLevel, expires *time.Time) (*models.SourceGrant, error) {
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return nil, err
	}
	if level != models.SourceAccessView && level != models.SourceAccessEdit {
		return nil, fmt.Errorf("%w: invalid access level", ErrSourceAccessInput)
	}
	if expires != nil && !expires.After(time.Now()) {
		return nil, fmt.Errorf("%w: expiration must be in the future", ErrSourceAccessInput)
	}
	if expires != nil {
		// SQLite compares persisted timestamps lexically. Match the UTC clock
		// used by authorization without mutating the caller's timestamp.
		utcExpiration := expires.UTC()
		expires = &utcExpiration
	}
	var delegate models.User
	err = gr.GormClient.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&delegate).Error
	if err != nil {
		return nil, fmt.Errorf("%w: delegate unavailable", ErrSourceAccessInput)
	}
	if delegate.ID == s.UserID {
		return nil, fmt.Errorf("%w: owners already have full access", ErrSourceAccessInput)
	}
	grant := models.SourceGrant{ID: uuid.New(), OwnerUserID: s.UserID, SourceID: source, DelegateUserID: delegate.ID, AccessLevel: level, ExpiresAt: expires}
	err = gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "owner_user_id"}, {Name: "source_id"}, {Name: "delegate_user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"access_level", "expires_at", "revoked_at", "updated_at"})}).Create(&grant).Error; err != nil {
			return err
		}
		grant = models.SourceGrant{}
		if err := tx.Where("owner_user_id = ? AND source_id = ? AND delegate_user_id = ?", s.UserID, source, delegate.ID).First(&grant).Error; err != nil {
			return err
		}
		return appendSourceAudit(tx, source, s.UserID, s.UserID, "grant", &grant, "", "")
	})
	grant.DelegateUsername = delegate.Username
	return &grant, err
}

func (gr *GormRepository) RevokeSourceGrant(ctx context.Context, source, id uuid.UUID) error {
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return err
	}
	return gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var grant models.SourceGrant
		if err := tx.Where("id = ? AND owner_user_id = ? AND source_id = ?", id, s.UserID, source).First(&grant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSourceAccessDenied
			}
			return err
		}
		now := time.Now().UTC()
		if err := tx.Model(&models.SourceGrant{}).Where("id = ? AND owner_user_id = ? AND source_id = ?", id, s.UserID, source).Updates(map[string]interface{}{"revoked_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		return appendSourceAudit(tx, source, s.UserID, s.UserID, "revoke", &grant, "", "")
	})
}

func (gr *GormRepository) ListSourceGrants(ctx context.Context, source uuid.UUID) ([]models.SourceGrant, error) {
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return nil, err
	}
	rows := []models.SourceGrant{}
	err = gr.GormClient.WithContext(ctx).Table("source_grants AS g").Select("g.*, u.username AS delegate_username").Joins("JOIN users AS u ON u.id = g.delegate_user_id AND u.deleted_at IS NULL").Where("g.source_id = ? AND g.owner_user_id = ?", source, s.UserID).Order("g.created_at, g.id").Find(&rows).Error
	return rows, err
}

func (gr *GormRepository) ListSourceAccessAudit(ctx context.Context, source uuid.UUID) ([]models.SourceAccessAudit, error) {
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return nil, err
	}
	rows := []models.SourceAccessAudit{}
	err = gr.GormClient.WithContext(ctx).Where("source_id = ? AND owner_user_id = ?", source, s.UserID).Order("created_at DESC, id").Limit(200).Find(&rows).Error
	return rows, err
}

func (gr *GormRepository) sharedSource(ctx context.Context, source uuid.UUID) (*models.SharedSource, error) {
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	db := gr.GormClient.WithContext(ctx)
	var s models.SharedSource
	err = db.Table("source_credentials").Select("id, user_id AS owner_user_id, display").Where("id = ? AND deleted_at IS NULL", source).
		Where("user_id = ? OR EXISTS (?)", actor.ID, activeSourceGrant(db, actor.ID, source, false)).Take(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSourceAccessDenied
	}
	if err != nil {
		return nil, err
	}
	if s.OwnerUserID == actor.ID {
		s.AccessLevel = models.SourceAccessEdit
	} else {
		var grant models.SourceGrant
		// Repeat the live authorization predicate rather than returning a cached grant.
		err = db.Where("source_id = ? AND owner_user_id = ? AND delegate_user_id = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > ?) AND access_level IN ?", source, s.OwnerUserID, actor.ID, time.Now().UTC(), []models.SourceAccessLevel{models.SourceAccessView, models.SourceAccessEdit}).First(&grant).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSourceAccessDenied
		}
		if err != nil {
			return nil, err
		}
		s.AccessLevel = grant.AccessLevel
	}
	return &s, nil
}

func (gr *GormRepository) ListSharedSources(ctx context.Context) ([]models.SharedSource, error) {
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	rows := []models.SharedSource{}
	err = gr.GormClient.WithContext(ctx).Table("source_credentials AS s").
		Select("s.id, s.user_id AS owner_user_id, s.display, g.access_level").
		Joins("JOIN source_grants AS g ON g.source_id = s.id AND g.owner_user_id = s.user_id").
		Where("g.delegate_user_id = ? AND g.revoked_at IS NULL AND (g.expires_at IS NULL OR g.expires_at > ?) AND g.access_level IN ? AND s.deleted_at IS NULL", actor.ID, time.Now().UTC(), []models.SourceAccessLevel{models.SourceAccessView, models.SourceAccessEdit}).Order("s.id").Scan(&rows).Error
	return rows, err
}

// Each resource statement repeats source ownership AND the current grant. The
// earlier metadata check is for a consistent error, never the authorization proof.
func (gr *GormRepository) sourceResourceQuery(ctx context.Context, actor uuid.UUID, source uuid.UUID, table string, edit bool) *gorm.DB {
	db := gr.GormClient.WithContext(ctx)
	sources := db.Table("source_credentials").Select("user_id").Where("id = ? AND deleted_at IS NULL", source).
		Where("user_id = ? OR EXISTS (?)", actor, activeSourceGrant(db, actor, source, edit))
	return db.Table(table).Where("source_id = ? AND user_id IN (?) AND deleted_at IS NULL", source, sources)
}

func (gr *GormRepository) ListSharedResources(ctx context.Context, source uuid.UUID, resourceType string, page int) ([]models.ResourceBase, error) {
	if _, err := gr.sharedSource(ctx, source); err != nil {
		return nil, err
	}
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	table, err := dm.GetTableNameByResourceType(resourceType)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid resource type", ErrSourceAccessInput)
	}
	if page < 0 || page > 100000 {
		return nil, fmt.Errorf("%w: invalid page", ErrSourceAccessInput)
	}
	rows := []models.ResourceBase{}
	err = gr.sourceResourceQuery(ctx, actor.ID, source, table, false).Where("source_resource_type = ?", resourceType).Order("source_resource_id").Limit(50).Offset(page * 50).Find(&rows).Error
	return rows, err
}

func (gr *GormRepository) GetSharedResource(ctx context.Context, source uuid.UUID, resourceType, resourceID string) (*models.SharedResource, error) {
	s, err := gr.sharedSource(ctx, source)
	if err != nil {
		return nil, err
	}
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	table, err := dm.GetTableNameByResourceType(resourceType)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid resource type", ErrSourceAccessInput)
	}
	var row models.ResourceBase
	err = gr.sourceResourceQuery(ctx, actor.ID, source, table, false).Where("source_resource_type = ? AND source_resource_id = ?", resourceType, resourceID).First(&row).Error
	if err != nil {
		return nil, err
	}
	var localCount int64
	err = gr.GormClient.WithContext(ctx).Model(&models.LocalResource{}).Where("owner_user_id = ? AND source_id = ? AND resource_type = ? AND resource_id = ?", row.UserID, source, resourceType, resourceID).Count(&localCount).Error
	if err != nil {
		return nil, err
	}
	return &models.SharedResource{ResourceBase: row, Editable: localCount == 1 && s.AccessLevel == models.SourceAccessEdit}, nil
}

func (gr *GormRepository) GetSharedSourceSummary(ctx context.Context, source uuid.UUID) (*models.SharedSourceSummary, error) {
	s, err := gr.sharedSource(ctx, source)
	if err != nil {
		return nil, err
	}
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	result := &models.SharedSourceSummary{Source: *s, Counts: map[string]int64{}}
	for _, rt := range dm.GetAllowedResourceTypes() {
		table, err := dm.GetTableNameByResourceType(rt)
		if err != nil {
			return nil, err
		}
		var count int64
		if err := gr.sourceResourceQuery(ctx, actor.ID, source, table, false).Where("source_resource_type = ?", rt).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			result.Counts[rt] = count
		}
	}
	return result, nil
}

func parseLocalResource(raw json.RawMessage, resourceType, resourceID string) (dm.IFhirResourceModel, error) {
	var identity struct {
		ResourceType string `json:"resourceType"`
		ID           string `json:"id"`
	}
	if err := json.Unmarshal(raw, &identity); err != nil {
		return nil, fmt.Errorf("%w: invalid FHIR JSON", ErrSourceAccessInput)
	}
	if identity.ResourceType != resourceType || identity.ID != resourceID {
		return nil, fmt.Errorf("%w: resource identity cannot change", ErrSourceAccessInput)
	}
	model, err := dm.NewFhirResourceModelByType(resourceType)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid resource type", ErrSourceAccessInput)
	}
	if err := model.PopulateAndExtractSearchParameters(raw); err != nil {
		return nil, fmt.Errorf("%w: invalid FHIR resource", ErrSourceAccessInput)
	}
	return model, nil
}

// CreateLocalResource is an explicit owner-only creation operation. It assigns a
// fresh identity and never promotes an existing imported record to editable.
func (gr *GormRepository) CreateLocalResource(ctx context.Context, source uuid.UUID, raw json.RawMessage) (*models.ResourceBase, error) {
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return nil, err
	}
	if s.PlatformType != sp.PlatformTypeFasten {
		return nil, ErrSourceAccessDenied
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
		return nil, fmt.Errorf("%w: invalid FHIR JSON", ErrSourceAccessInput)
	}
	rt, ok := payload["resourceType"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: resourceType is required", ErrSourceAccessInput)
	}
	id := uuid.NewString()
	payload["id"] = id
	raw, err = json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	model, err := parseLocalResource(raw, rt, id)
	if err != nil {
		return nil, err
	}
	model.SetOriginBase(models.OriginBase{ModelBase: models.ModelBase{ID: uuid.New()}, UserID: s.UserID, SourceID: source, SourceResourceType: rt, SourceResourceID: id})
	err = gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(model).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.LocalResource{OwnerUserID: s.UserID, SourceID: source, ResourceType: rt, ResourceID: id}).Error; err != nil {
			return err
		}
		return appendSourceAudit(tx, source, s.UserID, s.UserID, "create_local", nil, rt, id)
	})
	if err != nil {
		return nil, err
	}
	stored, err := gr.GetSharedResource(ctx, source, rt, id)
	if err != nil {
		return nil, err
	}
	return &stored.ResourceBase, nil
}

func (gr *GormRepository) UpdateSharedResource(ctx context.Context, source uuid.UUID, resourceType, resourceID string, raw json.RawMessage) error {
	s, err := gr.sharedSource(ctx, source)
	if err != nil {
		return err
	}
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return err
	}
	model, err := parseLocalResource(raw, resourceType, resourceID)
	if err != nil {
		return err
	}
	table, err := dm.GetTableNameByResourceType(resourceType)
	if err != nil {
		return err
	}
	return gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := *gr
		r.GormClient = tx
		var existing models.ResourceBase
		authorized := func() *gorm.DB {
			q := r.sourceResourceQuery(ctx, actor.ID, source, table, true).Where("source_resource_type = ? AND source_resource_id = ?", resourceType, resourceID)
			// Provenance is not inferred from platform or supplied by the client.
			return q.Where("EXISTS (?)", tx.Model(&models.LocalResource{}).Select("1").Where("owner_user_id = ? AND source_id = ? AND resource_type = ? AND resource_id = ?", s.OwnerUserID, source, resourceType, resourceID))
		}
		if err := authorized().First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSourceAccessDenied
			}
			return err
		}
		model.SetOriginBase(existing.OriginBase)
		model.SetSortTitle(existing.SortTitle)
		model.SetSortDate(existing.SortDate)
		model.SetSourceUri(existing.SourceUri)
		// Select all derived columns so removed values are cleared as well. Never
		// Save/upsert: updates must affect exactly the authorized existing row.
		res := authorized().Where("id = ?", existing.ID).Select("*").Omit("id", "user_id", "source_id", "source_resource_type", "source_resource_id", "created_at", clause.Associations).Updates(model)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrSourceAccessDenied
		}
		return appendSourceAudit(tx, source, existing.UserID, actor.ID, "edit_local", nil, resourceType, resourceID)
	})
}
