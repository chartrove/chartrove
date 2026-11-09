// Package _20261109120000 freezes the original clinical import evidence schema.
package _20261109120000

import (
	"time"

	"github.com/google/uuid"
)

type ClinicalSnapshotRecord struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerUserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	ImportedByUserID uuid.UUID `gorm:"type:uuid;not null"`
	Acquisition      string    `gorm:"not null"`
	SourceID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_clinical_source_digest"`
	SHA256           string    `gorm:"not null;uniqueIndex:idx_clinical_source_digest"`
	FHIRBaseURL      string    `gorm:"not null"`
	PatientID        string    `gorm:"not null"`
	ResourceCount    int       `gorm:"not null"`
	CreatedAt        time.Time `gorm:"not null"`
	Payload          []byte    `gorm:"not null"`
}
