//go:build postgres_integration

package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func clinicalPostgresConfig(t *testing.T) config.Interface {
	t.Helper()
	dsn := disposablePostgres(t)
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "postgres")
	cfg.Set("database.location", dsn.String())
	return cfg
}

func TestClinicalSnapshotPostgresPersistence(t *testing.T) {
	cfg := clinicalPostgresConfig(t)
	// Use the actual application migrations and the same explicit immutable
	// synthetic principals as SQLite, not a reduced hand-built PostgreSQL schema.
	r, _, contexts, users, events := clinicalConfiguredTestFixture(t, cfg)
	var err error
	var trace bytes.Buffer
	r.GormClient.Logger = logger.New(log.New(&trace, "", 0), logger.Config{LogLevel: logger.Info, Colorful: false})
	owner := contexts[0]
	raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
	first, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic clinic", raw)
	require.NoError(t, err)
	require.Equal(t, 2, first.Created)
	require.Zero(t, first.Updated)
	hash := sha256.Sum256(raw)
	require.Equal(t, hex.EncodeToString(hash[:]), first.SHA256)
	original, err := r.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
	require.NoError(t, err)
	require.Equal(t, raw, original, "bytea originals must not be regenerated from indexed JSONB")
	stored, err := r.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
	require.NoError(t, err)
	require.Contains(t, string(stored.ResourceRaw), "1.234567890123456789")
	require.Contains(t, string(stored.ResourceRaw), "synthetic-private-extension")
	require.False(t, stored.Editable)

	repeated, err := r.ImportClinicalSnapshot(owner, clinicalTestBase+"/", "Synthetic", raw)
	require.NoError(t, err)
	require.True(t, repeated.AlreadyImported)
	require.Equal(t, first.SnapshotID, repeated.SnapshotID)
	changedRaw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", `"valueQuantity":{"value":1.234567890123456788,"unit":"mg"}`)
	changed, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", changedRaw)
	require.NoError(t, err)
	require.Equal(t, first.SourceID, changed.SourceID)
	require.Equal(t, 1, changed.Updated, "only the changed Observation is updated")
	require.Equal(t, 1, changed.Unchanged, "the identical Patient stays unchanged on PostgreSQL")
	require.Zero(t, changed.Created)
	stored, err = r.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
	require.NoError(t, err)
	require.Contains(t, string(stored.ResourceRaw), "1.234567890123456788")
	// PostgreSQL JSONB value equality would collapse this recorded scale
	// change; compare its text representation without rounding either value.
	scaledRaw := bytes.Replace(changedRaw, []byte("1.234567890123456788"), []byte("1.2345678901234567880"), 1)
	scaled, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", scaledRaw)
	require.NoError(t, err)
	require.Equal(t, 1, scaled.Updated)
	require.Equal(t, 1, scaled.Unchanged)
	historical, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", raw)
	require.NoError(t, err)
	require.True(t, historical.AlreadyImported)
	stored, err = r.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
	require.NoError(t, err)
	require.Contains(t, string(stored.ResourceRaw), "1.2345678901234567880", "historical acknowledgement cannot roll back current values or recorded scale")

	grant, err := r.SetSourceGrant(owner, first.SourceID, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	for _, foreign := range contexts[1:] {
		_, err = r.ListClinicalSnapshots(foreign, first.SourceID, 20, 0)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		_, err = r.GetClinicalSnapshotOriginal(foreign, first.SourceID, first.SnapshotID)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
	}
	shared, err := r.GetSharedResource(contexts[1], first.SourceID, "Observation", "synthetic-observation")
	require.NoError(t, err)
	require.False(t, shared.Editable)
	require.NoError(t, r.RevokeSourceGrant(owner, first.SourceID, grant.ID))
	_, err = r.GetSharedResource(contexts[1], first.SourceID, "Observation", "synthetic-observation")
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	secondOwner, err := r.ImportClinicalSnapshot(contexts[1], clinicalTestBase, "Synthetic", raw)
	require.NoError(t, err)
	require.NotEqual(t, first.SourceID, secondOwner.SourceID)
	secondOrigin, err := r.ImportClinicalSnapshot(owner, "https://other-clinical.example.invalid/fhir", "Synthetic", clinicalTestPayload("https://other-clinical.example.invalid/fhir", "synthetic-patient", clinicalQuantity))
	require.NoError(t, err)
	require.NotEqual(t, first.SourceID, secondOrigin.SourceID)
	secondPatient, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", clinicalTestPayload(clinicalTestBase, "synthetic-other-patient", clinicalQuantity))
	require.NoError(t, err)
	require.NotEqual(t, first.SourceID, secondPatient.SourceID)
	rows, err := r.ListClinicalSnapshots(owner, first.SourceID, 20, 0)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	for _, snapshot := range []struct {
		id  models.ClinicalSnapshotImportResult
		raw []byte
	}{{*first, raw}, {*changed, changedRaw}, {*scaled, scaledRaw}} {
		original, err := r.GetClinicalSnapshotOriginal(owner, first.SourceID, snapshot.id.SnapshotID)
		require.NoError(t, err)
		require.Equal(t, snapshot.raw, original)
	}
	require.Contains(t, trace.String(), "SELECT CAST(", "exercise the actual PostgreSQL storage normalization query")
	for _, marker := range []string{"synthetic-private-patient-name", "synthetic-private-lab", "synthetic-private-extension", "1.234567890123456789", "1.234567890123456788"} {
		require.NotContains(t, trace.String(), marker, "clinical normalization and storage SQL must remain payload-free")
	}
	require.Zero(t, events.count.Load(), "the repository does not publish HTTP refresh notifications")
}

func TestClinicalSnapshotPostgresRollbackAndReopen(t *testing.T) {
	r, cfg, contexts, _, events := clinicalConfiguredTestFixture(t, clinicalPostgresConfig(t))
	owner := contexts[0]
	raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
	const callback = "synthetic:clinical-pg-failure"
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if row, ok := tx.Statement.Dest.(*models.SourceAccessAudit); ok && row.Action == "import_clinical_snapshot" {
			tx.AddError(errors.New("synthetic final-write refusal"))
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
	changed := clinicalTestPayload(clinicalTestBase, "synthetic-patient", `"valueString":"must-not-be-committed"`)
	// A failed normalization query must propagate and roll back, not panic on
	// a nil SQL Row or manufacture an unchanged result.
	const rowCallback = "synthetic:clinical-pg-normalization-failure"
	require.NoError(t, r.GormClient.Callback().Row().Before("gorm:row").Register(rowCallback, func(tx *gorm.DB) {
		if strings.HasPrefix(tx.Statement.SQL.String(), "SELECT CAST(") {
			tx.AddError(errors.New("synthetic-private-normalization-refusal"))
		}
	}))
	t.Cleanup(func() { r.GormClient.Callback().Row().Remove(rowCallback) })
	result, err = r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", changed)
	require.ErrorIs(t, err, ErrClinicalSnapshotStorage)
	require.NotContains(t, err.Error(), "synthetic-private-normalization-refusal")
	require.Nil(t, result)
	require.NoError(t, r.GormClient.Callback().Row().Remove(rowCallback))
	ctx, cancel := context.WithCancel(owner)
	defer cancel()
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if row, ok := tx.Statement.Dest.(*models.SourceAccessAudit); ok && row.Action == "import_clinical_snapshot" {
			cancel()
		}
	}))
	result, err = r.ImportClinicalSnapshot(ctx, clinicalTestBase, "Synthetic", changed)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, result)
	require.NoError(t, r.GormClient.Callback().Create().Remove(callback))
	clinicalTestCount(t, r, "clinical_snapshot_records", 1)
	clinicalTestCount(t, r, "source_access_audits", 1)
	stored, err := r.GetSharedResource(owner, first.SourceID, "Observation", "synthetic-observation")
	require.NoError(t, err)
	require.Contains(t, string(stored.ResourceRaw), "1.234567890123456789")
	require.NotContains(t, string(stored.ResourceRaw), "must-not-be-committed")
	require.NoError(t, r.Close())
	quiet := logrus.New()
	quiet.SetOutput(io.Discard)
	reopenedDB, err := newTestRepository(t, cfg, quiet, events)
	require.NoError(t, err)
	reopened := reopenedDB.(*GormRepository)
	original, err := reopened.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
	require.NoError(t, err)
	require.Equal(t, raw, original)
	require.NoError(t, reopened.GormClient.Model(&models.ClinicalSnapshotRecord{}).Where("id = ?", first.SnapshotID).UpdateColumn("payload", []byte(`{"corrupt":true}`)).Error)
	_, err = reopened.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
	require.ErrorIs(t, err, ErrClinicalSnapshotStorage)
	require.Zero(t, events.count.Load())
}

