package database

import (
	"bytes"
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
	dm "github.com/fastenhealth/fasten-onprem/backend/pkg/models/database"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Evaluation only: no runtime routes, migrations, index maintenance or model calls.
// Deliberately opt in; never point this fixture at an existing database.
func TestRetrievalEvaluationSQLite(t *testing.T) {
	if os.Getenv("CHARTROVE_RETRIEVAL_EVALUATION") != "1" {
		t.Skip("bounded synthetic retrieval evaluation is opt-in")
	}
	for _, size := range []int{1000, 10000} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "synthetic-retrieval.db")
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("database.location", path)
			cfg.Set("database.encryption.enabled", true)
			cfg.Set("database.encryption.key", "synthetic-retrieval-evaluation-key")
			log := logrus.New()
			log.SetOutput(io.Discard)
			db, err := newTestRepository(t, cfg, log, event_bus.NewNoopEventBusServer())
			require.NoError(t, err)
			r := db.(*GormRepository)
			r.GormClient.Logger = logger.Discard
			var version string
			require.NoError(t, r.GormClient.Raw("SELECT sqlite_version()").Scan(&version).Error)
			var options []string
			require.NoError(t, r.GormClient.Raw("PRAGMA compile_options").Scan(&options).Error)
			ftsOptions := []string{}
			for _, o := range options {
				if strings.Contains(o, "FTS") {
					ftsOptions = append(ftsOptions, o)
				}
			}
			sort.Strings(ftsOptions)
			t.Logf("SQLite version=%s FTS compile options=%v encrypted=true", version, ftsOptions)
			methods := []string{}
			for _, module := range []string{"fts4", "fts5"} {
				// Fixed developer-owned identifiers, not caller input. Unsupported modules
				// remain unavailable; this evaluation does not load extensions/change flags.
				ddl := "CREATE VIRTUAL TABLE retrieval_eval_" + module + " USING " + module + "(resource_key, body, notindexed=resource_key, tokenize=unicode61)"
				if module == "fts5" {
					ddl = "CREATE VIRTUAL TABLE retrieval_eval_fts5 USING fts5(resource_key UNINDEXED, body, tokenize='unicode61')"
				}
				err := r.GormClient.Exec(ddl).Error
				if err != nil {
					t.Logf("capability %s unavailable: %v", module, err)
					continue
				}
				methods = append(methods, module)
			}
			runRetrievalEvaluation(t, r, size, methods)
			require.NoError(t, r.GormClient.Exec("PRAGMA wal_checkpoint(TRUNCATE)").Error)
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.False(t, bytes.HasPrefix(data, []byte("SQLite format 3")))
			require.False(t, bytes.Contains(data, []byte("amberrare")), "synthetic index terms must not be plaintext in the database file")
			t.Logf("SQLite final file bytes=%d (includes clinical schema, corpus and all available indexes)", len(data))
		})
	}
}

type retrievalEvalRecord struct {
	models.OriginBase
	SortTitle *string
	SortDate  *time.Time
	SourceUri *string
	Text      datatypes.JSON
}

type retrievalEvalEntry struct {
	ResourceKey string `gorm:"column:resource_key"`
	Body        string `gorm:"column:body"`
}

type retrievalEvalFixture struct {
	plan                          *testing.T
	r                             *GormRepository
	table                         string
	owner, delegate, admin, other context.Context
	source                        uuid.UUID
	grant                         *models.SourceGrant
	rows                          []retrievalEvalRecord
	entries                       []retrievalEvalEntry
}

