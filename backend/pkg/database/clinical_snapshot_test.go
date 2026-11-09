package database

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const clinicalTestBase = "https://clinical.example.invalid/fhir"
const clinicalTestKey = "synthetic-clinical-storage-key"

type clinicalTestEvents struct {
	event_bus.Interface
	count atomic.Int32
}

func (e *clinicalTestEvents) PublishMessage(models.EventInterface) error { e.count.Add(1); return nil }

func clinicalTestConfig(t *testing.T, path string) config.Interface {
	t.Helper()
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", path)
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", clinicalTestKey)
	return cfg
}

func clinicalTestFixture(t *testing.T) (*GormRepository, config.Interface, []context.Context, []*models.User, *clinicalTestEvents) {
	t.Helper()
	cfg := clinicalTestConfig(t, filepath.Join(t.TempDir(), "synthetic-clinical.db"))
	events := &clinicalTestEvents{Interface: event_bus.NewNoopEventBusServer()}
	appLog := logrus.New()
	appLog.SetOutput(&bytes.Buffer{})
	db, err := newTestRepository(t, cfg, appLog, events)
	require.NoError(t, err)
	r := db.(*GormRepository)
	users := []*models.User{}
	contexts := []context.Context{}
	for i, name := range []string{"synthetic-clinical-owner", "synthetic-clinical-delegate", "synthetic-clinical-admin"} {
		u := &models.User{ModelBase: models.ModelBase{ID: uuid.New()}, Username: name, Password: "synthetic-placeholder-not-authenticated", Role: pkg.UserRoleUser}
		if i == 2 {
			u.Role = pkg.UserRoleAdmin
		}
		// This fixture tests storage authorization, not password authentication.
		// Insert only disposable synthetic principals with explicit UUIDs.
		require.NoError(t, r.GormClient.Session(&gorm.Session{SkipHooks: true}).Create(u).Error)
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, name)
		contexts = append(contexts, context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, u.ID))
		users = append(users, u)
	}
	return r, cfg, contexts, users, events
}

func clinicalTestPayload(base, patient, value string) []byte {
	return []byte(fmt.Sprintf(`{
 "resourceType":"Bundle", "type":"collection", "entry":[
 {"fullUrl":%q,"resource":{"resourceType":"Patient","id":%q,"birthDate":"1980-02","name":[{"text":"synthetic-private-patient-name"}]}},
 {"resource":{"resourceType":"Observation","id":"synthetic-observation","status":"final","code":{"text":"synthetic-private-lab"},"subject":{"reference":%q},"effectiveDateTime":"2026-11-01T08:10:00-05:00",%s,"extension":[{"url":"https://example.invalid/unknown","valueString":"synthetic-private-extension العربية café"}]}}
 ]}`, base+"/Patient/"+patient, patient, "Patient/"+patient, value))
}

const clinicalQuantity = `"valueQuantity":{"value":1.234567890123456789,"unit":"mg"}`

func clinicalTestCount(t *testing.T, r *GormRepository, table string, want int64) {
	t.Helper()
	var count int64
	require.NoError(t, r.GormClient.Table(table).Count(&count).Error)
	require.Equal(t, want, count, table)
}

