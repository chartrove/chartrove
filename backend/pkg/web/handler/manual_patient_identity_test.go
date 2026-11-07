package handler

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	databaseModel "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	sourcePkg "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/stretchr/testify/require"
)

func TestManualPatientIdentityPreservesCanonicalIDs(t *testing.T) {
	for _, id := range []string{"aa001", "001", "Patient", "a", "erXu", "A.B-9", strings.Repeat("a", 64)} {
		t.Run(id, func(t *testing.T) {
			for _, document := range []string{
				fmt.Sprintf(`{"resourceType":"Patient","id":%q}`, id),
				smartPatientBundle(id, "Synthetic"),
				fmt.Sprintf("{\"resourceType\":\"Observation\",\"id\":\"synthetic\"}\n{\"resourceType\":\"Patient\",\"id\":%q}\n{\"resourceType\":\"Patient\",\"id\":%q}\n", id, id),
			} {
				f, err := os.CreateTemp(t.TempDir(), "synthetic-identity-*.json")
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, f.Close()) })
				_, err = io.WriteString(f, document)
				require.NoError(t, err)
				actual, err := manualPatientID(f)
				require.NoError(t, err)
				require.Equal(t, id, actual)
				position, err := f.Seek(0, io.SeekCurrent)
				require.NoError(t, err)
				require.Zero(t, position)
				after, err := io.ReadAll(f)
				require.NoError(t, err)
				require.Equal(t, document, string(after), "the importer must receive unchanged bytes")
			}
		})
	}
}

func TestManualPatientIdentityRejectsUnknownOrMixedPatients(t *testing.T) {
	for _, document := range []string{
		``, `[]`, `{"resourceType":"Observation","id":"synthetic"}`,
		`{"resourceType":"Patient"}`, `{"resourceType":"Patient","id":7}`,
		`{"resourceType":"Patient","id":"Patient/synthetic"}`,
		fmt.Sprintf(`{"resourceType":"Patient","id":%q}`, strings.Repeat("a", 65)),
		`{"resourceType":"Bundle","entry":[{"resource":{"resourceType":"Patient","id":"aa001"}},{"resource":{"resourceType":"Patient","id":"001"}}]}`,
		"{\"resourceType\":\"Patient\",\"id\":\"aa001\"}\n{\"resourceType\":\"Patient\",\"id\":\"001\"}",
		`{"resourceType":"Patient","id":"aa001"} trailing`,
	} {
		f, err := os.CreateTemp(t.TempDir(), "synthetic-invalid-*.json")
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, f.Close()) })
		_, err = io.WriteString(f, document)
		require.NoError(t, err)
		id, err := manualPatientID(f)
		require.Error(t, err)
		require.Empty(t, id)
		require.Equal(t, "manual import must contain one distinct Patient with a valid FHIR ID", err.Error())
		position, err := f.Seek(0, io.SeekCurrent)
		require.NoError(t, err)
		require.Zero(t, position)
	}
}

func TestManualImportRefusesUnprovenLegacySourceReuse(t *testing.T) {
	h := newSMARTImportHarness(t)
	first := readSMARTImportResponse(t, h.post(t, nil, smartPatientBundle("aa001", "First synthetic patient")))
	db := h.repo.(*database.GormRepository)
	// Reproduce the old lossy metadata in this disposable database only.
	require.NoError(t, db.GormClient.Model(&models.SourceCredential{}).Where("id = ?", first.Source.ID).Update("patient", "001").Error)
	before, err := h.repo.ListResources(h.ctx, models.ListResourceQueryOptions{SourceID: first.Source.ID.String(), SourceResourceType: "Patient"})
	require.NoError(t, err)
	require.Len(t, before, 1)
	sourcesBefore, err := h.repo.GetSources(h.ctx)
	require.NoError(t, err)
	w := h.post(t, nil, smartPatientBundle("001", "Second synthetic patient"))
	require.Equal(t, 409, w.Code)
	require.NotContains(t, w.Body.String(), "aa001")
	unchanged, err := h.repo.ListResources(h.ctx, models.ListResourceQueryOptions{SourceID: first.Source.ID.String(), SourceResourceType: "Patient"})
	require.NoError(t, err)
	require.Equal(t, before, unchanged)
	sourcesAfter, err := h.repo.GetSources(h.ctx)
	require.NoError(t, err)
	require.Equal(t, sourcesBefore, sourcesAfter, "a legacy conflict must be rejected before credential writes")

	// These fields are *time.Time, not GORM's automatic soft-delete type.
	// Explicitly marked sources/Patients must remain refused and marked.
	deletedAt := time.Now().UTC()
	require.NoError(t, db.GormClient.Model(&models.SourceCredential{}).Where("id = ?", first.Source.ID).
		Updates(map[string]interface{}{"patient": "aa001", "deleted_at": deletedAt}).Error)
	require.Equal(t, 409, h.post(t, nil, smartPatientBundle("aa001", "Deleted source")).Code)
	var markedSource models.SourceCredential
	require.NoError(t, db.GormClient.Where("id = ?", first.Source.ID).First(&markedSource).Error)
	require.NotNil(t, markedSource.DeletedAt)
	require.NoError(t, db.GormClient.Model(&models.SourceCredential{}).Where("id = ?", first.Source.ID).Update("deleted_at", nil).Error)
	patientTable, err := databaseModel.GetTableNameByResourceType("Patient")
	require.NoError(t, err)
	require.NoError(t, db.GormClient.Table(patientTable).Where("user_id = ? AND source_id = ? AND source_resource_id = ?", first.Source.UserID, first.Source.ID, "aa001").Update("deleted_at", deletedAt).Error)
	require.Equal(t, 409, h.post(t, nil, smartPatientBundle("aa001", "Deleted Patient")).Code)
	var markedPatient models.ResourceBase
	require.NoError(t, db.GormClient.Table(patientTable).Where("user_id = ? AND source_id = ? AND source_resource_id = ?", first.Source.UserID, first.Source.ID, "aa001").First(&markedPatient).Error)
	require.NotNil(t, markedPatient.DeletedAt)

	unproven := models.SourceCredential{PlatformType: sourcePkg.PlatformTypeManual, Patient: "synthetic-incomplete"}
	require.NoError(t, h.repo.CreateSource(h.ctx, &unproven))
	w = h.post(t, nil, smartPatientBundle("synthetic-incomplete", "Unproven synthetic identity"))
	require.Equal(t, 409, w.Code, "an empty existing source does not prove its historical patient identity")
	rows, err := h.repo.ListResources(h.ctx, models.ListResourceQueryOptions{SourceID: unproven.ID.String(), SourceResourceType: "Patient"})
	require.NoError(t, err)
	require.Empty(t, rows, "a refused import must not fill an unproven source")
}

func TestManualImportRejectsMixedPatientsBeforeSourceCreation(t *testing.T) {
	h := newSMARTImportHarness(t)
	before, err := h.repo.GetSources(h.ctx)
	require.NoError(t, err)
	for _, document := range []string{
		`{"resourceType":"Bundle","entry":[{"resource":{"resourceType":"Patient","id":"aa001"}},{"resource":{"resourceType":"Patient","id":"001"}}]}`,
		"{\"resourceType\":\"Patient\",\"id\":\"aa001\"}\n{\"resourceType\":\"Patient\",\"id\":\"001\"}",
	} {
		w := h.post(t, nil, document)
		require.Equal(t, 400, w.Code)
		after, err := h.repo.GetSources(h.ctx)
		require.NoError(t, err)
		require.Equal(t, before, after, "mixed patient uploads must not create or change source credentials")
	}
}
