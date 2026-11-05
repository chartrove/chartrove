package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSourceSearchModesSQLite(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-search-modes.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-mode-key")
	log := logrus.New()
	log.SetOutput(io.Discard)
	db, err := newTestRepository(t, cfg, log, event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	r.GormClient.Logger = logger.Discard
	runSourceSearchModeCorpus(t, r)
}

func runSourceSearchModeCorpus(t *testing.T, r *GormRepository) {
	t.Helper()
	corpus := loadQualityCorpus(t)
	filler := 0
	if os.Getenv("CHARTROVE_SEARCH_MODE_ACCEPTANCE") != "" {
		filler = 1000 - len(corpus.Records)
	}
	if os.Getenv("CHARTROVE_SEARCH_MODE_ACCEPTANCE") == "long" {
		corpus = longQualityCorpus(t, corpus)
		filler = 1000 - len(corpus.Records)
	}
	f := seedQualityCorpus(t, r, corpus, filler)
	// Frozen accepted reference IDs are independent of the production methods.
	raw, err := os.ReadFile("../../../docs/maintenance/retrieval-quality-2026-11-05.json")
	require.NoError(t, err)
	var receipt struct {
		Engines map[string][]qualityMeasurement
	}
	require.NoError(t, json.Unmarshal(raw, &receipt))
	for _, q := range corpus.Queries {
		t.Run(q.ID, func(t *testing.T) {
			for _, mode := range []string{"literal", "exact_id"} {
				var expected []string
				for _, row := range receipt.Engines[r.GormClient.Dialector.Name()] {
					if row.Case == q.ID && row.Mode == mode {
						expected = row.IDs
					}
				}
				require.NotNil(t, expected)
				for _, ctx := range []context.Context{f.owner, f.delegate} {
					hits, err := actualSearchMode(r, ctx, f.source, mode, q.Text, 0)
					require.NoError(t, err)
					require.Equal(t, expected, qualityIDs(t, f.source, hits), "mode=%s", mode)
				}
			}
		})
	}
	if os.Getenv("CHARTROVE_SEARCH_MODE_ACCEPTANCE") != "" {
		measureAcceptedSearchModes(t, f, corpus)
	}
	checkSearchModeLifecycle(t, f)
}

// Opt-in actual implementation measurement. Three interleaved blocks retain all
// samples; aggregate family p95 over all blocks, never select the fastest run.
func measureAcceptedSearchModes(t *testing.T, f *qualityFixture, corpus qualityCorpus) {
	t.Helper()
	type measurement struct {
		Case, Family, Mode string
		Block              int
		Samples            []float64
	}
	data := []measurement{}
	selected := map[string]bool{"exact-resource-id": true, "phrase-spacing": true, "and-distant": true, "unicode-case": true, "cjk-prefix": true, "literal-percent": true, "literal-quotes": true, "phrase-denial": true, "term-contradiction": true, "unreviewed-body": true, "external-title": true, "reviewed-body": true, "electrolyte-short": true, "missing-term": true}
	modes := []string{"baseline", "exact_id"}
	families := map[string]map[string][]float64{}
	for block := 0; block < 3; block++ {
		for _, q := range corpus.Queries {
			if os.Getenv("CHARTROVE_SEARCH_MODE_ACCEPTANCE") == "long" && !selected[q.ID] {
				continue
			}
			samples := map[string][]float64{}
			for i := 0; i < 28; i++ {
				for step := range modes {
					mode := modes[(step+i+block)%len(modes)]
					start := time.Now()
					var err error
					if mode == "baseline" {
						_, err = f.r.SearchSourceResources(f.delegate, f.source, "DocumentReference", q.Text, 0)
					} else {
						_, err = f.r.SearchSourceResourceID(f.delegate, f.source, "DocumentReference", q.Text, 0)
					}
					ms := float64(time.Since(start).Microseconds()) / 1000
					require.NoError(t, err)
					if i >= 3 {
						samples[mode] = append(samples[mode], ms)
					}
				}
			}
			if families[q.Family] == nil {
				families[q.Family] = map[string][]float64{}
			}
			for _, mode := range modes {
				data = append(data, measurement{Case: q.ID, Family: q.Family, Mode: mode, Block: block, Samples: samples[mode]})
				families[q.Family][mode] = append(families[q.Family][mode], samples[mode]...)
			}
		}
	}
	raw, err := json.Marshal(data)
	require.NoError(t, err)
	t.Logf("actual-mode-acceptance engine=%s profile=%s data=%s", f.r.GormClient.Dialector.Name(), os.Getenv("CHARTROVE_SEARCH_MODE_ACCEPTANCE"), raw)
	p95 := func(values []float64) float64 {
		ordered := append([]float64{}, values...)
		sort.Float64s(ordered)
		return ordered[(95*len(ordered)+99)/100-1]
	}
	caseSamples := map[string][]float64{}
	for _, entry := range data {
		if entry.Mode == "exact_id" {
			caseSamples[entry.Case] = append(caseSamples[entry.Case], entry.Samples...)
		}
	}
	for name, samples := range caseSamples {
		require.LessOrEqual(t, p95(samples), 250.0, "case=%s", name)
	}
	for family, methods := range families {
		baseline := p95(methods["baseline"])
		for _, mode := range []string{"exact_id"} {
			value := p95(methods[mode])
			t.Logf("actual-mode-family %s/%s control_p95_ms=%.3f candidate_p95_ms=%.3f ratio=%.3f", family, mode, baseline, value, value/baseline)
			require.LessOrEqual(t, value, 250.0)
			require.LessOrEqual(t, value, baseline*1.25, fmt.Sprintf("family=%s mode=%s", family, mode))
		}
	}
}

func checkSearchModeLifecycle(t *testing.T, f *qualityFixture) {
	t.Helper()
	query := func(ctx context.Context, mode, text string) []string {
		t.Helper()
		hits, err := actualSearchMode(f.r, ctx, f.source, mode, text, 0)
		require.NoError(t, err)
		return qualityIDs(t, f.source, hits)
	}
	for _, mode := range []string{"literal", "exact_id"} {
		for _, ctx := range []context.Context{f.admin, f.other} {
			hits, err := actualSearchMode(f.r, ctx, f.source, mode, "LAB-0042", 0)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
			require.Empty(t, hits)
		}
		require.Empty(t, query(f.delegate, mode, "foreign-owner-canary"))
		hits, err := actualSearchMode(f.r, f.other, f.otherSource, mode, "LAB-0042", 0)
		require.NoError(t, err)
		require.NotEmpty(t, qualityIDs(t, f.otherSource, hits))
	}

	for _, text := range []string{"", "  ", strings.Repeat("界", 129), string([]byte{0xff}), "\x00"} {
		_, err := f.r.SearchSourceResourceID(f.owner, f.source, "DocumentReference", text, 0)
		require.ErrorIs(t, err, ErrSourceAccessInput)
	}
	for _, page := range []int{-1, 100001} {
		_, err := f.r.SearchSourceResourceID(f.owner, f.source, "DocumentReference", "LAB-0042", page)
		require.ErrorIs(t, err, ErrSourceAccessInput)
	}
	hits, err := f.r.SearchSourceResourceID(f.owner, f.source, "DocumentReference", "LAB-0042", 1)
	require.NoError(t, err)
	require.Empty(t, hits)
	ctx, cancel := context.WithCancel(f.delegate)
	cancel()
	hits, err = f.r.SearchSourceResourceID(ctx, f.source, "DocumentReference", "LAB-0042", 0)
	require.Error(t, err)
	require.Empty(t, hits)
	wrong := context.WithValue(f.owner, pkg.ContextKeyTypeAuthUserID, f.users[1].ID)
	hits, err = f.r.SearchSourceResourceID(wrong, f.source, "DocumentReference", "LAB-0042", 0)
	require.Error(t, err)
	require.Empty(t, hits)
	for _, id := range []string{"synthetic-É", "synthetic-E\u0301", "synthetic-界"} {
		row := retrievalEvalRecord{OriginBase: models.OriginBase{UserID: f.users[0].ID, SourceID: f.source, SourceResourceType: "DocumentReference", SourceResourceID: id}}
		require.NoError(t, f.r.GormClient.Table(f.table).Create(&row).Error)
		require.Equal(t, []string{id}, query(f.delegate, "exact_id", id), "IDs are not Unicode-normalized")
	}
	for _, change := range []struct {
		field string
		value interface{}
	}{{"expires_at", time.Now().UTC().Add(-time.Minute)}, {"revoked_at", time.Now().UTC()}} {
		require.Equal(t, []string{"LAB-0042"}, query(f.delegate, "exact_id", "LAB-0042"))
		require.NoError(t, f.r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", f.grant.ID).Update(change.field, change.value).Error)
		_, err := f.r.SearchSourceResourceID(f.delegate, f.source, "DocumentReference", "LAB-0042", 0)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		require.NoError(t, f.r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", f.grant.ID).Update(change.field, nil).Error)
	}
	// Search IDs remain metadata-only as current text/provenance changes. The
	// literal body match must disappear; exact ID still identifies the same row.
	row := f.r.GormClient.Table(f.table).Where("source_id = ? AND user_id = ? AND source_resource_id = ?", f.source, f.users[0].ID, "doc-01")
	for _, change := range []struct {
		field          string
		value, restore interface{}
	}{{"text", datatypes.JSON(`["replacement"]`), datatypes.JSON(`["pneumonia"]`)}, {"source_uri", "synthetic-external", documentImportOrigin}} {
		require.Contains(t, query(f.delegate, "literal", "pneumonia"), "doc-01")
		require.NoError(t, row.Update(change.field, change.value).Error)
		require.NotContains(t, query(f.delegate, "literal", "pneumonia"), "doc-01")
		require.Equal(t, []string{"doc-01"}, query(f.delegate, "exact_id", "doc-01"))
		require.NoError(t, row.Update(change.field, change.restore).Error)
	}
	require.Equal(t, []string{"doc-01"}, query(f.delegate, "exact_id", "doc-01"))
	require.NoError(t, row.Update("deleted_at", time.Now().UTC()).Error)
	require.Empty(t, query(f.delegate, "exact_id", "doc-01"))
	require.Equal(t, []string{"LAB-0042"}, query(f.delegate, "exact_id", "LAB-0042"))
	require.NoError(t, f.r.GormClient.Model(&models.SourceCredential{}).Where("id = ?", f.source).Update("deleted_at", time.Now().UTC()).Error)
	_, err = f.r.SearchSourceResourceID(f.delegate, f.source, "DocumentReference", "LAB-0042", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.NoError(t, f.r.GormClient.Model(&models.SourceCredential{}).Where("id = ?", f.source).Update("deleted_at", nil).Error)
	require.Equal(t, []string{"LAB-0042"}, query(f.delegate, "exact_id", "LAB-0042"))
	require.NoError(t, f.r.GormClient.Delete(&f.users[1]).Error)
	replacement := models.User{Username: f.users[1].Username}
	require.NoError(t, f.r.GormClient.Create(&replacement).Error)
	require.NotEqual(t, f.users[1].ID, replacement.ID)
	_, err = f.r.SearchSourceResourceID(f.delegate, f.source, "DocumentReference", "LAB-0042", 0)
	require.EqualError(t, err, "could not retrieve current user: record not found")
	t.Log("actual mode lifecycle PASS: owner/source/UUID isolation, expiry/revocation, current provenance/text, record/source/actor deletion; canceled request and bounded inputs")
}

func actualSearchMode(r *GormRepository, ctx context.Context, source uuid.UUID, mode, text string, page int) ([]models.SourceSearchHit, error) {
	if mode == "literal" {
		return r.SearchSourceResources(ctx, source, "DocumentReference", text, page)
	}
	return r.SearchSourceResourceID(ctx, source, "DocumentReference", text, page)
}

func TestSourceExactIDDoesNotTrustCaseInsensitiveCollation(t *testing.T) {
	// A fresh in-memory query fixture deliberately differs from normal BINARY
	// schemas. The main application fixture above verifies encrypted SQLite.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.SourceCredential{}, &models.SourceGrant{}))
	require.NoError(t, db.Exec("CREATE TABLE fhir_observation (user_id TEXT, source_id TEXT, source_resource_type TEXT, source_resource_id TEXT COLLATE NOCASE, sort_title TEXT, sort_date DATETIME, deleted_at DATETIME)").Error)
	owner := models.User{Username: "synthetic-collation-owner"}
	require.NoError(t, db.Create(&owner).Error)
	source := models.SourceCredential{UserID: owner.ID}
	require.NoError(t, db.Create(&source).Error)
	ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, owner.Username)
	ctx = context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, owner.ID)
	r := &GormRepository{GormClient: db}
	for _, id := range []string{"Mixed-ID", "Unicode-É", "Unicode-E\u0301"} {
		require.NoError(t, db.Exec("INSERT INTO fhir_observation (user_id,source_id,source_resource_type,source_resource_id) VALUES (?,?,?,?)", owner.ID, source.ID, "Observation", id).Error)
	}
	for _, id := range []string{"Mixed-ID", "Unicode-É", "Unicode-E\u0301"} {
		hits, err := r.SearchSourceResourceID(ctx, source.ID, "Observation", id, 0)
		require.NoError(t, err)
		require.Len(t, hits, 1)
		require.Equal(t, id, hits[0].SourceResourceID)
	}
	var ordinary int64
	require.NoError(t, db.Table("fhir_observation").Where("source_resource_id = ?", "mixed-id").Count(&ordinary).Error)
	require.EqualValues(t, 1, ordinary, "precondition: ordinary equality ignores case")
	hits, err := r.SearchSourceResourceID(ctx, source.ID, "Observation", "mixed-id", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
}
