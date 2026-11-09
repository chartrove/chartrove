package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrClinicalSnapshotConflict = errors.New("clinical snapshot identity requires review")
var ErrClinicalSnapshotStorage = errors.New("clinical snapshot storage failed")

// This optional owner-only boundary does not grant delegates import authority or
// attest native retrieval. Origin metadata is supplied by the importing owner.
type ClinicalSnapshotRepository interface {
	ImportClinicalSnapshot(context.Context, string, string, []byte) (*models.ClinicalSnapshotImportResult, error)
	ListClinicalSnapshots(context.Context, uuid.UUID, int, int) ([]models.ClinicalSnapshotRecord, error)
	GetClinicalSnapshotOriginal(context.Context, uuid.UUID, uuid.UUID) ([]byte, error)
}

var _ ClinicalSnapshotRepository = (*GormRepository)(nil)

func (gr *GormRepository) clinicalRepository(ctx context.Context) *GormRepository {
	r := *gr
	r.GormClient = gr.GormClient.WithContext(ctx).Session(&gorm.Session{Logger: privateSQLLogger{gr.GormClient.Logger}})
	return &r
}

// ImportClinicalSnapshot atomically stores the complete validated collection and
// its immutable original. Missing Observations are never interpreted as deletes.
// Historical exact replay does not roll current records back to an old snapshot.
func (gr *GormRepository) ImportClinicalSnapshot(ctx context.Context, fhirBase, display string, raw []byte) (*models.ClinicalSnapshotImportResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(raw) > clinicalsnapshot.MaxBytes {
		return nil, ErrSourceAccessInput
	}
	raw = bytes.Clone(raw)
	fhirBase, display = strings.TrimRight(fhirBase, "/"), strings.TrimSpace(display)
	if !utf8.ValidString(display) || utf8.RuneCountInString(display) < 1 || utf8.RuneCountInString(display) > 100 || strings.IndexFunc(display, unicode.IsControl) >= 0 {
		return nil, ErrSourceAccessInput
	}
	snapshot, err := clinicalsnapshot.Parse(raw, fhirBase)
	if err != nil {
		return nil, ErrSourceAccessInput
	}
	repo := gr.clinicalRepository(ctx)
	actor, err := repo.sourceAccessActor(ctx)
	if err != nil {
		return nil, ErrSourceAccessDenied
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Build projections before reserving a writer. Never feed the result back
	// into the originals: unknown fields and decimal lexemes remain untouched.
	projections := make([]dm.IFhirResourceModel, len(snapshot.Entries))
	for i, entry := range snapshot.Entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		model, err := dm.NewFhirResourceModelByType(entry.Type)
		if err != nil || model.PopulateAndExtractSearchParameters(entry.Raw) != nil {
			return nil, ErrSourceAccessInput
		}
		// Match the existing JSON serializer's whitespace/HTML escaping only;
		// RawMessage never round-trips numbers through float64. The snapshot
		// record below additionally preserves every original document byte.
		projectedRaw, err := json.Marshal(entry.Raw)
		if err != nil {
			return nil, ErrSourceAccessInput
		}
		model.SetResourceRaw(datatypes.JSON(projectedRaw))
		title, date := clinicalPresentation(entry.Raw, entry.Type)
		model.SetSortTitle(&title)
		model.SetSortDate(date)
		uri := entry.FullURL
		model.SetSourceUri(&uri)
		projections[i] = model
	}
	digest := sha256.Sum256(raw)
	sha := hex.EncodeToString(digest[:])
	endpoint := uuid.NewSHA1(uuid.NameSpaceURL, []byte("chartrove-direct-smart:"+fhirBase))
	result := &models.ClinicalSnapshotImportResult{SHA256: sha, PatientID: snapshot.PatientID}
	err = repo.GormClient.Transaction(func(tx *gorm.DB) error {
		// First transactional statement: serialize this owner's imports and
		// acquire SQLite's write reservation before any authorization reads.
		if err := reserveOwnerWrite(tx, actor); err != nil {
			return err
		}
		var sources []models.SourceCredential
		if err := tx.Where("user_id = ? AND endpoint_id = ? AND patient = ?", actor.ID, endpoint, snapshot.PatientID).Limit(2).Find(&sources).Error; err != nil {
			return err
		}
		source := models.SourceCredential{UserID: actor.ID, EndpointID: endpoint, Patient: snapshot.PatientID, PlatformType: sp.PlatformTypeManual,
			Display: fmt.Sprintf("Direct SMART snapshot: %s (%s)", display, fhirBase)}
		if len(sources) > 1 {
			return ErrClinicalSnapshotConflict
		}
		if len(sources) == 1 {
			source = sources[0]
			if source.ID == uuid.Nil || source.DeletedAt != nil || source.PlatformType != sp.PlatformTypeManual || source.ClientId != "" || source.AccessToken != "" || source.RefreshToken != "" || source.IdToken != "" || source.CodeVerifier != "" || source.CodeChallenge != "" || source.DynamicClientId != "" || len(source.DynamicClientJWKS) != 0 {
				return ErrClinicalSnapshotConflict
			}
		} else if err := tx.Omit(clause.Associations).Create(&source).Error; err != nil {
			return err
		}
		result.SourceID = source.ID
		// Legacy source metadata alone is not proof of a single Patient. Never
		// add a second Patient to a misidentified or interrupted old source.
		var patients []models.ResourceBase
		if err := tx.Table("fhir_patient").Where("source_id = ?", source.ID).Limit(2).Find(&patients).Error; err != nil {
			return err
		}
		if len(patients) > 1 || (len(patients) == 1 && !clinicalExistingPatient(patients[0], actor.ID, snapshot.PatientID)) {
			return ErrClinicalSnapshotConflict
		}
		var prior []models.ClinicalSnapshotRecord
		if err := tx.Select("id", "owner_user_id", "fhir_base_url", "patient_id").Where("source_id = ? AND sha256 = ?", source.ID, sha).Limit(2).Find(&prior).Error; err != nil {
			return err
		}
		if len(prior) > 0 {
			if len(prior) != 1 || prior[0].OwnerUserID != actor.ID || prior[0].FHIRBaseURL != fhirBase || prior[0].PatientID != snapshot.PatientID {
				return ErrClinicalSnapshotConflict
			}
			result.SnapshotID, result.AlreadyImported = prior[0].ID, true
			return nil
		}
		for i, entry := range snapshot.Entries {
			model := projections[i]
			model.SetOriginBase(models.OriginBase{UserID: actor.ID, SourceID: source.ID, SourceResourceType: entry.Type, SourceResourceID: entry.ID})
			table, err := dm.GetTableNameByResourceType(entry.Type)
			if err != nil {
				return err
			}
			var rows []models.ResourceBase
			if err := tx.Table(table).Where("source_id = ? AND source_resource_type = ? AND source_resource_id = ?", source.ID, entry.Type, entry.ID).Limit(2).Find(&rows).Error; err != nil {
				return err
			}
			if len(rows) > 1 || (len(rows) == 1 && (rows[0].UserID != actor.ID || rows[0].DeletedAt != nil)) {
				return ErrClinicalSnapshotConflict
			}
			if err := tx.Where("owner_user_id = ? AND source_id = ? AND resource_type = ? AND resource_id = ?", actor.ID, source.ID, entry.Type, entry.ID).Delete(&models.LocalResource{}).Error; err != nil {
				return err
			}
			unchanged := false
			if len(rows) == 1 && rows[0].SourceUri != nil && *rows[0].SourceUri == entry.FullURL {
				unchanged, err = clinicalStoredJSONMatches(tx, rows[0].ResourceRaw, model.GetResourceRaw())
				if err != nil {
					return err
				}
			}
			if len(rows) == 0 {
				if err := tx.Omit(clause.Associations).Create(model).Error; err != nil {
					return err
				}
				result.Created++
			} else if unchanged {
				// fullUrl is entry provenance, not part of the resource JSON.
				// Matching stored JSON alone cannot make this row unchanged.
				result.Unchanged++
			} else {
				// Select all projection fields, including zero values, so removed
				// valueQuantity/valueString/search fields do not survive an update.
				updated := tx.Table(table).Where("id = ? AND user_id = ? AND source_id = ? AND deleted_at IS NULL", rows[0].ID, actor.ID, source.ID).
					Select("*").Omit("id", "user_id", "source_id", "source_resource_type", "source_resource_id", "created_at", "deleted_at", clause.Associations).Updates(model)
				if updated.Error != nil {
					return updated.Error
				}
				if updated.RowsAffected != 1 {
					return ErrClinicalSnapshotConflict
				}
				result.Updated++
			}
		}
		r := *repo
		r.GormClient = tx
		for _, entry := range snapshot.Entries {
			if entry.Type == "Observation" {
				if err := r.AddResourceAssociation(ctx, &source, "Observation", entry.ID, &source, "Patient", snapshot.PatientID); err != nil {
					return err
				}
			}
		}
		record := models.ClinicalSnapshotRecord{ID: uuid.New(), OwnerUserID: actor.ID, ImportedByUserID: actor.ID, Acquisition: models.ClinicalSnapshotOwnerAcquisition, SourceID: source.ID,
			SHA256: sha, FHIRBaseURL: fhirBase, PatientID: snapshot.PatientID, ResourceCount: len(snapshot.Entries), CreatedAt: time.Now().UTC(), Payload: raw}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		result.SnapshotID = record.ID
		// Last write: any failure here must roll back originals, source,
		// projections, provenance markers and associations together.
		return appendSourceAudit(tx, source.ID, actor.ID, actor.ID, "import_clinical_snapshot", nil, "Bundle", record.ID.String())
	})
	if err != nil {
		return nil, clinicalSnapshotError(ctx, err)
	}
	return result, nil
}