func seedRetrievalEvaluation(t *testing.T, r *GormRepository, size int) *retrievalEvalFixture {
	t.Helper()
	f := &retrievalEvalFixture{r: r}
	var err error
	f.table, err = dm.GetTableNameByResourceType("DocumentReference")
	require.NoError(t, err)
	users := []models.User{{Username: "synthetic-retrieval-owner"}, {Username: "synthetic-retrieval-delegate"}, {Username: "synthetic-retrieval-admin", Role: pkg.UserRoleAdmin}, {Username: "synthetic-retrieval-other"}}
	contexts := []context.Context{}
	// This fixture exercises database retrieval only, not password authentication.
	for i := range users {
		require.NoError(t, r.GormClient.Create(&users[i]).Error)
		ctx := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, users[i].Username)
		contexts = append(contexts, context.WithValue(ctx, pkg.ContextKeyTypeAuthUserID, users[i].ID))
	}
	f.owner, f.delegate, f.admin, f.other = contexts[0], contexts[1], contexts[2], contexts[3]
	for _, ownerIndex := range []int{0, 3} {
		source := models.SourceCredential{UserID: users[ownerIndex].ID, Display: "Synthetic retrieval source"}
		require.NoError(t, r.GormClient.Create(&source).Error)
		if ownerIndex == 0 {
			f.source = source.ID
		}
		title := "Synthetic reviewed document"
		origin := documentImportOrigin
		for i := 0; i < size; i++ {
			text := "commonword " + strings.Repeat("synthetic document reviewed local provenance evidence ", 8)
			if i%100 == 0 {
				text += " amberrare"
			}
			projection, err := json.Marshal([]string{text})
			require.NoError(t, err)
			f.rows = append(f.rows, retrievalEvalRecord{OriginBase: models.OriginBase{UserID: users[ownerIndex].ID, SourceID: source.ID, SourceResourceType: "DocumentReference", SourceResourceID: fmt.Sprintf("synthetic-doc-%06d", i)}, SortTitle: &title, SourceUri: &origin, Text: datatypes.JSON(projection)})
		}
	}
	require.NoError(t, r.GormClient.Table(f.table).CreateInBatches(&f.rows, 100).Error)
	for _, row := range f.rows {
		var texts []string
		require.NoError(t, json.Unmarshal(row.Text, &texts))
		f.entries = append(f.entries, retrievalEvalEntry{ResourceKey: row.ID.String(), Body: strings.Join(texts, " ")})
	}
	f.grant, err = r.SetSourceGrant(f.owner, f.source, users[1].Username, models.SourceAccessView, nil)
	require.NoError(t, err)
	return f
}

// Candidate index membership is never authorization. Recheck both live source
// predicates and the current reviewed transcript. Even stale text/provenance must
// not produce a hit. Literal single-word queries only: not a production FTS parser.
func (f *retrievalEvalFixture) query(ctx context.Context, method, term string, metadataCheck bool) ([]models.SourceSearchHit, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if method == "baseline" {
		return f.r.SearchSourceResources(ctx, f.source, "DocumentReference", term, 0)
	}
	if metadataCheck {
		if _, err := f.r.sharedSource(ctx, f.source); err != nil {
			return nil, err
		}
	}
	actor, err := f.r.sourceAccessActor(ctx)
	if err != nil {
		return nil, err
	}
	q := f.r.sourceResourceQuery(ctx, actor.ID, f.source, f.table, false).
		Select("source_id, source_resource_type, source_resource_id, sort_title, sort_date").Where("source_resource_type = ? AND source_uri = ?", "DocumentReference", documentImportOrigin)
	switch method {
	case "fts4":
		q = q.Where("id IN (SELECT resource_key FROM retrieval_eval_fts4 WHERE retrieval_eval_fts4 MATCH ?)", `"`+term+`"`)
	case "fts5":
		q = q.Where("id IN (SELECT resource_key FROM retrieval_eval_fts5 WHERE retrieval_eval_fts5 MATCH ?)", `"`+term+`"`)
	case "postgres_gin_materialized":
		q = q.Where("id IN (WITH matched AS MATERIALIZED (SELECT resource_key FROM retrieval_eval_gin WHERE terms @@ plainto_tsquery('simple', ?)) SELECT resource_key FROM matched)", term)
	case "postgres_gin":
		q = q.Where("id IN (SELECT resource_key FROM retrieval_eval_gin WHERE terms @@ plainto_tsquery('simple', ?))", term)
	default:
		return nil, fmt.Errorf("unknown evaluation method")
	}
	pattern := "%" + term + "%" // Fixed alphabetic synthetic words, never runtime user input.
	if f.r.GormClient.Dialector.Name() == "postgres" {
		q = q.Where("EXISTS (SELECT 1 FROM jsonb_array_elements_text(COALESCE(NULLIF(CAST(text AS TEXT), ''), '[]')::jsonb) AS current_text(value) WHERE LOWER(current_text.value) LIKE LOWER(?))", pattern)
	} else {
		q = q.Where("EXISTS (SELECT 1 FROM json_each(COALESCE(NULLIF(text, ''), '[]')) AS current_text WHERE LOWER(current_text.value) LIKE LOWER(?))", pattern)
	}
	rows := []models.SourceSearchHit{}
	q = q.Order("source_resource_id").Limit(50)
	if f.plan != nil {
		f.explainCandidatePlan(f.plan, q, method, term)
		f.plan = nil
	}
	err = q.Find(&rows).Error
	return rows, err
}

