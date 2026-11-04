//go:build postgres_integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Query fixture only: avoid migrating unrelated generated association graphs.
// Full application migrations remain covered by the existing repository suites.
type sourceSearchPGRecord struct {
	models.OriginBase
	SortTitle *string
	SortDate  *time.Time
}

func TestSourceSearchPostgresAuthorizationAndLiteralMatching(t *testing.T) {
	dsn := disposablePostgres(t)
	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.SourceCredential{}, &models.SourceGrant{}, &models.SourceAccessAudit{}))
	r := &GormRepository{GormClient: db}
	owner := models.User{Username: "synthetic-postgres-search-owner"}
	delegate := models.User{Username: "synthetic-postgres-search-delegate"}
	require.NoError(t, db.Create(&owner).Error)
	require.NoError(t, db.Create(&delegate).Error)
	source := models.SourceCredential{UserID: owner.ID, Display: "Synthetic search source"}
	require.NoError(t, db.Create(&source).Error)
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, delegate.Username)
	table, err := dm.GetTableNameByResourceType("Observation")
	require.NoError(t, err)
	require.NoError(t, db.Table(table).AutoMigrate(&sourceSearchPGRecord{}))
	titles := []string{"Synthetic clinic 100% follow_up!", "Synthetic clinic 100units followXup!"}
	for i, title := range titles {
		record := sourceSearchPGRecord{OriginBase: models.OriginBase{UserID: owner.ID, SourceID: source.ID, SourceResourceType: "Observation", SourceResourceID: []string{"synthetic-one", "synthetic-two"}[i]}, SortTitle: &title}
		require.NoError(t, db.Table(table).Create(&record).Error)
	}
	_, err = r.SearchSourceResources(ctx, source.ID, "Observation", "clinic", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	ownerCtx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	grant, err := r.SetSourceGrant(ownerCtx, source.ID, delegate.Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	listed, err := r.ListSharedResources(ctx, source.ID, "Observation", 0)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	for _, query := range []string{"100%", "follow_up!", "_"} {
		hits, err := r.SearchSourceResources(ctx, source.ID, "Observation", query, 0)
		require.NoError(t, err)
		require.Len(t, hits, 1)
		require.Equal(t, "synthetic-one", hits[0].SourceResourceID)
	}
	// Both physical owner-column types exist in legacy/native schemas. They
	// must preserve the same exact actor/source comparison, never a fallback.
	require.NoError(t, db.Exec("ALTER TABLE fhir_observation ALTER COLUMN user_id TYPE uuid USING user_id::uuid").Error)
	hits, err := r.SearchSourceResources(ctx, source.ID, "Observation", "100%", 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.NoError(t, r.RevokeSourceGrant(ownerCtx, source.ID, grant.ID))
	_, err = r.SearchSourceResources(ctx, source.ID, "Observation", "clinic", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
}
