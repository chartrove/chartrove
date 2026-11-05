package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/norm"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// These reference query modes exist only in tests. No runtime parser, index,
// endpoint, dependency or clinical text destination is added.
var qualityModes = []string{"literal", "exact_id", "token_phrase", "all_tokens"}

type qualityDocument struct {
	ID, Title, Text string
	Reviewed        bool
}

type qualityQuery struct {
	ID, Family, Text, Rationale string
	PreferredMode               string `json:"preferred_mode"`
	Relevant                    []string
	LiteralSQLite               []string `json:"literal_sqlite"`
	LiteralPostgres             []string `json:"literal_postgres"`
}

type qualityCorpus struct {
	Version int
	Notice  string
	Records []qualityDocument
	Queries []qualityQuery
}

func loadQualityCorpus(t *testing.T) qualityCorpus {
	t.Helper()
	raw, err := os.ReadFile("testdata/retrieval-quality-v1.json")
	require.NoError(t, err)
	var c qualityCorpus
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	require.NoError(t, decoder.Decode(&c))
	require.Equal(t, 1, c.Version)
	require.NotEmpty(t, c.Notice)
	ids := map[string]bool{}
	for _, row := range c.Records {
		require.NotEmpty(t, row.ID)
		require.False(t, ids[row.ID], "duplicate corpus ID")
		ids[row.ID] = true
	}
	queries := map[string]bool{}
	for _, q := range c.Queries {
		require.False(t, queries[q.ID], "duplicate query case")
		queries[q.ID] = true
		require.Contains(t, qualityModes, q.PreferredMode)
		require.NotEmpty(t, q.Rationale)
		for _, labels := range [][]string{q.Relevant, q.LiteralSQLite, q.LiteralPostgres} {
			require.NotNil(t, labels, "empty gold sets must be explicit []")
			seen := map[string]bool{}
			for _, id := range labels {
				require.True(t, ids[id], "unknown gold ID %s", id)
				require.False(t, seen[id], "duplicate gold ID")
				seen[id] = true
			}
		}
	}
	t.Logf("corpus sha256=%x documents=%d queries=%d; authored synthetic labels", sha256.Sum256(raw), len(c.Records), len(c.Queries))
	return c
}

func TestRetrievalQualitySQLite(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "synthetic-quality.db")
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", path)
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-retrieval-quality-key")
	log := logrus.New()
	log.SetOutput(io.Discard)
	db, err := newTestRepository(t, cfg, log, event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	r.GormClient.Logger = logger.Discard
	var version string
	require.NoError(t, r.GormClient.Raw("SELECT sqlite_version()").Scan(&version).Error)
	t.Logf("encrypted SQLite %s", version)
	runRetrievalQuality(t, r)
	require.NoError(t, r.GormClient.Exec("PRAGMA wal_checkpoint(TRUNCATE)").Error)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.False(t, bytes.HasPrefix(raw, []byte("SQLite format 3")))
	require.NotContains(t, string(raw), "hidden-external-canary")
}

type qualityFixture struct {
	r                             *GormRepository
	table                         string
	source, otherSource           uuid.UUID
	owner, delegate, admin, other context.Context
	users                         []models.User
	grant                         *models.SourceGrant
}