func TestClinicalSnapshotPostgresEntryProvenance(t *testing.T) {
	r, _, contexts, _, _ := clinicalConfiguredTestFixture(t, clinicalPostgresConfig(t))
	for _, resourceType := range []string{"Patient", "Observation"} {
		t.Run(resourceType, func(t *testing.T) {
			patient := "synthetic-" + strings.ToLower(resourceType)
			id := patient
			if resourceType == "Observation" {
				id = "synthetic-observation"
			}
			canonical := clinicalTestBase + "/" + resourceType + "/" + id
			raw := clinicalTestPayload(clinicalTestBase, patient, clinicalQuantity)
			if resourceType == "Observation" {
				raw = bytes.Replace(raw, []byte(`{"resource":{"resourceType":"Observation"`), []byte(`{"fullUrl":"`+canonical+`","resource":{"resourceType":"Observation"`), 1)
			}
			first, err := r.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic", raw)
			require.NoError(t, err)
			before, err := r.GetSharedResource(contexts[0], first.SourceID, resourceType, id)
			require.NoError(t, err)
			urn := "urn:uuid:00000000-0000-4000-8000-000000000001"
			changedRaw := bytes.Replace(raw, []byte(canonical), []byte(urn), 1)
			changed, err := r.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic", changedRaw)
			require.NoError(t, err)
			require.Equal(t, 1, changed.Updated)
			require.Equal(t, 1, changed.Unchanged)
			after, err := r.GetSharedResource(contexts[0], first.SourceID, resourceType, id)
			require.NoError(t, err)
			require.Equal(t, before.ResourceRaw, after.ResourceRaw)
			require.NotNil(t, after.SourceUri)
			require.Equal(t, urn, *after.SourceUri)
			replayed, err := r.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic", raw)
			require.NoError(t, err)
			require.True(t, replayed.AlreadyImported)
			after, err = r.GetSharedResource(contexts[0], first.SourceID, resourceType, id)
			require.NoError(t, err)
			require.Equal(t, urn, *after.SourceUri)
			original, err := r.GetClinicalSnapshotOriginal(contexts[0], first.SourceID, changed.SnapshotID)
			require.NoError(t, err)
			require.Equal(t, changedRaw, original)
		})
	}
}

