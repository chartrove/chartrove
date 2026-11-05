//go:build postgres_integration

package database

import (
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRetrievalQualityPostgres(t *testing.T) {
	// Existing local Unix-socket / pinned-image / loopback / tmpfs fixture;
	// never an existing database, host volume, remote service or image download.
	dsn := disposablePostgres(t)
	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{Logger: logger.Discard, IgnoreRelationshipsWhenMigrating: true})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.SourceCredential{}, &models.SourceGrant{}, &models.SourceAccessAudit{}))
	table, err := dm.GetTableNameByResourceType("DocumentReference")
	require.NoError(t, err)
	require.NoError(t, db.Table(table).AutoMigrate(&retrievalEvalRecord{}))
	var version, locale string
	require.NoError(t, db.Raw("SELECT version()").Scan(&version).Error)
	require.NoError(t, db.Raw("SELECT datctype FROM pg_database WHERE datname = current_database()").Scan(&locale).Error)
	t.Logf("PostgreSQL %s locale=%s", version, locale)
	runRetrievalQuality(t, &GormRepository{GormClient: db})
}
