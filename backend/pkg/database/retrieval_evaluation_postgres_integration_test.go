//go:build postgres_integration

package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRetrievalEvaluationPostgres(t *testing.T) {
	if os.Getenv("CHARTROVE_RETRIEVAL_EVALUATION") != "1" {
		t.Skip("bounded synthetic retrieval evaluation is opt-in")
	}
	for _, size := range []int{1000, 10000} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			// Existing loopback-only, pinned-image, tmpfs fixture: no existing DB,
			// external Docker endpoint, host volume or image download is accepted.
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
			var version string
			require.NoError(t, db.Raw("SELECT version()").Scan(&version).Error)
			t.Logf("PostgreSQL version=%s", version)
			runRetrievalEvaluation(t, &GormRepository{GormClient: db}, size, []string{"postgres_gin"})
		})
	}
}
