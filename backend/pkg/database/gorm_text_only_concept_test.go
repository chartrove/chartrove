package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	mock_config "github.com/fastenhealth/fasten-onprem/backend/pkg/config/mock"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	databaseModel "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTextOnlyConceptsPersistInEncryptedRepository(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic-concepts.db")
	settings := mock_config.NewMockInterface(gomock.NewController(t))
	settings.EXPECT().GetString("database.location").Return(path).AnyTimes()
	settings.EXPECT().GetString("database.type").Return("sqlite").AnyTimes()
	settings.EXPECT().GetString("log.level").Return("INFO").AnyTimes()
	settings.EXPECT().GetBool("database.validation_mode").Return(false).AnyTimes()
	settings.EXPECT().GetBool("database.encryption.enabled").Return(true).AnyTimes()
	settings.EXPECT().IsSet("database.encryption.key").Return(true).AnyTimes()
	settings.EXPECT().GetString("database.encryption.key").Return("synthetic-concept-key-only").AnyTimes()
	open := func() DatabaseRepository {
		repo, err := newTestRepository(t, settings, logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		return repo
	}
	repo := open()
	owner := &models.User{Username: "synthetic-concept-owner", Password: "synthetic-password-only"}
	require.NoError(t, repo.CreateUser(context.Background(), owner))
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	source := &models.SourceCredential{ModelBase: models.ModelBase{ID: uuid.New()}, UserID: owner.ID}
	raw := []byte(`{"resourceType":"AllergyIntolerance","id":"synthetic-only","code":{"text":"Synthetic concept Ω"},"reaction":[{"manifestation":[{"text":"Synthetic manifestation Ω"}]}],"extension":[{"url":"urn:synthetic:unknown","valueDecimal":12345678901234567890}],"unknown":{"number":0,"flag":false}}`)
	created, err := repo.UpsertRawResource(ctx, source, sourceModels.RawResourceFhir{
		SourceResourceType: "AllergyIntolerance", SourceResourceID: "synthetic-only", ResourceRaw: raw,
	})
	require.NoError(t, err)
	require.True(t, created)
	check := func(repo DatabaseRepository) {
		var stored databaseModel.FhirAllergyIntolerance
		err := repo.(*GormRepository).GormClient.Where("user_id = ? AND source_id = ? AND source_resource_id = ?", owner.ID, source.ID, "synthetic-only").First(&stored).Error
		require.NoError(t, err)
		require.JSONEq(t, `[{"text":"Synthetic concept Ω"}]`, string(stored.Code))
		require.JSONEq(t, `[{"text":"Synthetic manifestation Ω"}]`, string(stored.Manifestation))
		require.Equal(t, raw, []byte(stored.ResourceRaw), "raw values and unknown fields must survive unchanged")
	}
	check(repo)
	require.NoError(t, repo.Close())
	header, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Greater(t, len(header), 16)
	require.NotEqual(t, "SQLite format 3\x00", string(header[:16]), "the fixture must actually be encrypted")
	check(open())
}