// PostgreSQL JSONB normalizes whitespace, object order and numeric notation on
// storage. Compare against its own precision-preserving text representation, not a Go
// float/map reserialization or JSONB numeric-value equality. Exact whole-snapshot
// bytes and SHA remain separately retained; SQLite keeps its byte comparison.
func clinicalStoredJSONMatches(tx *gorm.DB, stored, incoming []byte) (bool, error) {
	if bytes.Equal(stored, incoming) {
		return true, nil
	}
	if tx.Dialector.Name() != "postgres" {
		return false, nil
	}
	// GORM Scan's temporary recorder bypasses the private parameter filter;
	// Rows preserves it and propagates query errors without a nil Row panic.
	rows, err := tx.Raw("SELECT CAST(? AS jsonb)::text", string(incoming)).Rows()
	if err != nil {
		return false, err
	}
	if rows == nil {
		return false, ErrClinicalSnapshotStorage
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return false, err
		}
		return false, ErrClinicalSnapshotStorage
	}
	var normalized string
	if err := rows.Scan(&normalized); err != nil {
		return false, err
	}
	if rows.Next() {
		return false, ErrClinicalSnapshotStorage
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return bytes.Equal(stored, []byte(normalized)), nil
}

// Imports and source/account deletion share this reservation. Otherwise a
// concurrent import could create originals after deletion enumerates sources,
// leaving private orphaned payloads behind a deleted account.
func reserveOwnerWrite(tx *gorm.DB, actor *models.User) error {
	check := tx.Model(&models.User{}).Where("id = ? AND username = ? AND deleted_at IS NULL", actor.ID, actor.Username).UpdateColumn("id", gorm.Expr("id"))
	if check.Error != nil {
		return check.Error
	}
	if check.RowsAffected != 1 {
		return ErrSourceAccessDenied
	}
	return nil
}

