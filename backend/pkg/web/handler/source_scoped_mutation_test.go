package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sm "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOwnerMutationsRequireUnambiguousSource(t *testing.T) {
	for _, method := range []string{http.MethodPatch, http.MethodDelete, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			resourceType := "Observation"
			if method == http.MethodPut {
				resourceType = "Practitioner"
			}
			cfg, err := config.Create()
			require.NoError(t, err)
			cfg.Set("database.type", "sqlite")
			cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-source-scoped.db"))
			cfg.Set("database.encryption.enabled", true)
			cfg.Set("database.encryption.key", "synthetic-source-scoped-key")
			db, err := database.NewRepository(cfg, logrus.New(), event_bus.NewNoopEventBusServer())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			owner := &models.User{Username: "synthetic-target-owner", Password: "synthetic-password"}
			other := &models.User{Username: "synthetic-target-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}
			for _, u := range []*models.User{owner, other} {
				require.NoError(t, db.CreateUser(context.Background(), u))
			}
			ctx := func(u *models.User) context.Context {
				c := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username)
				return context.WithValue(c, pkg.ContextKeyTypeAuthUserID, u.ID)
			}
			sources := []*models.SourceCredential{{Patient: "synthetic-source-A"}, {Patient: "synthetic-source-B"}, {Patient: "synthetic-other-source"}}
			id := "synthetic-colliding-observation"
			for i, s := range sources {
				actor := owner
				if i == 2 {
					actor = other
				}
				require.NoError(t, db.CreateSource(ctx(actor), s))
				clinical := fmt.Sprintf(`{"resourceType":"Observation","id":%q,"status":"final","code":{"text":"Synthetic original source %d"}}`, id, i)
				if method == http.MethodPut {
					clinical = fmt.Sprintf(`{"resourceType":"Practitioner","id":%q,"name":[{"text":"Synthetic original source %d"}]}`, id, i)
				}
				_, err := db.UpsertRawResource(ctx(actor), s, sm.RawResourceFhir{SourceResourceType: resourceType, SourceResourceID: id, ResourceRaw: json.RawMessage(clinical)})
				require.NoError(t, err)
			}
			access := db.(database.SourceAccessRepository)
			_, err = access.SetSourceGrant(ctx(other), sources[2].ID, owner.Username, models.SourceAccessEdit, nil)
			require.NoError(t, err)
			actor := owner
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(pkg.ContextKeyTypeDatabase, db)
				c.Set(pkg.ContextKeyTypeLogger, logrus.NewEntry(logrus.New()))
				c.Set(pkg.ContextKeyTypeAuthUsername, actor.Username)
				c.Set(pkg.ContextKeyTypeAuthUserID, actor.ID)
			})
			router.PATCH("/resource/fhir/:resourceType/:resourceId", UpdateResourceFhir)
			router.DELETE("/resource/fhir/:resourceType/:resourceId", DeleteResourceFhir)
			router.PUT("/practitioners/:practitionerId", UpdatePractitioner)
			payload := fmt.Sprintf(`{"resource_raw":{"resourceType":"Observation","id":%q,"status":"amended","code":{"text":"Synthetic explicitly selected edit"}}}`, id)
			if method == http.MethodPut {
				payload = fmt.Sprintf(`{"resource":{"resourceType":"Practitioner","id":%q,"name":[{"text":"Synthetic explicitly selected edit"}]}}`, id)
			}
			request := func(resourceID, sourceJSON, query string, status int) {
				t.Helper()
				body := strings.TrimSuffix(payload, "}")
				if sourceJSON != "" {
					body += `,"source_id":` + sourceJSON
				}
				body += "}"
				w := httptest.NewRecorder()
				requestPath := "/resource/fhir/" + resourceType + "/" + resourceID + query
				if method == http.MethodPut {
					requestPath = "/practitioners/" + resourceID + query
				}
				req := httptest.NewRequest(method, requestPath, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, req)
				require.Equal(t, status, w.Code, w.Body.String())
			}
			request(id, "", "", http.StatusConflict)
			for _, bad := range []string{"", "not-a-uuid", uuid.Nil.String()} {
				if method != http.MethodDelete {
					request(id, fmt.Sprintf("%q", bad), "", 400)
				} else {
					request(id, "", "?source_id="+bad, 400)
				}
			}
			if method != http.MethodDelete {
				for _, bad := range []string{"null", "42", "[]"} {
					request(id, bad, "", 400)
				}
				request(id, fmt.Sprintf(`%q} {"source_id":%q`, sources[0].ID, sources[1].ID), "", 400)
			} else {
				request(id, "", "?source_id=%zz", 400)
				request(id, "", "?source_id="+sources[0].ID.String()+"&source_id="+sources[1].ID.String(), 400)
			}
			choose := func(source uuid.UUID, status int) {
				t.Helper()
				if method != http.MethodDelete {
					request(id, fmt.Sprintf("%q", source), "", status)
				} else {
					request(id, "", "?source_id="+source.String(), status)
				}
			}
			choose(uuid.New(), 404)
			choose(sources[2].ID, 404) // Even an active grant cannot authorize owner mutations.
			actor = other
			choose(sources[0].ID, 404) // Administration grants no clinical ownership.
			actor = owner
			choose(sources[1].ID, 200)
			selected, err := access.GetSharedResource(ctx(owner), sources[1].ID, resourceType, id)
			if method != http.MethodDelete {
				require.NoError(t, err)
				require.Contains(t, string(selected.ResourceRaw), "Synthetic explicitly selected edit")
			} else {
				require.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
			for i, s := range []*models.SourceCredential{sources[0], sources[2]} {
				u := owner
				original := 0
				if i == 1 {
					u, original = other, 2
				}
				r, err := access.GetSharedResource(ctx(u), s.ID, resourceType, id)
				require.NoError(t, err)
				require.Contains(t, string(r.ResourceRaw), fmt.Sprintf("Synthetic original source %d", original))
			}
		})
	}
}