// Keep the declared maximum reproducible, rather than inferring whole-import
// feasibility from a parser test or a per-record timing extrapolation.
func TestClinicalSnapshotMaximumCollection(t *testing.T) {
	r, _, contexts, _, _ := clinicalTestFixture(t)
	entries := []json.RawMessage{json.RawMessage(`{"resource":{"resourceType":"Patient","id":"synthetic-patient"}}`)}
	for i := range clinicalsnapshot.MaxObservations {
		entries = append(entries, json.RawMessage(fmt.Sprintf(`{"resource":{"resourceType":"Observation","id":"synthetic-%d","status":"final","code":{"text":"Synthetic maximum collection"},"subject":{"reference":"Patient/synthetic-patient"},%s}}`, i, clinicalQuantity)))
	}
	raw, err := json.Marshal(struct {
		ResourceType string            `json:"resourceType"`
		Type         string            `json:"type"`
		Entry        []json.RawMessage `json:"entry"`
	}{"Bundle", "collection", entries})
	require.NoError(t, err)
	result, err := r.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic maximum collection", raw)
	require.NoError(t, err)
	require.Equal(t, clinicalsnapshot.MaxObservations+1, result.Created)
	clinicalTestCount(t, r, "fhir_observation", int64(clinicalsnapshot.MaxObservations))
	original, err := r.GetClinicalSnapshotOriginal(contexts[0], result.SourceID, result.SnapshotID)
	require.NoError(t, err)
	require.Equal(t, raw, original)
}