func clinicalSnapshotError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, ErrSourceAccessDenied) {
		return ErrSourceAccessDenied
	}
	if errors.Is(err, ErrClinicalSnapshotConflict) {
		return ErrClinicalSnapshotConflict
	}
	return ErrClinicalSnapshotStorage
}

func (gr *GormRepository) clinicalSnapshotQuery(ctx context.Context, source uuid.UUID) (*gorm.DB, error) {
	r := gr.clinicalRepository(ctx)
	actor, err := r.sourceAccessActor(ctx)
	if err != nil || source == uuid.Nil {
		return nil, ErrSourceAccessDenied
	}
	if _, err := r.ownedAccessSource(ctx, source); err != nil {
		return nil, ErrSourceAccessDenied
	}
	return r.GormClient.Model(&models.ClinicalSnapshotRecord{}).Where("owner_user_id = ? AND source_id = ?", actor.ID, source).
		Where("EXISTS (SELECT 1 FROM source_credentials WHERE id = ? AND user_id = ? AND deleted_at IS NULL)", source, actor.ID), nil
}

func (gr *GormRepository) ListClinicalSnapshots(ctx context.Context, source uuid.UUID, limit, offset int) ([]models.ClinicalSnapshotRecord, error) {
	if limit < 1 || limit > 200 || offset < 0 || offset > 1000000 {
		return nil, ErrSourceAccessInput
	}
	query, err := gr.clinicalSnapshotQuery(ctx, source)
	if err != nil {
		return nil, err
	}
	rows := []models.ClinicalSnapshotRecord{}
	if err := query.Omit("payload").Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, clinicalSnapshotError(ctx, err)
	}
	return rows, nil
}

