package database

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestObservationReferenceQueries(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-reference.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-reference-query-key")
	cfg.Set("database.validation_mode", false)
	repo, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	owner := &models.User{Username: "synthetic-reference-owner", Password: "synthetic-password"}
	other := &models.User{Username: "synthetic-reference-other", Password: "synthetic-password"}
	for _, user := range []*models.User{owner, other} {
		require.NoError(t, repo.CreateUser(context.Background(), user))
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, user.Username)
		source := &models.SourceCredential{ModelBase: models.ModelBase{ID: uuid.New()}, UserID: user.ID}
		for i := 1; i <= 2; i++ {
			raw := []byte(fmt.Sprintf(`{"resourceType":"Observation","id":"synthetic-%d","meta":{"profile":["https://synthetic.invalid/profile/common","https://synthetic.invalid/profile/specific-%d"]},"status":"final","code":{"coding":[{"code":"synthetic"}]},"subject":{"reference":"Patient/synthetic-%d"},"encounter":{"reference":"Encounter/synthetic-common"},"focus":[{"reference":"Patient/synthetic-common"},{"reference":"Patient/synthetic-%d"}],"valueQuantity":{"value":12.5,"code":"mg","system":"http://unitsofmeasure.org"},"component":[{"code":{"text":"Synthetic"},"valueQuantity":{"value":0,"code":"mg","system":"http://unitsofmeasure.org"}}]}`, i, i, i, i))
			created, err := repo.UpsertRawResource(ctx, source, sm.RawResourceFhir{SourceResourceType: "Observation", SourceResourceID: fmt.Sprintf("synthetic-%d", i), ResourceRaw: raw})
			require.NoError(t, err)
			require.True(t, created)
		}
	}
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	for _, tc := range []struct {
		name  string
		where map[string]interface{}
		ids   []string
	}{
		{"subject", map[string]interface{}{"subject": "Patient/synthetic-1"}, []string{"synthetic-1"}},
		{"encounter", map[string]interface{}{"encounter": "Encounter/synthetic-common"}, []string{"synthetic-1", "synthetic-2"}},
		{"no match", map[string]interface{}{"subject": "Patient/synthetic-absent"}, []string{}},
		{"OR", map[string]interface{}{"subject": "Patient/synthetic-1,Patient/synthetic-2"}, []string{"synthetic-1", "synthetic-2"}},
		{"AND", map[string]interface{}{"subject": []string{"Patient/synthetic-1", "Patient/synthetic-2"}}, []string{}},
		{"multiple fields", map[string]interface{}{"subject": "Patient/synthetic-2", "encounter": "Encounter/synthetic-common"}, []string{"synthetic-2"}},
		{"array matches without duplicate rows", map[string]interface{}{"focus": "Patient/synthetic-common,Patient/synthetic-1"}, []string{"synthetic-1", "synthetic-2"}},
		{"AND across array entries", map[string]interface{}{"focus": []string{"Patient/synthetic-common", "Patient/synthetic-1"}}, []string{"synthetic-1"}},
		{"canonical profile", map[string]interface{}{"metaProfile": "https://synthetic.invalid/profile/common"}, []string{"synthetic-1", "synthetic-2"}},
		{"canonical profile AND", map[string]interface{}{"metaProfile": []string{"https://synthetic.invalid/profile/common", "https://synthetic.invalid/profile/specific-1"}}, []string{"synthetic-1"}},
		{"canonical no match", map[string]interface{}{"metaProfile": "https://synthetic.invalid/profile/absent"}, []string{}},
		{"object and canonical references", map[string]interface{}{"subject": "Patient/synthetic-1", "metaProfile": "https://synthetic.invalid/profile/common"}, []string{"synthetic-1"}},
		{"quantity subpath", map[string]interface{}{"subject": "Patient/synthetic-1", "valueQuantity": "12.5|http://unitsofmeasure.org|mg", "comboValueQuantity": "0|http://unitsofmeasure.org|mg"}, []string{"synthetic-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := repo.QueryResources(ctx, models.QueryResource{From: "Observation", Where: tc.where})
			require.NoError(t, err)
			encoded, err := json.Marshal(rows)
			require.NoError(t, err)
			var observations []dm.FhirObservation
			require.NoError(t, json.Unmarshal(encoded, &observations))
			ids := []string{}
			for _, row := range observations {
				require.Equal(t, owner.ID, row.UserID)
				ids = append(ids, row.SourceResourceID)
			}
			sort.Strings(ids)
			require.Equal(t, tc.ids, ids)
		})
	}
	for _, tc := range []struct{ field, value string }{
		{"subject", ""}, {"subject", " "}, {"subject:identifier", "synthetic"}, {"subject:Patient", "synthetic-1"},
	} {
		t.Run("unsupported "+tc.field+tc.value, func(t *testing.T) {
			_, err := repo.QueryResources(ctx, models.QueryResource{From: "Observation", Where: map[string]interface{}{tc.field: tc.value}})
			require.Error(t, err, "unsupported or empty reference semantics must be explicit")
		})
	}
	t.Run("encrypted cold reopen", func(t *testing.T) {
		require.NoError(t, repo.Close())
		reopened, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		rows, err := reopened.QueryResources(ctx, models.QueryResource{From: "Observation", Where: map[string]interface{}{"subject": "Patient/synthetic-1"}})
		require.NoError(t, err)
		encoded, err := json.Marshal(rows)
		require.NoError(t, err)
		var observations []dm.FhirObservation
		require.NoError(t, json.Unmarshal(encoded, &observations))
		require.Len(t, observations, 1)
		require.Equal(t, owner.ID, observations[0].UserID)
		require.Equal(t, "synthetic-1", observations[0].SourceResourceID)
		rows, err = reopened.QueryResources(ctx, models.QueryResource{From: "Observation", Where: map[string]interface{}{"metaProfile": "https://synthetic.invalid/profile/common"}})
		require.NoError(t, err)
		encoded, err = json.Marshal(rows)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(encoded, &observations))
		require.Len(t, observations, 2)
		for _, row := range observations {
			require.Equal(t, owner.ID, row.UserID)
		}
	})
}
