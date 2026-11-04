package database

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Only disposable synthetic rows are used. A second owner's secret detects
// subqueries which bypass the ordinary outer user_id predicate.
func aggregationSecurityFixture(t *testing.T) (*GormRepository, context.Context) {
	t.Helper()
	client, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "aggregation.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	pool, err := client.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, client.AutoMigrate(&models.User{}))
	ownerID, otherID := uuid.New(), uuid.New()
	for _, user := range []models.User{
		{ModelBase: models.ModelBase{ID: ownerID}, Username: "synthetic-owner", Password: "owner-placeholder"},
		{ModelBase: models.ModelBase{ID: otherID}, Username: "synthetic-other", Password: "OTHER_OWNER_SECRET"},
	} {
		require.NoError(t, client.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)
	}
	require.NoError(t, client.Exec(`CREATE TABLE fhir_observation (
		id TEXT, user_id TEXT, sort_date TEXT, source_resource_type TEXT,
		code TEXT, valueQuantity TEXT, valueString TEXT)`).Error)
	for i, row := range []struct {
		owner uuid.UUID
		value int
	}{{ownerID, 0}, {ownerID, 4}, {otherID, 999}} {
		require.NoError(t, client.Exec(`INSERT INTO fhir_observation VALUES (?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("row-%d", i), row.owner.String(), fmt.Sprintf("2026-01-%02d", i+1), "Observation",
			`[{"system":"synthetic","code":"test"}]`, fmt.Sprintf(`[{"value":%d,"nested":{"value":%d}}]`, row.value, row.value),
			`[{"value":"synthetic"}]`).Error)
	}
	return &GormRepository{GormClient: client}, context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, "synthetic-owner")
}

func TestAggregationQueryIsolatesSQLSyntax(t *testing.T) {
	repo, ctx := aggregationSecurityFixture(t)
	function := "(SELECT password FROM users WHERE username='synthetic-other')||"
	property := "code')||(SELECT/**/password/**/FROM/**/users/**/WHERE/**/username='synthetic-other')||('"
	for _, test := range []struct {
		name         string
		aggregations models.QueryResourceAggregations
	}{
		{"group-function", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "id", Function: function}}},
		{"order-function-value", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "id"}, OrderBy: &models.QueryResourceAggregation{Field: "id", Function: function}}},
		{"group-property", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "code:" + property}}},
		{"order-property-value", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "id"}, OrderBy: &models.QueryResourceAggregation{Field: "code:" + property}}},
		{"count-property", models.QueryResourceAggregations{CountBy: &models.QueryResourceAggregation{Field: "code:" + property}}},
		{"star-function", models.QueryResourceAggregations{OrderBy: &models.QueryResourceAggregation{Field: "*", Function: "count(*)+(SELECT length(password) FROM users WHERE username='synthetic-other')--"}}},
		{"extension-function", models.QueryResourceAggregations{OrderBy: &models.QueryResourceAggregation{Field: "id", Function: "load_extension"}}},
		{"unknown-field", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "unknown"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query, err := repo.sqlQueryResources(ctx, models.QueryResource{From: "Observation", Aggregations: &test.aggregations})
			if err == nil && test.name != "extension-function" {
				var rows []map[string]interface{}
				if test.name == "group-property" || test.name == "order-property-value" || test.name == "count-property" {
					statement := query.Session(&gorm.Session{DryRun: true}).Find(&[]map[string]interface{}{}).Statement
					require.NotContains(t, statement.SQL.String(), property)
					require.Contains(t, statement.Vars, "$."+property)
				}
				result := query.Find(&rows)
				require.NoError(t, result.Error)
				for _, row := range rows {
					for key, value := range row {
						row[key] = aggregationScalar(value)
					}
				}
				t.Logf("synthetic SQL syntax control: execution=%v rows=%v", result.Error, rows)
				require.NotContains(t, fmt.Sprint(rows), "OTHER_OWNER_SECRET", "aggregation cannot read another owner's account")
			}
			switch test.name {
			case "group-property", "order-property-value", "count-property":
				require.NoError(t, err, "JSON property paths must remain bound data")
			default:
				require.Error(t, err, "SQL functions and identifiers must be validated before execution")
			}
		})
	}
}

func TestAggregationQueryPreservesOwnerResults(t *testing.T) {
	repo, ctx := aggregationSecurityFixture(t)
	for _, test := range []struct {
		name           string
		aggregations   models.QueryResourceAggregations
		labels, values []interface{}
	}{
		{"count-all", models.QueryResourceAggregations{CountBy: &models.QueryResourceAggregation{Field: "*"}}, []interface{}{"Observation"}, []interface{}{int64(2)}},
		{"count-token", models.QueryResourceAggregations{CountBy: &models.QueryResourceAggregation{Field: "code"}}, []interface{}{"synthetic|test"}, []interface{}{int64(2)}},
		{"quantity-max", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "code:code"}, OrderBy: &models.QueryResourceAggregation{Field: "valueQuantity:value", Function: "MAX"}}, []interface{}{"test"}, []interface{}{int64(4)}},
		{"quantity-sum", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "code:code"}, OrderBy: &models.QueryResourceAggregation{Field: "valueQuantity:value", Function: "sum"}}, []interface{}{"test"}, []interface{}{int64(4)}},
		{"nested-property-zero", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "valueQuantity:nested.value"}}, []interface{}{int64(0), int64(4)}, []interface{}{int64(1), int64(1)}},
		{"date-group", models.QueryResourceAggregations{GroupBy: &models.QueryResourceAggregation{Field: "sort_date", Function: "date"}}, []interface{}{"2026-01-01", "2026-01-02"}, []interface{}{int64(1), int64(1)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input, err := json.Marshal(test.aggregations)
			require.NoError(t, err)
			query, err := repo.sqlQueryResources(ctx, models.QueryResource{From: "Observation", Aggregations: &test.aggregations})
			require.NoError(t, err)
			after, err := json.Marshal(test.aggregations)
			require.NoError(t, err)
			require.JSONEq(t, string(input), string(after), "compiling must preserve the saved query")
			var rows []map[string]interface{}
			require.NoError(t, query.Find(&rows).Error)
			var labels, values []interface{}
			for _, row := range rows {
				labels = append(labels, aggregationScalar(row["label"]))
				values = append(values, aggregationScalar(row["value"]))
			}
			require.ElementsMatch(t, test.labels, labels)
			require.ElementsMatch(t, test.values, values)
		})
	}
}

func aggregationScalar(value interface{}) interface{} {
	v := reflect.ValueOf(value)
	for v.IsValid() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

func TestAggregationQueryPreservesOwnerOrdering(t *testing.T) {
	repo, ctx := aggregationSecurityFixture(t)
	for _, test := range []struct {
		name, field string
		ids         []string
	}{
		{"default", "", []string{"row-1", "row-0"}},
		{"primitive-date", "sort_date", []string{"row-1", "row-0"}},
		{"primitive-keyword", "id", []string{"row-0", "row-1"}},
		{"quantity-zero", "valueQuantity:value", []string{"row-0", "row-1"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := models.QueryResource{From: "Observation"}
			if test.field != "" {
				input.Aggregations = &models.QueryResourceAggregations{OrderBy: &models.QueryResourceAggregation{Field: test.field}}
			}
			query, err := repo.sqlQueryResources(ctx, input)
			require.NoError(t, err)
			var rows []map[string]interface{}
			require.NoError(t, query.Find(&rows).Error)
			var ids []string
			for _, row := range rows {
				ids = append(ids, row["id"].(string))
			}
			require.Equal(t, test.ids, ids)
		})
	}
}
