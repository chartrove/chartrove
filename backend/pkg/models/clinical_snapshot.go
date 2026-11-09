package models

import (
	"time"

	"github.com/google/uuid"
)

const ClinicalSnapshotOwnerAcquisition = "owner-supplied-snapshot"

// ClinicalSnapshotRecord is append-only import evidence, not a provider grant or
// proof of native retrieval. Payload is the exact owner-supplied collection;
// it must never appear in metadata responses, ordinary logs or source grants.
// Source/account deletion explicitly removes these originals too.
type ClinicalSnapshotRecord struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OwnerUserID      uuid.UUID `json:"-" gorm:"type:uuid;not null;index"`
	ImportedByUserID uuid.UUID `json:"imported_by_user_id" gorm:"type:uuid;not null"`
	Acquisition      string    `json:"acquisition" gorm:"not null"`
	SourceID         uuid.UUID `json:"source_id" gorm:"type:uuid;not null;uniqueIndex:idx_clinical_source_digest"`
	SHA256           string    `json:"sha256" gorm:"not null;uniqueIndex:idx_clinical_source_digest"`
	FHIRBaseURL      string    `json:"fhir_base_url" gorm:"not null"`
	PatientID        string    `json:"patient_id" gorm:"not null"`
	ResourceCount    int       `json:"resource_count" gorm:"not null"`
	CreatedAt        time.Time `json:"imported_at" gorm:"not null"`
	Payload          []byte    `json:"-" gorm:"not null"`
}

type ClinicalSnapshotImportResult struct {
	SourceID        uuid.UUID `json:"source_id"`
	SnapshotID      uuid.UUID `json:"snapshot_id"`
	SHA256          string    `json:"sha256"`
	PatientID       string    `json:"patient_id"`
	Created         int       `json:"created"`
	Updated         int       `json:"updated"`
	Unchanged       int       `json:"unchanged"`
	AlreadyImported bool      `json:"already_imported"`
}