func seedQualityCorpus(t *testing.T, r *GormRepository, corpus qualityCorpus, filler int) *qualityFixture {
	t.Helper()
	f := &qualityFixture{r: r, users: []models.User{{Username: "quality-owner"}, {Username: "quality-delegate"}, {Username: "quality-admin", Role: pkg.UserRoleAdmin}, {Username: "quality-other"}}}
	var err error
	f.table, err = dm.GetTableNameByResourceType("DocumentReference")
	require.NoError(t, err)
	contexts := []context.Context{}
	for i := range f.users {
		require.NoError(t, r.GormClient.Create(&f.users[i]).Error)
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, f.users[i].Username)
		contexts = append(contexts, context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, f.users[i].ID))
	}
	f.owner, f.delegate, f.admin, f.other = contexts[0], contexts[1], contexts[2], contexts[3]
	for _, owner := range []int{0, 3} {
		source := models.SourceCredential{UserID: f.users[owner].ID, Display: "Synthetic quality source"}
		require.NoError(t, r.GormClient.Create(&source).Error)
		if owner == 0 {
			f.source = source.ID
		} else {
			f.otherSource = source.ID
		}
		rows := []retrievalEvalRecord{}
		docs := append([]qualityDocument{}, corpus.Records...)
		for i := 0; i < filler; i++ {
			docs = append(docs, qualityDocument{ID: fmt.Sprintf("filler-%05d", i), Title: "Synthetic administrative attachment", Text: strings.Repeat("Dummy invoice filing entry. ", 12), Reviewed: true})
		}
		for _, doc := range docs {
			origin := "synthetic-external"
			if doc.Reviewed {
				origin = documentImportOrigin
			}
			text, err := json.Marshal([]string{doc.Text})
			require.NoError(t, err)
			title := doc.Title
			rows = append(rows, retrievalEvalRecord{OriginBase: models.OriginBase{UserID: f.users[owner].ID, SourceID: source.ID, SourceResourceType: "DocumentReference", SourceResourceID: doc.ID}, SortTitle: &title, SourceUri: &origin, Text: datatypes.JSON(text)})
		}
		require.NoError(t, r.GormClient.Table(f.table).CreateInBatches(&rows, 100).Error)
	}
	// A corrupt/misowned row must not become visible merely by claiming this
	// source ID. Both metadata and transcript are unique leak canaries.
	title, origin := "foreign-owner-canary", documentImportOrigin
	foreign := retrievalEvalRecord{OriginBase: models.OriginBase{UserID: f.users[3].ID, SourceID: f.source, SourceResourceType: "DocumentReference", SourceResourceID: "foreign-owner-canary"}, SortTitle: &title, SourceUri: &origin, Text: datatypes.JSON(`["foreign-owner-canary"]`)}
	require.NoError(t, r.GormClient.Table(f.table).Create(&foreign).Error)
	f.grant, err = r.SetSourceGrant(f.owner, f.source, f.users[1].Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	return f
}

// Explicit semantics: NFC + Unicode simple lowercase; maximal letter/digit
// tokens; no stemming, accent removal, abbreviation expansion, CJK segmentation,
// stopwords, clinical interpretation, operators or fuzzy matching. Match all
// tokens or an adjacent ordered phrase within ONE field, never across fields.
func qualityTokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(norm.NFC.String(s)), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsMark(r)
	})
}

func qualityFieldMatches(field string, words []string, mode string) bool {
	tokens := qualityTokens(field)
	if len(words) == 0 {
		return false
	}
	if mode == "all_tokens" {
		present := map[string]bool{}
		for _, token := range tokens {
			present[token] = true
		}
		for _, word := range words {
			if !present[word] {
				return false
			}
		}
		return true
	}
	for i := 0; i+len(words) <= len(tokens); i++ {
		if strings.Join(tokens[i:i+len(words)], "\x00") == strings.Join(words, "\x00") {
			return true
		}
	}
	return false
}

func (f *qualityFixture) query(ctx context.Context, source uuid.UUID, mode, text string) ([]models.SourceSearchHit, error) {
	if mode == "literal" {
		return f.r.SearchSourceResources(ctx, source, "DocumentReference", text, 0)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := f.r.sharedSource(ctx, source); err != nil {
		return nil, err
	}
	actor, err := f.r.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	// Reference scan, NOT an index candidate or runtime implementation. Fetch
	// current rows using the actual database ownership/grant/deletion predicates.
	rows := []retrievalEvalRecord{}
	q := f.r.sourceResourceQuery(ctx, actor.ID, source, f.table, false).
		Select("source_id, source_resource_type, source_resource_id, sort_title, sort_date, source_uri, text").
		Where("source_resource_type = ?", "DocumentReference").Order("source_resource_id").Limit(1100)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) >= 1100 {
		return nil, fmt.Errorf("synthetic reference scan exceeded its bound")
	}
	if mode != "exact_id" && mode != "token_phrase" && mode != "all_tokens" {
		return nil, fmt.Errorf("unknown evaluation mode")
	}
	words := qualityTokens(text)
	hits := []models.SourceSearchHit{}
	for _, row := range rows {
		match := row.SourceResourceID == text
		if mode != "exact_id" {
			fields := []string{row.SourceResourceID}
			if row.SortTitle != nil {
				fields = append(fields, *row.SortTitle)
			}
			if row.SourceUri != nil && *row.SourceUri == documentImportOrigin {
				var texts []string
				if err := json.Unmarshal(row.Text, &texts); err != nil {
					return nil, err
				}
				fields = append(fields, texts...)
			}
			match = false
			for _, field := range fields {
				if qualityFieldMatches(field, words, mode) {
					match = true
					break
				}
			}
		}
		if match {
			hits = append(hits, models.SourceSearchHit{SourceID: row.SourceID, SourceResourceType: row.SourceResourceType, SourceResourceID: row.SourceResourceID, SortTitle: row.SortTitle, SortDate: row.SortDate})
		}
		if len(hits) == 50 {
			break
		}
	}
	return hits, nil
}

