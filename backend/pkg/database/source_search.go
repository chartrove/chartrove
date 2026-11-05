package database

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Optional capability leaves the existing owner and delegation API contracts intact.
type SourceSearchRepository interface {
	SearchSourceResources(context.Context, uuid.UUID, string, string, int) ([]models.SourceSearchHit, error)
}

// Optional exact-ID capability keeps existing literal clients on their original API.
type SourceExactIDSearchRepository interface {
	SearchSourceResourceID(context.Context, uuid.UUID, string, string, int) ([]models.SourceSearchHit, error)
}

type SourceMetadataRepository interface {
	ListSourceMetadata(context.Context, uuid.UUID, string, int) ([]models.SourceSearchHit, error)
}

var _ SourceSearchRepository = (*GormRepository)(nil)
var _ SourceExactIDSearchRepository = (*GormRepository)(nil)

func (gr *GormRepository) SearchSourceResources(ctx context.Context, source uuid.UUID, resourceType, text string, page int) ([]models.SourceSearchHit, error) {
	text = strings.TrimSpace(text)
	if !utf8.ValidString(text) || text == "" || utf8.RuneCountInString(text) > 128 {
		return nil, fmt.Errorf("%w: invalid search text or page", ErrSourceAccessInput)
	}
	return gr.sourceMetadata(ctx, source, resourceType, page, &text, false)
}

func (gr *GormRepository) SearchSourceResourceID(ctx context.Context, source uuid.UUID, resourceType, text string, page int) ([]models.SourceSearchHit, error) {
	if !utf8.ValidString(text) || strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > 128 || strings.ContainsRune(text, 0) {
		return nil, fmt.Errorf("%w: invalid resource ID", ErrSourceAccessInput)
	}
	return gr.sourceMetadata(ctx, source, resourceType, page, &text, true)
}

func (gr *GormRepository) ListSourceMetadata(ctx context.Context, source uuid.UUID, resourceType string, page int) ([]models.SourceSearchHit, error) {
	return gr.sourceMetadata(ctx, source, resourceType, page, nil, false)
}

func (gr *GormRepository) sourceMetadata(ctx context.Context, source uuid.UUID, resourceType string, page int, text *string, exactID bool) ([]models.SourceSearchHit, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if page < 0 || page > 100000 {
		return nil, ErrSourceAccessInput
	}
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
	// Treat wildcards as literal text. Only fixed, allowlisted identifiers enter
	// SQL; both patterns and all ownership/grant values remain bound parameters.
	rows := []models.SourceSearchHit{}
	query := gr.sourceResourceQuery(ctx, actor.ID, source, table, false).
		Session(&gorm.Session{Logger: privateSQLLogger{gr.GormClient.Logger}}).
		Select("source_id, source_resource_type, source_resource_id, sort_title, sort_date").
		Where("source_resource_type = ?", resourceType)
	if text != nil && exactID {
		// Keep ordinary equality for the existing composite index, then compare
		// bytes so a legacy/custom case-insensitive collation cannot merge IDs.
		query = query.Where("source_resource_id = ?", *text)
		switch gr.GormClient.Dialector.Name() {
		case "sqlite":
			query = query.Where("CAST(source_resource_id AS BLOB) = CAST(? AS BLOB)", *text)
		case "postgres":
			query = query.Where("convert_to(source_resource_id, 'UTF8') = convert_to(?, 'UTF8')", *text)
		default:
			return nil, fmt.Errorf("exact resource ID search unavailable")
		}
	} else if text != nil {
		pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*text) + "%"
		condition := "(LOWER(sort_title) LIKE LOWER(?) ESCAPE '!' OR LOWER(source_resource_id) LIKE LOWER(?) ESCAPE '!')"
		if resourceType == "DocumentReference" {
			// Only this server's reviewed-document projection is searched. JSON
			// expansion compares decoded strings, preserving literal quotes,
			// Unicode and backslashes instead of matching JSON syntax/escapes.
			var values string
			switch gr.GormClient.Dialector.Name() {
			case "sqlite":
				values = "json_each(CASE WHEN source_uri = ? THEN COALESCE(NULLIF(text, ''), '[]') ELSE '[]' END) AS document_text"
			case "postgres":
				values = "jsonb_array_elements_text((CASE WHEN source_uri = ? THEN COALESCE(NULLIF(CAST(text AS TEXT), ''), '[]') ELSE '[]' END)::jsonb) AS document_text(value)"
			default:
				return nil, fmt.Errorf("document transcript search unavailable")
			}
			condition = "(" + condition + " OR EXISTS (SELECT 1 FROM " + values + " WHERE LOWER(document_text.value) LIKE LOWER(?) ESCAPE '!'))"
			query = query.Where(condition, pattern, pattern, documentImportOrigin, pattern)
		} else {
			query = query.Where(condition, pattern, pattern)
		}
	}
	err = query.Order("source_resource_id").Limit(50).Offset(page * 50).Find(&rows).Error
	return rows, err
}
