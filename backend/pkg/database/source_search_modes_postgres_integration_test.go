//go:build postgres_integration

package database

import (
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func TestSourceSearchModesPostgres(t *testing.T) {
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
	runSourceSearchModeCorpus(t, &GormRepository{GormClient: db})
}
