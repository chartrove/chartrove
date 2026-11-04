package models

import (
	"github.com/google/uuid"
	"time"
)

// SourceSearchHit is navigation metadata, never an indexed clinical payload.
// Opening a hit goes through the existing live-authorized detail operation.
type SourceSearchHit struct {
	SourceID           uuid.UUID  `json:"source_id"`
	SourceResourceType string     `json:"source_resource_type"`
	SourceResourceID   string     `json:"source_resource_id"`
	SortTitle          *string    `json:"sort_title"`
	SortDate           *time.Time `json:"sort_date"`
}