func TestClinicalSnapshotPostgresMaximumCollection(t *testing.T) {
	r, _, contexts, _, _ := clinicalConfiguredTestFixture(t, clinicalPostgresConfig(t))
	raw := clinicalMaximumTestPayload(t)
	started := time.Now()
	var writerStarts atomic.Int32
	const phaseCallback = "synthetic:clinical-pg-maximum-phase"
	require.NoError(t, r.GormClient.Callback().Update().Before("gorm:update").Register(phaseCallback, func(tx *gorm.DB) {
		if tx.Statement.Table == "users" {
			writerStarts.Add(1)
			t.Logf("Maximum PostgreSQL reached writer reservation after %s", time.Since(started))
		}
	}))
	t.Cleanup(func() {
		r.GormClient.Callback().Update().Remove(phaseCallback)
		t.Logf("Maximum PostgreSQL writer reservations reached: %d", writerStarts.Load())
	})
	first, err := r.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic maximum", raw)
	require.NoError(t, err)
	t.Logf("Maximum PostgreSQL import: %d bytes, %s", len(raw), time.Since(started))
	require.Equal(t, clinicalsnapshot.MaxObservations+1, first.Created)
	clinicalTestCount(t, r, "fhir_observation", int64(clinicalsnapshot.MaxObservations))
	changed := bytes.Replace(raw, []byte("1.234567890123456789"), []byte("1.234567890123456788"), 1)
	started = time.Now()
	updated, err := r.ImportClinicalSnapshot(contexts[0], clinicalTestBase, "Synthetic maximum", changed)
	require.NoError(t, err)
	t.Logf("Maximum PostgreSQL changed snapshot: %s", time.Since(started))
	require.Equal(t, 1, updated.Updated)
	require.Equal(t, clinicalsnapshot.MaxObservations, updated.Unchanged)
	require.Zero(t, updated.Created)
	for _, snapshot := range []struct {
		id  models.ClinicalSnapshotImportResult
		raw []byte
	}{{*first, raw}, {*updated, changed}} {
		original, err := r.GetClinicalSnapshotOriginal(contexts[0], first.SourceID, snapshot.id.SnapshotID)
		require.NoError(t, err)
		require.Equal(t, snapshot.raw, original)
	}
}

