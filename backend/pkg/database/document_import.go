package database

// Retains the document capture/review/provenance goal of TheoAtTechStack's
// Fasten #609 / Chartrove #212. OCR and review happen locally; no terminology
// service, inferred clinical concepts or externally editable import is added.
import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DocumentImportRepository interface {
	ListDocumentImportSources(context.Context) ([]models.DocumentImportSource, error)
	AuthorizeDocumentImportSource(context.Context, uuid.UUID) error
	ImportReviewedDocument(context.Context, uuid.UUID, models.ReviewedDocument) (*models.DocumentImportResult, error)
}

var _ DocumentImportRepository = (*GormRepository)(nil)

const documentImportOrigin = "urn:chartrove:document-import:user-reviewed-local-extraction"

func (gr *GormRepository) ListDocumentImportSources(ctx context.Context) ([]models.DocumentImportSource, error) {
	actor, err := gr.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	rows := []models.DocumentImportSource{}
	err = gr.GormClient.WithContext(ctx).Model(&models.SourceCredential{}).Select("id, display").Where("user_id = ? AND platform_type = ? AND deleted_at IS NULL", actor.ID, sp.PlatformTypeFasten).Order("id").Find(&rows).Error
	return rows, err
}

func (gr *GormRepository) AuthorizeDocumentImportSource(ctx context.Context, source uuid.UUID) error {
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return err
	}
	if s.PlatformType != sp.PlatformTypeFasten {
		return ErrSourceAccessDenied
	}
	return nil
}

func (gr *GormRepository) ImportReviewedDocument(ctx context.Context, source uuid.UUID, input models.ReviewedDocument) (*models.DocumentImportResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	s, err := gr.ownedAccessSource(ctx, source)
	if err != nil {
		return nil, err
	}
	if s.PlatformType != sp.PlatformTypeFasten {
		return nil, ErrSourceAccessDenied
	}
	input.Title = strings.TrimSpace(input.Title)
	if !input.Reviewed || len(input.Original) == 0 || len(input.Original) > models.DocumentMaxBytes || !utf8.ValidString(input.Transcript) || strings.TrimSpace(input.Transcript) == "" || len(input.Transcript) > models.DocumentMaxTranscriptBytes || !utf8.ValidString(input.Title) || input.Title == "" || utf8.RuneCountInString(input.Title) > 256 || !utf8.ValidString(input.Filename) || input.Filename == "" || utf8.RuneCountInString(input.Filename) > 200 {
		return nil, ErrSourceAccessInput
	}
	mime := http.DetectContentType(input.Original)
	if mime != "image/png" && mime != "image/jpeg" && mime != "application/pdf" {
		return nil, ErrSourceAccessInput
	}
	digest := sha256.Sum256(input.Original)
	originalHash := hex.EncodeToString(digest[:])
	textDigest := sha256.Sum256([]byte(input.Transcript))
	fingerprint, _ := json.Marshal([]string{s.UserID.String(), source.String(), originalHash, hex.EncodeToString(textDigest[:]), input.Title, input.Filename})
	idDigest := sha256.Sum256(fingerprint)
	// Retry-safe server identity; the hash is never an authorization credential.
	id := "doc-" + hex.EncodeToString(idDigest[:])[:60]
	now := time.Now().UTC()
	attachment := func(contentType, title string, size int) map[string]interface{} {
		return map[string]interface{}{"attachment": map[string]interface{}{"contentType": contentType, "title": title, "size": size}}
	}
	content := []map[string]interface{}{attachment(mime, input.Filename, len(input.Original)), attachment("text/plain", input.Title+" — reviewed transcript", len([]byte(input.Transcript)))}
	document := map[string]interface{}{
		"resourceType": "DocumentReference", "id": id, "status": "current", "date": now.Format(time.RFC3339), "description": input.Title,
		"meta":       map[string]string{"source": documentImportOrigin},
		"identifier": []map[string]string{{"system": "urn:chartrove:document:sha256", "value": originalHash}},
		"content":    content,
	}
	// Derive search fields from the same controlled metadata without feeding
	// opaque binary attachments through the FHIR JavaScript extractor.
	metadata, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	model, err := parseLocalResource(metadata, "DocumentReference", id)
	if err != nil {
		return nil, err
	}
	content[0]["attachment"].(map[string]interface{})["data"] = base64.StdEncoding.EncodeToString(input.Original)
	content[1]["attachment"].(map[string]interface{})["data"] = base64.StdEncoding.EncodeToString([]byte(input.Transcript))
	raw, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	model.SetResourceRaw(datatypes.JSON(raw))
	model.SetOriginBase(models.OriginBase{ModelBase: models.ModelBase{ID: uuid.New()}, UserID: s.UserID, SourceID: source, SourceResourceType: "DocumentReference", SourceResourceID: id})
	model.SetSortTitle(&input.Title)
	model.SetSortDate(&now)
	origin := documentImportOrigin
	model.SetSourceUri(&origin)
	result := &models.DocumentImportResult{SourceID: source, ResourceType: "DocumentReference", ResourceID: id, OriginalSHA256: originalHash}
	err = gr.GormClient.WithContext(ctx).Session(&gorm.Session{Logger: privateSQLLogger{gr.GormClient.Logger}}).Transaction(func(tx *gorm.DB) error {
		// Reserve the SQLite writer / PostgreSQL source row, repeating real
		// ownership before any insert. No source credentials/timestamps change.
		check := tx.Model(&models.SourceCredential{}).Where("id = ? AND user_id = ? AND platform_type = ? AND deleted_at IS NULL", source, s.UserID, sp.PlatformTypeFasten).
			Where("EXISTS (SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL)", s.UserID).UpdateColumn("id", gorm.Expr("id"))
		if check.Error != nil {
			return check.Error
		}
		if check.RowsAffected != 1 {
			return ErrSourceAccessDenied
		}
		created := tx.Omit(clause.Associations).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "source_id"}, {Name: "source_resource_type"}, {Name: "source_resource_id"}}, DoNothing: true}).Create(model)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			table, err := dm.GetTableNameByResourceType("DocumentReference")
			if err != nil {
				return err
			}
			var count int64
			if err = tx.Table(table).Where("source_id = ? AND user_id = ? AND source_resource_type = ? AND source_resource_id = ? AND source_uri = ? AND deleted_at IS NULL", source, s.UserID, "DocumentReference", id, documentImportOrigin).Count(&count).Error; err != nil {
				return err
			}
			if count != 1 {
				return fmt.Errorf("%w: document identity conflict", ErrSourceAccessInput)
			}
			return nil
		}
		result.Created = true
		// Deliberately no LocalResource marker: imported originals/transcripts
		// remain read-only to delegates even with a source-level EDIT grant.
		return appendSourceAudit(tx, source, s.UserID, s.UserID, "import_document", nil, "DocumentReference", id)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
