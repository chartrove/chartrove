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
	input := models.ReviewedDocument{Filename: "synthetic.pdf", Title: "Synthetic import", Original: []byte("%PDF-1.7\n% synthetic original\n%%EOF"), Transcript: "Synthetic transcript", Reviewed: true}
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
	require.NoError(t, r.RevokeSourceGrant(ctx, source.ID, grant.ID))
	_, err = r.ListSourceMetadata(other, source.ID, "DocumentReference", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
}
