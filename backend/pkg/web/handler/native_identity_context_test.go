package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHTTPDerivedSyncContextCannotSwitchNativeAccounts(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	cfg.Set("database.location", filepath.Join(t.TempDir(), "synthetic-sync-identity.db"))
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-sync-identity-key")
	log := logrus.New()
	log.SetOutput(io.Discard)
	repo, err := database.NewRepository(cfg, log, event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, repo.Close()) })
	user := &models.User{Username: "synthetic-sync-reused-name", Password: "synthetic-first-password"}
	require.NoError(t, repo.CreateUser(context.Background(), user))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(pkg.ContextKeyTypeAuthUsername, user.Username)
	c.Set(pkg.ContextKeyTypeAuthUserID, user.ID)
	ctx := GetBackgroundContext(c)
	stored, err := repo.GetCurrentUser(ctx)
	require.NoError(t, err)
	require.Equal(t, user.ID, stored.ID)
	require.NoError(t, repo.DeleteCurrentUser(ctx))
	replacement := &models.User{Username: user.Username, Password: "synthetic-second-password"}
	require.NoError(t, repo.CreateUser(context.Background(), replacement))
	require.NotEqual(t, user.ID, replacement.ID)
	_, err = repo.GetCurrentUser(ctx)
	require.Error(t, err)
	job := &models.BackgroundJob{JobType: pkg.BackgroundJobTypeSync, JobStatus: pkg.BackgroundJobStatusReady}
	require.Error(t, repo.CreateBackgroundJob(CreateBackgroundJobContext(ctx, "synthetic-job"), job))
	var jobs int64
	require.NoError(t, repo.(*database.GormRepository).GormClient.Model(&models.BackgroundJob{}).Where("user_id = ?", replacement.ID).Count(&jobs).Error)
	require.Zero(t, jobs)
	// A missing verified UUID in an HTTP-derived context also fails closed.
	c.Set(pkg.ContextKeyTypeAuthUsername, replacement.Username)
	delete(c.Keys, pkg.ContextKeyTypeAuthUserID)
	_, err = repo.GetCurrentUser(GetBackgroundContext(c))
	require.Error(t, err)
}

func TestHTTPDerivedJobContextSurvivesRequestReuseAndCancellation(t *testing.T) {
	request, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/synthetic-manual-job", nil).WithContext(request)
	owner := uuid.New()
	c.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-first-owner")
	c.Set(pkg.ContextKeyTypeAuthUserID, owner)
	c.Set(pkg.ContextKeyTypeAuthToken, "synthetic-private-local-token")
	job := CreateBackgroundJobContext(GetBackgroundContext(c), "synthetic-job")
	cancel()
	// Model request-pool reuse without a SQL timing assumption. Deferred SQL
	// cleanup and the tracked job must never inspect this mutable HTTP object.
	c.Request = httptest.NewRequest(http.MethodGet, "/synthetic-next-user", nil)
	c.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-second-owner")
	c.Set(pkg.ContextKeyTypeAuthUserID, uuid.New())
	require.Equal(t, owner, job.Value(pkg.ContextKeyTypeAuthUserID))
	require.Equal(t, "synthetic-first-owner", job.Value(pkg.ContextKeyTypeAuthUsername))
	require.Equal(t, "synthetic-job", job.Value(pkg.ContextKeyTypeBackgroundJobID))
	require.Nil(t, job.Value(pkg.ContextKeyTypeAuthToken), "detached jobs do not need the local JWT")
	require.NoError(t, job.Err())
	require.Nil(t, job.Done(), "retain the legacy tracked-job lifetime, not a new request-cancel contract")
	_, bounded := job.Deadline()
	require.False(t, bounded)
}
