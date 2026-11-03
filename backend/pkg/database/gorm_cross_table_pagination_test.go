package database

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCrossTablePaginationPreservesScope(t *testing.T) {
	appConfig, err := config.Create()
	require.NoError(t, err)
	appConfig.SetDefault("database.type", "sqlite")
	appConfig.SetDefault("database.location", filepath.Join(t.TempDir(), "synthetic.db"))
	appConfig.SetDefault("database.encryption.enabled", false)
	appConfig.SetDefault("database.validation_mode", false)
	repo, err := newTestRepository(t, appConfig, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	owner := &models.User{Username: "synthetic-pagination-owner", Password: "synthetic-password"}
	other := &models.User{Username: "synthetic-pagination-other", Password: "synthetic-password"}
	require.NoError(t, repo.CreateUser(context.Background(), owner))
	require.NoError(t, repo.CreateUser(context.Background(), other))
	targetSource := uuid.New()
	for _, scope := range []struct {
		user   *models.User
		source uuid.UUID
	}{
		{owner, targetSource}, {owner, uuid.New()}, {other, targetSource},
	} {
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, scope.user.Username)
		source := &models.SourceCredential{ModelBase: models.ModelBase{ID: scope.source}, UserID: scope.user.ID}
		for _, kind := range []string{"Patient", "Observation"} {
			for i := 0; i < 3; i++ {
				id := fmt.Sprintf("%s-%s-%d", scope.user.Username, kind, i)
				raw := fmt.Sprintf(`{"resourceType":%q,"id":%q}`, kind, id)
				created, err := repo.UpsertRawResource(ctx, source, sourceModels.RawResourceFhir{
					SourceResourceType: kind, SourceResourceID: id, ResourceRaw: []byte(raw),
				})
				require.NoError(t, err)
				require.True(t, created)
			}
		}
	}
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	for _, offset := range []int{0, 1, 3} {
		t.Run(fmt.Sprintf("offset-%d", offset), func(t *testing.T) {
			rows, err := repo.ListResources(ctx, models.ListResourceQueryOptions{
				SourceID: targetSource.String(), Limit: 1, Offset: offset,
			})
			require.NoError(t, err)
			if offset == 3 {
				require.Empty(t, rows, "offset must apply independently to each table")
				return
			}
			require.Len(t, rows, 2, "one result per populated resource table")
			kinds := make(map[string]int)
			for _, row := range rows {
				require.Equal(t, owner.ID, row.UserID)
				require.Equal(t, targetSource, row.SourceID)
				kinds[row.SourceResourceType]++
			}
			require.Equal(t, map[string]int{"Patient": 1, "Observation": 1}, kinds)
		})
	}
}
