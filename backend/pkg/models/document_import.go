package models

import "github.com/google/uuid"

const DocumentMaxBytes = 8 << 20
const DocumentMaxTranscriptBytes = 128 << 10

type ReviewedDocument struct {
	Filename   string
	Title      string
	Original   []byte
	Transcript string
	Reviewed   bool
}

type DocumentImportSource struct {
	ID      uuid.UUID `json:"id"`
	Display string    `json:"display"`
}

type DocumentImportResult struct {
	SourceID       uuid.UUID `json:"source_id"`
	ResourceType   string    `json:"source_resource_type"`
	ResourceID     string    `json:"source_resource_id"`
	OriginalSHA256 string    `json:"original_sha256"`
	Created        bool      `json:"created"`
}