func runRetrievalEvaluation(t *testing.T, r *GormRepository, size int, methods []string) {
	t.Helper()
	f := seedRetrievalEvaluation(t, r, size)
	for _, method := range methods {
		start := time.Now()
		var beforePages, pageBytes int64
		if r.GormClient.Dialector.Name() == "sqlite" {
			require.NoError(t, r.GormClient.Raw("PRAGMA page_count").Scan(&beforePages).Error)
			require.NoError(t, r.GormClient.Raw("PRAGMA page_size").Scan(&pageBytes).Error)
		}
		if method == "postgres_gin" {
			require.NoError(t, r.GormClient.Exec("CREATE TABLE retrieval_eval_gin (resource_key uuid PRIMARY KEY, body text, terms tsvector)").Error)
			require.NoError(t, r.GormClient.Table("retrieval_eval_gin").CreateInBatches(&f.entries, 100).Error)
			require.NoError(t, r.GormClient.Exec("UPDATE retrieval_eval_gin SET terms = to_tsvector('simple', body)").Error)
			require.NoError(t, r.GormClient.Exec("CREATE INDEX retrieval_eval_gin_terms ON retrieval_eval_gin USING GIN (terms)").Error)
			require.NoError(t, r.GormClient.Exec("ANALYZE retrieval_eval_gin").Error)
			var bytes int64
			require.NoError(t, r.GormClient.Raw("SELECT pg_total_relation_size('retrieval_eval_gin')").Scan(&bytes).Error)
			t.Logf("index method=%s total_relation_bytes=%d", method, bytes)
		} else {
			// Tables were capability-probed before corpus creation; time insertion.
			require.NoError(t, r.GormClient.Table("retrieval_eval_"+method).CreateInBatches(&f.entries, 100).Error)
		}
		if r.GormClient.Dialector.Name() == "sqlite" {
			var afterPages int64
			require.NoError(t, r.GormClient.Raw("PRAGMA page_count").Scan(&afterPages).Error)
			t.Logf("index method=%s allocated_bytes_delta=%d", method, (afterPages-beforePages)*pageBytes)
		}
		t.Logf("build method=%s indexed_rows=%d elapsed_ms=%.3f", method, len(f.entries), float64(time.Since(start).Microseconds())/1000)
	}
	if r.GormClient.Dialector.Name() == "postgres" {
		methods = append(methods, "postgres_gin_materialized")
	}
	methods = append([]string{"baseline"}, methods...)
	for _, method := range methods {
		for _, actor := range []struct {
			name string
			ctx  context.Context
		}{{"owner", f.owner}, {"delegate", f.delegate}} {
			for _, term := range []string{"amberrare", "commonword"} {
				want := 50
				if term == "amberrare" && size/100 < 50 {
					want = size / 100
				}
				durations := []time.Duration{}
				if method != "baseline" {
					f.plan = t
				}
				for i := 0; i < 28; i++ {
					start := time.Now()
					hits, err := f.query(actor.ctx, method, term, true)
					elapsed := time.Since(start)
					require.NoError(t, err)
					require.Len(t, hits, want)
					for _, hit := range hits {
						require.Equal(t, f.source, hit.SourceID)
						require.Equal(t, "DocumentReference", hit.SourceResourceType)
					}
					payload, err := json.Marshal(hits)
					require.NoError(t, err)
					require.NotContains(t, string(payload), "commonword")
					require.NotContains(t, string(payload), "user_id")
					if i >= 3 {
						durations = append(durations, elapsed)
					}
				}
				sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
				t.Logf("query rows_per_owner=%d method=%s actor=%s term=%s samples=25 hits=%d p50_ms=%.3f p95_ms=%.3f", size, method, actor.name, term, want, float64(durations[12].Microseconds())/1000, float64(durations[23].Microseconds())/1000)
			}
		}
	}
	checkRetrievalEvaluationLifecycle(t, f, methods)
}