func TestClinicalSnapshotPostgresConcurrentDeletion(t *testing.T) {
	r, _, contexts, _, _ := clinicalConfiguredTestFixture(t, clinicalPostgresConfig(t))
	owner := contexts[0]
	raw := clinicalTestPayload(clinicalTestBase, "synthetic-patient", clinicalQuantity)
	type outcome struct {
		result *models.ClinicalSnapshotImportResult
		err    error
	}
	start := make(chan struct{})
	results := make(chan outcome, 4)
	var workers sync.WaitGroup
	t.Cleanup(workers.Wait)
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			result, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", raw)
			results <- outcome{result, err}
		}()
	}
	close(start)
	var first *models.ClinicalSnapshotImportResult
	created, replayed := 0, 0
	for range 4 {
		result := <-results
		require.NoError(t, result.err)
		if first == nil {
			first = result.result
		}
		require.Equal(t, first.SourceID, result.result.SourceID)
		require.Equal(t, first.SnapshotID, result.result.SnapshotID)
		created += result.result.Created
		if result.result.AlreadyImported {
			replayed++
		}
	}
	require.Equal(t, 2, created)
	require.Equal(t, 3, replayed)
	const callback = "synthetic:clinical-pg-delete-refusal"
	t.Cleanup(func() { r.GormClient.Callback().Delete().Remove(callback) })
	for _, table := range []string{"source_credentials", "users"} {
		require.NoError(t, r.GormClient.Callback().Delete().Before("gorm:delete").Register(callback, func(tx *gorm.DB) {
			if tx.Statement.Table == table {
				tx.AddError(errors.New("synthetic final-delete refusal"))
			}
		}))
		if table == "users" {
			require.Error(t, r.DeleteCurrentUser(owner))
		} else {
			_, err := r.DeleteSource(owner, first.SourceID.String())
			require.Error(t, err)
		}
		require.NoError(t, r.GormClient.Callback().Delete().Remove(callback))
		for _, preserved := range []string{"source_credentials", "fhir_patient", "fhir_observation", "clinical_snapshot_records"} {
			clinicalTestCount(t, r, preserved, 1)
		}
		original, err := r.GetClinicalSnapshotOriginal(owner, first.SourceID, first.SnapshotID)
		require.NoError(t, err)
		require.Equal(t, raw, original)
	}
	// Competing transactions use actual PostgreSQL owner-row locking. A
	// completed account deletion must not leave a newly imported orphan.
	start = make(chan struct{})
	deletion := make(chan error, 1)
	workers.Add(2)
	go func() { defer workers.Done(); <-start; deletion <- r.DeleteCurrentUser(owner) }()
	go func() {
		defer workers.Done()
		<-start
		result, err := r.ImportClinicalSnapshot(owner, clinicalTestBase, "Synthetic", clinicalTestPayload(clinicalTestBase, "synthetic-patient", `"valueString":"synthetic-racing-update"`))
		results <- outcome{result, err}
	}()
	close(start)
	require.NoError(t, <-deletion)
	result := <-results
	if result.err != nil {
		require.ErrorIs(t, result.err, ErrSourceAccessDenied)
	}
	for _, table := range []string{"source_credentials", "fhir_patient", "fhir_observation", "clinical_snapshot_records", "related_resources"} {
		clinicalTestCount(t, r, table, 0)
	}
}
