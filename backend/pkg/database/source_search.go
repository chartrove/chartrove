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
	"gorm.io/gorm/logger"
)

// Optional capability leaves the existing owner and delegation API contracts intact.
type SourceSearchRepository interface {
	SearchSourceResources(context.Context, uuid.UUID, string, string, int) ([]models.SourceSearchHit, error)
}

var _ SourceSearchRepository = (*GormRepository)(nil)

// Preserve query structure, timing and row-count diagnostics without putting
// potentially private search text into GORM's debug/slow/error SQL traces.
type sourceSearchLogger struct{ logger.Interface }

func (sourceSearchLogger) ParamsFilter(_ context.Context, sql string, _ ...interface{}) (string, []interface{}) {
	return sql, nil
}

func (gr *GormRepository) SearchSourceResources(ctx context.Context, source uuid.UUID, resourceType, text string, page int) ([]models.SourceSearchHit, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	text = strings.TrimSpace(text)
	if !utf8.ValidString(text) || text == "" || utf8.RuneCountInString(text) > 128 || page < 0 || page > 100000 {
		return nil, fmt.Errorf("%w: invalid search text or page", ErrSourceAccessInput)
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
	pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(text) + "%"
	rows := []models.SourceSearchHit{}
	err = gr.sourceResourceQuery(ctx, actor.ID, source, table, false).
		Session(&gorm.Session{Logger: sourceSearchLogger{gr.GormClient.Logger}}).
		Select("source_id, source_resource_type, source_resource_id, sort_title, sort_date").
		Where("source_resource_type = ?", resourceType).
		Where("(LOWER(sort_title) LIKE LOWER(?) ESCAPE '!' OR LOWER(source_resource_id) LIKE LOWER(?) ESCAPE '!')", pattern, pattern).
		Order("source_resource_id").Limit(50).Offset(page * 50).Find(&rows).Error
	return rows, err
}