func (gr *GormRepository) GetClinicalSnapshotOriginal(ctx context.Context, source, snapshot uuid.UUID) ([]byte, error) {
	query, err := gr.clinicalSnapshotQuery(ctx, source)
	if err != nil {
		return nil, err
	}
	var row models.ClinicalSnapshotRecord
	if err := query.Select("payload", "sha256").Where("id = ? AND length(payload) BETWEEN 1 AND ?", snapshot, clinicalsnapshot.MaxBytes).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSourceAccessDenied
		}
		return nil, clinicalSnapshotError(ctx, err)
	}
	digest := sha256.Sum256(row.Payload)
	if hex.EncodeToString(digest[:]) != row.SHA256 {
		return nil, ErrClinicalSnapshotStorage
	}
	return row.Payload, nil
}

func clinicalExistingPatient(row models.ResourceBase, owner uuid.UUID, patientID string) bool {
	if row.UserID != owner || row.DeletedAt != nil || row.SourceResourceType != "Patient" || row.SourceResourceID != patientID {
		return false
	}
	fields, err := decodeLocalResource(json.RawMessage(row.ResourceRaw))
	var kind, id string
	return err == nil && json.Unmarshal(fields["resourceType"], &kind) == nil && json.Unmarshal(fields["id"], &id) == nil && kind == "Patient" && id == patientID
}

// Presentation fields are projections only. Partial dates remain unaltered in
// raw FHIR; do not silently invent a day, time or timezone for timeline sorting.
func clinicalPresentation(raw []byte, kind string) (string, *time.Time) {
	if kind == "Patient" {
		return "Patient", nil
	}
	var fields struct {
		Code struct {
			Text   string `json:"text"`
			Coding []struct {
				Display string `json:"display"`
			} `json:"coding"`
		} `json:"code"`
		EffectiveDateTime string `json:"effectiveDateTime"`
		EffectivePeriod   struct {
			Start string `json:"start"`
		} `json:"effectivePeriod"`
	}
	_ = json.Unmarshal(raw, &fields)
	title := fields.Code.Text
	if title == "" && len(fields.Code.Coding) > 0 {
		title = fields.Code.Coding[0].Display
	}
	if title == "" {
		title = "Observation"
	}
	if runes := []rune(title); len(runes) > 256 {
		title = string(runes[:256]) + "…"
	}
	date := fields.EffectiveDateTime
	if date == "" {
		date = fields.EffectivePeriod.Start
	}
	if parsed, err := time.Parse(time.RFC3339Nano, date); err == nil {
		return title, &parsed
	}
	return title, nil
}