func qualityIDs(t *testing.T, source uuid.UUID, hits []models.SourceSearchHit) []string {
	t.Helper()
	ids := []string{}
	for _, hit := range hits {
		require.Equal(t, source, hit.SourceID)
		require.Equal(t, "DocumentReference", hit.SourceResourceType)
		ids = append(ids, hit.SourceResourceID)
	}
	require.True(t, sort.StringsAreSorted(ids), "all modes preserve ID ordering, not relevance ranking")
	raw, err := json.Marshal(hits)
	require.NoError(t, err)
	for _, field := range []string{"resource_raw", "source_uri", "user_id", "foreign-owner-canary"} {
		require.NotContains(t, string(raw), field)
	}
	return ids
}

type qualityMeasurement struct {
	Case, Family, Mode string
	IDs                []string
	Precision, Recall  float64
	ExactSet           bool
	P50MS, P95MS       float64
	SamplesMS          []float64 `json:",omitempty"`
}

func qualityScore(got, want []string) (float64, float64) {
	matched := 0
	for _, id := range got {
		for _, expected := range want {
			if id == expected {
				matched++
			}
		}
	}
	precision, recall := 1.0, 1.0
	if len(got) != 0 {
		precision = float64(matched) / float64(len(got))
	} else if len(want) != 0 {
		precision = 0
	}
	if len(want) != 0 {
		recall = float64(matched) / float64(len(want))
	} else if len(got) != 0 {
		recall = 0
	}
	return precision, recall
}

func runRetrievalQuality(t *testing.T, r *GormRepository) {
	t.Helper()
	corpus := loadQualityCorpus(t)
	filler, repeats := 0, 1
	measure := os.Getenv("CHARTROVE_RETRIEVAL_QUALITY") == "1"
	if measure {
		filler, repeats = 1000-len(corpus.Records), 28 // 3 warmups, 25 measured per query/mode
	}
	f := seedQualityCorpus(t, r, corpus, filler)
	measurements := []qualityMeasurement{}
	for _, q := range corpus.Queries {
		t.Run(q.ID, func(t *testing.T) {
			for _, mode := range qualityModes {
				var ids []string
				timings := []float64{}
				for i := 0; i < repeats; i++ {
					start := time.Now()
					hits, err := f.query(f.delegate, f.source, mode, q.Text)
					elapsed := float64(time.Since(start).Microseconds()) / 1000
					require.NoError(t, err)
					current := qualityIDs(t, f.source, hits)
					if i > 0 {
						require.Equal(t, ids, current, "nondeterministic IDs")
					}
					ids = current
					if measure && i >= 3 {
						timings = append(timings, elapsed)
					}
				}
				if mode == "literal" {
					gold := q.LiteralSQLite
					if r.GormClient.Dialector.Name() == "postgres" {
						gold = q.LiteralPostgres
					}
					require.Equal(t, gold, ids, "documented baseline contract: %s", q.Rationale)
				}
				if mode == q.PreferredMode {
					require.Equal(t, q.Relevant, ids, "explicit query-intent gold set: %s", q.Rationale)
				}
				precision, recall := qualityScore(ids, q.Relevant)
				m := qualityMeasurement{Case: q.ID, Family: q.Family, Mode: mode, IDs: ids, Precision: precision, Recall: recall, ExactSet: strings.Join(ids, "\x00") == strings.Join(q.Relevant, "\x00")}
				if measure {
					m.SamplesMS = append([]float64{}, timings...)
					sort.Float64s(timings)
					m.P50MS, m.P95MS = timings[12], timings[23]
				}
				measurements = append(measurements, m)
			}
		})
	}
	raw, err := json.Marshal(measurements)
	require.NoError(t, err)
	t.Logf("quality-results engine=%s rows_per_owner=%d measured=%t samples=%d data=%s", r.GormClient.Dialector.Name(), len(corpus.Records)+filler, measure, max(0, repeats-3), raw)
	f.checkLifecycle(t)
}

