//go:build postgres_integration

package database

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDocumentImportPostgresOwnerRetryAndDelegation(t *testing.T) {
	dsn := disposablePostgres(t)
	// Query/transaction fixture: migrate only these models, not the unrelated
	// generated association graph. This is not full application migration proof.
	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{Logger: logger.Discard, IgnoreRelationshipsWhenMigrating: true})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.SourceCredential{}, &models.SourceGrant{}, &models.SourceAccessAudit{}, &models.LocalResource{}, &dm.FhirDocumentReference{}))
	r := &GormRepository{GormClient: db}
	owner := models.User{Username: "synthetic-pg-document-owner"}
	delegate := models.User{Username: "synthetic-pg-document-delegate"}
	require.NoError(t, db.Create(&owner).Error)
	require.NoError(t, db.Create(&delegate).Error)
	source := models.SourceCredential{UserID: owner.ID, PlatformType: sp.PlatformTypeFasten, Display: "Synthetic local source"}
	require.NoError(t, db.Create(&source).Error)
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	other := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, delegate.Username)
	input := models.ReviewedDocument{Filename: "synthetic.pdf", Title: "Synthetic import", Original: []byte("%PDF-1.7\n% synthetic original\n%%EOF"), Transcript: "Private reviewed transcript café 100% follow_up! quoted \"text\" and slash \\ العربية", Reviewed: true}
	first, err := r.ImportReviewedDocument(ctx, source.ID, input)
	require.NoError(t, err)
	require.True(t, first.Created)
	second, err := r.ImportReviewedDocument(ctx, source.ID, input)
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Equal(t, first.ResourceID, second.ResourceID)
	grant, err := r.SetSourceGrant(ctx, source.ID, delegate.Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	shared, err := r.GetSharedResource(other, source.ID, "DocumentReference", first.ResourceID)
	require.NoError(t, err)
	require.False(t, shared.Editable)
	require.ErrorIs(t, r.UpdateSharedResource(other, source.ID, "DocumentReference", first.ResourceID, json.RawMessage(shared.ResourceRaw)), ErrSourceAccessDenied)
	_, err = r.ImportReviewedDocument(other, source.ID, input)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	metadata, err := r.ListSourceMetadata(other, source.ID, "DocumentReference", 0)
	require.NoError(t, err)
	require.Len(t, metadata, 1)
	var projection struct{ Value string }
	require.NoError(t, db.Table("fhir_document_reference").Select("text AS value").Where("source_id = ? AND source_resource_id = ?", source.ID, first.ResourceID).Find(&projection).Error)
	var projected []string
	require.NoError(t, json.Unmarshal([]byte(projection.Value), &projected))
	require.Equal(t, []string{input.Transcript}, projected)
	for _, term := range []string{"reviewed transcript", "café", "100%", "follow_up!", `"text"`, `\`, "العربية"} {
		hits, err := r.SearchSourceResources(other, source.ID, "DocumentReference", term, 0)
		require.NoError(t, err)
		require.Len(t, hits, 1, term)
		require.Equal(t, first.ResourceID, hits[0].SourceResourceID)
	}
	hits, err := r.SearchSourceResources(other, source.ID, "DocumentReference", "100units", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
	// Native datatypes.JSON can migrate to jsonb; legacy schemas can keep text.
	// The query must work with both without trying to parse an empty JSON literal.
	require.NoError(t, db.Exec("ALTER TABLE fhir_document_reference ALTER COLUMN text TYPE text USING text::text").Error)
	hits, err = r.SearchSourceResources(other, source.ID, "DocumentReference", "café", 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.NoError(t, db.Model(&dm.FhirDocumentReference{}).Where("source_id = ? AND source_resource_id = ?", source.ID, first.ResourceID).Update("text", nil).Error)
	hits, err = r.SearchSourceResources(other, source.ID, "DocumentReference", "reviewed transcript", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
	hits, err = r.SearchSourceResources(other, source.ID, "DocumentReference", input.Title, 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.NoError(t, r.RevokeSourceGrant(ctx, source.ID, grant.ID))
	_, err = r.ListSourceMetadata(other, source.ID, "DocumentReference", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	_, err = r.SearchSourceResources(other, source.ID, "DocumentReference", "reviewed transcript", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
}