func checkRetrievalEvaluationLifecycle(t *testing.T, f *retrievalEvalFixture, methods []string) {
	t.Helper()
	for _, method := range methods {
		for _, ctx := range []context.Context{f.admin, f.other} {
			_, err := f.query(ctx, method, "amberrare", true)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
		}
	}
	// Change current text while intentionally retaining stale index terms.
	row := f.rows[0]
	require.NoError(t, f.r.GormClient.Table(f.table).Where("id = ?", row.ID).Update("text", datatypes.JSON(`["synthetic replacement transcript"]`)).Error)
	for _, method := range methods {
		hits, err := f.query(f.owner, method, "amberrare", true)
		require.NoError(t, err)
		for _, h := range hits {
			require.NotEqual(t, row.SourceResourceID, h.SourceResourceID)
		}
	}
	// A stale index cannot turn an unreviewed/external projection into a hit.
	row = f.rows[100]
	require.NoError(t, f.r.GormClient.Table(f.table).Where("id = ?", row.ID).Update("source_uri", "synthetic-unreviewed-origin").Error)
	for _, method := range methods {
		hits, err := f.query(f.owner, method, "amberrare", true)
		require.NoError(t, err)
		for _, h := range hits {
			require.NotEqual(t, row.SourceResourceID, h.SourceResourceID)
		}
	}
	// Deleted clinical rows stay absent even with intact index candidates.
	row = f.rows[200]
	require.NoError(t, f.r.GormClient.Table(f.table).Where("id = ?", row.ID).Update("deleted_at", time.Now().UTC()).Error)
	for _, method := range methods {
		hits, err := f.query(f.owner, method, "amberrare", true)
		require.NoError(t, err)
		for _, h := range hits {
			require.NotEqual(t, row.SourceResourceID, h.SourceResourceID)
		}
	}
	for _, change := range []map[string]interface{}{{"expires_at": time.Now().Add(-time.Minute)}, {"expires_at": nil, "revoked_at": time.Now().UTC()}} {
		require.NoError(t, f.r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", f.grant.ID).Updates(change).Error)
		for _, method := range methods {
			_, err := f.query(f.delegate, method, "amberrare", true)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
			if method != "baseline" {
				hits, err := f.query(f.delegate, method, "amberrare", false)
				require.NoError(t, err)
				require.Empty(t, hits, "candidate statement must repeat current grant checks")
			}
		}
	}
	require.NoError(t, f.r.GormClient.Model(&models.SourceCredential{}).Where("id = ?", f.source).Update("deleted_at", time.Now().UTC()).Error)
	for _, method := range methods {
		_, err := f.query(f.owner, method, "amberrare", true)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		if method != "baseline" {
			hits, err := f.query(f.owner, method, "amberrare", false)
			require.NoError(t, err)
			require.Empty(t, hits)
		}
	}
	// A deleted native actor cannot retrieve through a still-indexed corpus.
	var actor models.User
	require.NoError(t, f.r.GormClient.Where("username = ?", "synthetic-retrieval-owner").First(&actor).Error)
	require.NoError(t, f.r.GormClient.Delete(&actor).Error)
	for _, method := range methods {
		_, err := f.query(f.owner, method, "amberrare", true)
		require.Error(t, err)
	}
	t.Log("lifecycle PASS: other owner/admin denied; current text/provenance rechecked; stale index deletion, expired/revoked grants, source deletion and native actor deletion denied")
}

// Record optimizer choices once during warmup; measured samples exclude EXPLAIN.
func (f *retrievalEvalFixture) explainCandidatePlan(t *testing.T, q *gorm.DB, method, term string) {
	t.Helper()
	dry := q.Session(&gorm.Session{DryRun: true}).Find(&[]models.SourceSearchHit{})
	require.NoError(t, dry.Error)
	pool, err := q.DB()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if f.r.GormClient.Dialector.Name() == "sqlite" {
		rows, err := pool.QueryContext(ctx, "EXPLAIN QUERY PLAN "+dry.Statement.SQL.String(), dry.Statement.Vars...)
		require.NoError(t, err)
		defer rows.Close()
		steps := []string{}
		for rows.Next() {
			var id, parent, unused int
			var detail string
			require.NoError(t, rows.Scan(&id, &parent, &unused, &detail))
			steps = append(steps, detail)
		}
		require.NoError(t, rows.Err())
		t.Logf("plan method=%s term=%s steps=%v", method, term, steps)
	} else {
		var raw []byte
		require.NoError(t, pool.QueryRowContext(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "+dry.Statement.SQL.String(), dry.Statement.Vars...).Scan(&raw))
		var result []struct {
			Plan map[string]interface{} `json:"Plan"`
		}
		require.NoError(t, json.Unmarshal(raw, &result))
		require.Len(t, result, 1)
		steps := []string{}
		var visit func(map[string]interface{})
		visit = func(node map[string]interface{}) {
			steps = append(steps, fmt.Sprintf("%v index=%v rows=%v loops=%v", node["Node Type"], node["Index Name"], node["Actual Rows"], node["Actual Loops"]))
			if children, ok := node["Plans"].([]interface{}); ok {
				for _, child := range children {
					visit(child.(map[string]interface{}))
				}
			}
		}
		visit(result[0].Plan)
		t.Logf("plan method=%s term=%s steps=%v", method, term, steps)
	}
}