func (f *qualityFixture) checkLifecycle(t *testing.T) {
	t.Helper()
	for _, mode := range qualityModes {
		for _, ctx := range []context.Context{f.admin, f.other} {
			_, err := f.query(ctx, f.source, mode, "pneumonia")
			require.ErrorIs(t, err, ErrSourceAccessDenied)
		}
		for _, ctx := range []context.Context{f.owner, f.delegate} {
			hits, err := f.query(ctx, f.source, mode, "foreign-owner-canary")
			require.NoError(t, err)
			require.Empty(t, hits)
		}
		hits, err := f.query(f.other, f.otherSource, mode, "LAB-0042")
		require.NoError(t, err)
		require.NotEmpty(t, qualityIDs(t, f.otherSource, hits), "duplicate FHIR ID must resolve in its own source")
	}
	for _, change := range []struct {
		field string
		value interface{}
	}{{"expires_at", time.Now().UTC().Add(-time.Minute)}, {"revoked_at", time.Now().UTC()}} {
		for _, mode := range qualityModes {
			hits, err := f.query(f.delegate, f.source, mode, "LAB-0042")
			require.NoError(t, err)
			require.NotEmpty(t, hits, "grant must work immediately before its independent invalidation")
		}
		require.NoError(t, f.r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", f.grant.ID).Update(change.field, change.value).Error)
		for _, mode := range qualityModes {
			_, err := f.query(f.delegate, f.source, mode, "pneumonia")
			require.ErrorIs(t, err, ErrSourceAccessDenied)
		}
		require.NoError(t, f.r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", f.grant.ID).Update(change.field, nil).Error)
	}
	// Mutate only synthetic current rows. Every invalidation has a positive
	// precondition and is restored before the next independent check.
	for _, mutation := range []struct {
		field          string
		value, restore interface{}
	}{
		{"text", datatypes.JSON(`["replacement text"]`), datatypes.JSON(`["pneumonia"]`)},
		{"source_uri", "synthetic-external", documentImportOrigin},
		{"deleted_at", time.Now().UTC(), nil},
	} {
		row := f.r.GormClient.Table(f.table).Where("source_id = ? AND user_id = ? AND source_resource_id = ?", f.source, f.users[0].ID, "doc-01")
		for _, mode := range []string{"literal", "token_phrase", "all_tokens"} {
			hits, err := f.query(f.delegate, f.source, mode, "pneumonia")
			require.NoError(t, err)
			require.Contains(t, qualityIDs(t, f.source, hits), "doc-01")
		}
		require.NoError(t, row.Update(mutation.field, mutation.value).Error)
		for _, mode := range []string{"literal", "token_phrase", "all_tokens"} {
			hits, err := f.query(f.delegate, f.source, mode, "pneumonia")
			require.NoError(t, err)
			require.NotContains(t, qualityIDs(t, f.source, hits), "doc-01")
		}
		require.NoError(t, row.Update(mutation.field, mutation.restore).Error)
	}
	for _, mode := range qualityModes {
		hits, err := f.query(f.delegate, f.source, mode, "LAB-0042")
		require.NoError(t, err)
		require.NotEmpty(t, hits)
	}
	require.NoError(t, f.r.GormClient.Model(&models.SourceCredential{}).Where("id = ?", f.source).Update("deleted_at", time.Now().UTC()).Error)
	for _, mode := range qualityModes {
		_, err := f.query(f.delegate, f.source, mode, "LAB-0042")
		require.ErrorIs(t, err, ErrSourceAccessDenied)
	}
	require.NoError(t, f.r.GormClient.Model(&models.SourceCredential{}).Where("id = ?", f.source).Update("deleted_at", nil).Error)
	for _, mode := range qualityModes {
		hits, err := f.query(f.delegate, f.source, mode, "LAB-0042")
		require.NoError(t, err)
		require.NotEmpty(t, hits)
	}
	require.NoError(t, f.r.GormClient.Delete(&f.users[1]).Error)
	replacement := models.User{Username: f.users[1].Username}
	require.NoError(t, f.r.GormClient.Create(&replacement).Error)
	require.NotEqual(t, f.users[1].ID, replacement.ID)
	for _, mode := range qualityModes {
		_, err := f.query(f.delegate, f.source, mode, "LAB-0042")
		require.EqualError(t, err, "could not retrieve current user: "+gorm.ErrRecordNotFound.Error())
	}
	t.Log("quality lifecycle PASS: native UUID/source/owner isolation, administrator denial, expiry/revocation, current text/provenance, record/source deletion and actor replacement")
}