func TestClinicalSnapshotPersistence(t *testing.T) {
	t.Run("fidelity replay updates ownership and encrypted recovery", func(t *testing.T) {
		r, cfg, contexts, users, events := clinicalTestFixture(t)
		owner := contexts[0]
		raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
		var trace bytes.Buffer
		r.GormClient.Logger = logger.New(log.New(&trace, "", 0), logger.Config{LogLevel: logger.Info, Colorful: false})
		first, err := r.ImportClinicalSnapshot(owner, clinicalTestBase+"/", "Synthetic clinic", raw)
		require.NoError(t, err)
		require.Equal(t, 2, first.Created)
		require.False(t, first.AlreadyImported)
		require.Contains(t, trace.String(), "INSERT INTO", "exercise actual SQL tracing")
		for _, marker := range []string{"synthetic-private-patient-name", "synthetic-private-lab", "synthetic-private-extension", "1.234567890123456789"} {
			require.NotContains(t, trace.String(), marker)
		}
		original, err := r.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
		require.NoError(t, err)
		require.Equal(t, raw, original, "preserve every original byte, not a regenerated FHIR document")
		parsed, err := clinicalsnapshot.Parse(raw, clinicalTestBase)
		require.NoError(t, err)
		for _, entry := range parsed.Entries {
			stored, err := r.GetSharedResource(owner, first.SourceID, entry.Type, entry.ID)
			require.NoError(t, err)
			require.False(t, stored.Editable)
			expected, err := json.Marshal(entry.Raw)
			require.NoError(t, err)
			require.Equal(t, string(expected), string(stored.ResourceRaw), "raw lexemes and unknown fields survive indexed projections")
		}
		// An absent row in a later bounded snapshot is not a provider deletion.
		omitted := &dm.FhirObservation{}
		omitted.SetOriginBase(models.OriginBase{UserID: users[0].ID, SourceID: first.SourceID, SourceResourceType: "Observation", SourceResourceID: "synthetic-omitted"})
		require.NoError(t, omitted.PopulateAndExtractSearchParameters(json.RawMessage(`{"resourceType":"Observation","id":"synthetic-omitted","status":"final","subject":{"reference":"Patient/synthetic-patient"}}`)))
		require.NoError(t, r.GormClient.Create(omitted).Error)
		source, err := r.GetSource(owner, first.SourceID.String())
		require.NoError(t, err)
		require.Equal(t, sp.PlatformTypeManual, source.PlatformType)
		require.Empty(t, source.AccessToken)
		require.Empty(t, source.DynamicClientJWKS)
		grant, err := r.SetSourceGrant(owner, first.SourceID, users[1].Username, models.SourceAccessEdit, nil)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, grant.ID)
		for _, ctx := range contexts[1:] {
			_, err = r.ListClinicalSnapshots(ctx, first.SourceID, 20, 0)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
			_, err = r.GetClinicalSnapshotOriginal(ctx, first.SourceID, first.SnapshotID)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
		}
		shared, err := r.GetSharedResource(contexts[1], first.SourceID, "Observation", "synthetic-observation")
		require.NoError(t, err)
		require.False(t, shared.Editable, "EDIT grants do not promote imports to local-edit provenance")
		_, err = r.ImportClinicalSnapshot(context.Background(), clinicalTestBase, "Synthetic", raw)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		repeat, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Changed label is not an identity", raw)
		require.NoError(t, err)
		require.True(t, repeat.AlreadyImported)
		require.Equal(t, first.SnapshotID, repeat.SnapshotID)
		changedRaw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", `"valueString":"synthetic-updated-result"`)
		changed, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic clinic", changedRaw)
		require.NoError(t, err)
		require.Equal(t, first.SourceID, changed.SourceID)
		require.Equal(t, 1, changed.Unchanged)
		require.Equal(t, 1, changed.Updated)
		require.Zero(t, changed.Created)
		var indexed dm.FhirObservation
		require.NoError(t, r.GormClient.Where("source_id = ? AND source_resource_id = ?", first.SourceID, "synthetic-observation").First(&indexed).Error)
		require.Contains(t, string(indexed.ValueString), "synthetic-updated-result")
		require.True(t, len(indexed.ValueQuantity) == 0 || string(indexed.ValueQuantity) == "null" || string(indexed.ValueQuantity) == "[]", "removed quantity projections must be cleared")
		historical, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", raw)
		require.NoError(t, err)
		require.True(t, historical.AlreadyImported)
		stored, err := r.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
		require.NoError(t, err)
		require.Contains(t, string(stored.ResourceRaw), "synthetic-updated-result", "old replay cannot roll back newer records")
		rows, err := r.ListClinicalSnapshots(owner, first.SourceID, 20, 0)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		for _, row := range rows {
			require.Empty(t, row.Payload)
			require.Equal(t, models.ClinicalSnapshotOwnerAcquisition, row.Acquisition, "do not attest provider/native origin from owner-supplied metadata")
			require.Equal(t, users[0].ID, row.ImportedByUserID)
		}
		metadata, err := json.Marshal(rows)
		require.NoError(t, err)
		require.NotContains(t, string(metadata), "synthetic-private-")
		clinicalTestCount(t, r, "local_resources", 0)
		clinicalTestCount(t, r, "source_access_audits", 3) // two imports and one explicit grant
		clinicalTestCount(t, r, "related_resources", 1)
		clinicalTestCount(t, r, "fhir_observation", 2)
		require.Zero(t, events.count.Load(), "internal repository must not publish pre-commit success")
		data, err := os.ReadFile(cfg.GetString("database.location"))
		require.NoError(t, err)
		require.False(t, bytes.HasPrefix(data, []byte("SQLite format 3")))
		require.NotContains(t, string(data), "synthetic-private-patient-name")
		backupParent := t.TempDir()
		require.NoError(t, os.Chmod(backupParent, 0700))
		backup := filepath.Join(backupParent, "recovered")
		_, err = BackupEncryptedSQLite(owner, cfg.GetString("database.location"), backup, clinicalTestKey)
		require.NoError(t, err)
		// Boot a restored copy, never the immutable backup artifact itself.
		restoredPath := filepath.Join(t.TempDir(), "restored.db")
		backupBytes, err := os.ReadFile(filepath.Join(backup, "database.sqlite"))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(restoredPath, backupBytes, 0600))
		recoveredDB, err := newTestRepository(t, clinicalTestConfig(t, restoredPath), logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		recovered := recoveredDB.(*GormRepository)
		restored, err := recovered.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
		require.NoError(t, err)
		require.Equal(t, raw, restored)
		restoredCurrent, err := recovered.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
		require.NoError(t, err)
		require.Contains(t, string(restoredCurrent.ResourceRaw), "synthetic-updated-result")
		immutableBackup, err := os.ReadFile(filepath.Join(backup, "database.sqlite"))
		require.NoError(t, err)
		require.Equal(t, backupBytes, immutableBackup, "restore verification must not mutate its backup artifact")
		require.NoError(t, r.GormClient.Model(&models.ClinicalSnapshotRecord{}).Where("id = ?", first.SnapshotID).UpdateColumn("payload", []byte(`{"corrupt":true}`)).Error)
		_, err = r.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
		require.ErrorIs(t, err, ErrClinicalSnapshotStorage, "never export an original whose stored hash no longer matches")
		require.NoError(t, r.GormClient.Model(&models.ClinicalSnapshotRecord{}).Where("id = ?", first.SnapshotID).UpdateColumn("payload", raw).Error)
		require.NoError(t, r.Close())
		reopenedDB, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		reopened := reopenedDB.(*GormRepository)
		restored, err = reopened.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
		require.NoError(t, err)
		require.Equal(t, raw, restored)
	})

	t.Run("last-write failure and cancellation roll back new and changed imports", func(t *testing.T) {
		r, _, contexts, _, events := clinicalTestFixture(t)
		owner := contexts[0]
		raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
		const callback = "synthetic:clinical-final-failure"
		require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
			if row, ok := tx.Statement.Dest.(*models.SourceAccessAudit); ok && row.Action == "import_clinical_snapshot" {
				tx.AddError(errors.New("synthetic final-write failure"))
			}
		}))
		t.Cleanup(func() { r.GormClient.Callback().Create().Remove(callback) })
		result, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", raw)
		require.ErrorIs(t, err, ErrClinicalSnapshotStorage)
		require.Nil(t, result)
		for _, table := range []string{"source_credentials", "fhir_patient", "fhir_observation", "related_resources", "clinical_snapshot_records", "source_access_audits"} {
			clinicalTestCount(t, r, table, 0)
		}
		require.NoError(t, r.GormClient.Callback().Create().Remove(callback))
		first, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", raw)
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(owner)
		require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
			if row, ok := tx.Statement.Dest.(*models.SourceAccessAudit); ok && row.Action == "import_clinical_snapshot" {
				cancel()
			}
		}))
		changed := clinicalTestPayload(clinicalTestBase, "synthetic-patient", `"valueString":"must-not-be-committed"`)
		result, err = r.ImportClinicalSnapshot(ctx, clinicalTestBase, "Synthetic", changed)
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, result)
		stored, err := r.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
		require.NoError(t, err)
		require.Contains(t, string(stored.ResourceRaw), "1.234567890123456789")
		clinicalTestCount(t, r, "clinical_snapshot_records", 1)
		clinicalTestCount(t, r, "source_access_audits", 1)
		require.Zero(t, events.count.Load())
		result, err = r.ImportClinicalSnapshot(ctx, clinicalTestBase, "Synthetic", raw)
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, result)
	})

	t.Run("distinct identity scopes and deletion of originals", func(t *testing.T) {
		r, _, contexts, _, _ := clinicalTestFixture(t)
		results := []*models.ClinicalSnapshotImportResult{}
		for _, tc := range []struct {
			ctx           context.Context
			base, patient string
		}{
			{contexts[0], clinicalTestBase, "synthetic-patient"},
			{contexts[1], clinicalTestBase, "synthetic-patient"},
			{contexts[0], "https://other.example.invalid/fhir", "synthetic-patient"},
			{contexts[0], clinicalTestBase, "different-patient"},
		} {
			result, err := r.ImportClinicalSnapshot(tc.ctx, tc.base, "Synthetic", clinicalTestPayload(tc.base, tc.patient, clinicalQuantity))
			require.NoError(t, err)
			for _, prior := range results {
				require.NotEqual(t, prior.SourceID, result.SourceID)
			}
			results = append(results, result)
		}
		_, err := r.DeleteSource(contexts[1], results[0].SourceID.String())
		require.NoError(t, err) // legacy endpoint is a scoped no-op for foreign owners
		clinicalTestCount(t, r, "clinical_snapshot_records", 4)
		const callback = "synthetic:clinical-source-delete-failure"
		require.NoError(t, r.GormClient.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
			if tx.Statement.Table == "source_credentials" {
				tx.AddError(errors.New("synthetic final source-delete failure"))
			}
		}))
		t.Cleanup(func() { r.GormClient.Callback().Delete().Remove(callback) })
		_, err = r.DeleteSource(contexts[0], results[0].SourceID.String())
		require.Error(t, err)
		clinicalTestCount(t, r, "clinical_snapshot_records", 4)
		clinicalTestCount(t, r, "fhir_observation", 4)
		require.NoError(t, r.GormClient.Callback().Delete().Remove(callback))
		_, err = r.DeleteSource(contexts[0], results[0].SourceID.String())
		require.NoError(t, err)
		clinicalTestCount(t, r, "clinical_snapshot_records", 3)
		_, err = r.GetClinicalSnapshotOriginal(contexts[0], results[0].SourceID, results[0].SnapshotID)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		const accountFailure = "synthetic:clinical-account-delete-failure"
		require.NoError(t, r.GormClient.Callback().Delete().Before("gorm:delete").Register(accountFailure, func(tx *gorm.DB) {
			if tx.Statement.Table == "users" {
				tx.AddError(errors.New("synthetic final account-delete failure"))
			}
		}))
		t.Cleanup(func() { r.GormClient.Callback().Delete().Remove(accountFailure) })
		require.Error(t, r.DeleteCurrentUser(contexts[0]))
		clinicalTestCount(t, r, "clinical_snapshot_records", 3)
		clinicalTestCount(t, r, "fhir_observation", 3)
		require.NoError(t, r.GormClient.Callback().Delete().Remove(accountFailure))
		require.NoError(t, r.DeleteCurrentUser(contexts[0]))
		clinicalTestCount(t, r, "clinical_snapshot_records", 1)
		original, err := r.GetClinicalSnapshotOriginal(contexts[1], results[1].SourceID, results[1].SnapshotID)
		require.NoError(t, err)
		require.NotEmpty(t, original)
	})

	t.Run("revalidation and identity conflict deny all writes", func(t *testing.T) {
		r, _, contexts, _, _ := clinicalTestFixture(t)
		owner := contexts[0]
		raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
		for _, bad := range [][]byte{[]byte(`{}`), bytes.Replace(raw, []byte(`Patient/synthetic-patient`), []byte(`Patient/wrong-patient`), 1), append(bytes.Clone(raw), []byte(` {}`)...), []byte(strings.Repeat(" ", clinicalsnapshot.MaxBytes+1))} {
			_, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", bad)
			require.ErrorIs(t, err, ErrSourceAccessInput)
		}
		clinicalTestCount(t, r, "source_credentials", 0)
		endpoint := uuid.NewSHA1(uuid.NameSpaceURL, []byte("chartrove-direct-smart:"+clinicalTestBase))
		conflict := &models.SourceCredential{EndpointID: endpoint, Patient: "synthetic-patient", PlatformType: sp.PlatformTypeFasten, AccessToken: "synthetic-existing-secret"}
		require.NoError(t, r.CreateSource(owner, conflict))
		_, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", raw)
		require.ErrorIs(t, err, ErrClinicalSnapshotConflict)
		preserved, err := r.GetSource(owner, conflict.ID.String())
		require.NoError(t, err)
		require.Equal(t, "synthetic-existing-secret", preserved.AccessToken)
		clinicalTestCount(t, r, "clinical_snapshot_records", 0)
		clinicalTestCount(t, r, "fhir_observation", 0)
	})

	t.Run("legacy Patient-only source reuse and malformed Patient refusal", func(t *testing.T) {
		r, _, contexts, users, _ := clinicalTestFixture(t)
		owner := contexts[0]
		endpoint := uuid.NewSHA1(uuid.NameSpaceURL, []byte("chartrove-direct-smart:"+clinicalTestBase))
		source := &models.SourceCredential{EndpointID: endpoint, Patient: "synthetic-patient", PlatformType: sp.PlatformTypeManual, Display: "Owner-chosen existing label"}
		require.NoError(t, r.CreateSource(owner, source))
		patient := &dm.FhirPatient{}
		patient.SetOriginBase(models.OriginBase{UserID: users[0].ID, SourceID: source.ID, SourceResourceType: "Patient", SourceResourceID: source.Patient})
		require.NoError(t, patient.PopulateAndExtractSearchParameters(json.RawMessage(`{"resourceType":"Patient","id":"synthetic-patient"}`)))
		require.NoError(t, r.GormClient.Create(patient).Error)
		raw := clinicalTestPayload(clinicalTestBase, source.Patient, clinicalQuantity)
		result, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "New profile label", raw)
		require.NoError(t, err)
		require.Equal(t, source.ID, result.SourceID)
		require.Equal(t, 1, result.Created)
		require.Equal(t, 1, result.Updated)
		preserved, err := r.GetSource(owner, source.ID.String())
		require.NoError(t, err)
		require.Equal(t, source.Display, preserved.Display)
		require.NoError(t, r.GormClient.Model(&dm.FhirPatient{}).Where("source_id = ?", source.ID).UpdateColumn("source_resource_id", "wrong-patient").Error)
		changed := clinicalTestPayload(clinicalTestBase, source.Patient, `"valueString":"must-not-overwrite-conflict"`)
		_, err = r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", changed)
		require.ErrorIs(t, err, ErrClinicalSnapshotConflict)
		clinicalTestCount(t, r, "clinical_snapshot_records", 1)
		clinicalTestCount(t, r, "fhir_patient", 1)
		stored, err := r.GetSharedResource(owner, source.ID, "Observation", "synthetic-observation")
		require.NoError(t, err)
		require.Contains(t, string(stored.ResourceRaw), "1.234567890123456789")
	})

	t.Run("concurrent account deletion cannot orphan new originals", func(t *testing.T) {
		r, cfg, contexts, _, _ := clinicalTestFixture(t)
		secondDB, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		second := secondDB.(*GormRepository)
		start := make(chan struct{})
		imported, deleted := make(chan error, 1), make(chan error, 1)
		go func() {
			<-start
			_, err := second.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic", clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity))
			imported <- err
		}()
		go func() { <-start; deleted <- r.DeleteCurrentUser(contexts[0]) }()
		close(start)
		require.NoError(t, <-deleted)
		err = <-imported
		require.True(t, err == nil || errors.Is(err, ErrSourceAccessDenied), "import must either precede deletion or fail closed: %v", err)
		for _, table := range []string{"clinical_snapshot_records", "source_credentials", "fhir_patient", "fhir_observation", "related_resources"} {
			clinicalTestCount(t, r, table, 0)
		}
	})

	t.Run("two database connections serialize exact concurrent replay", func(t *testing.T) {
		r, cfg, contexts, _, _ := clinicalTestFixture(t)
		secondDB, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
		require.NoError(t, err)
		second := secondDB.(*GormRepository)
		start := make(chan struct{})
		type outcome struct {
			result *models.ClinicalSnapshotImportResult
			err    error
		}
		out := make(chan outcome, 2)
		raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
		for _, repo := range []*GormRepository{r, second} {
			go func() {
				<-start
				result, err := repo.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic", raw)
				out <- outcome{result, err}
			}()
		}
		close(start)
		created, replayed := 0, 0
		for range 2 {
			got := <-out
			require.NoError(t, got.err)
			created += got.result.Created
			if got.result.AlreadyImported {
				replayed++
			}
		}
		require.Equal(t, 2, created)
		require.Equal(t, 1, replayed)
		clinicalTestCount(t, r, "source_credentials", 1)
		clinicalTestCount(t, r, "clinical_snapshot_records", 1)
		clinicalTestCount(t, r, "source_access_audits", 1)
	})
}
