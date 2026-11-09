package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestApplicationPreservesRequestCancellationInRepositoryContext(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	app := &AppEngine{Config: cfg, Logger: logrus.NewEntry(logger), EventBus: event_bus.NewNoopEventBusServer(), StandbyMode: true}
	_, router := app.Setup()
	actor := uuid.New()
	deadline := time.Now().Add(time.Minute)
	parent, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	cancel()
	var retained context.Context
	router.GET("/synthetic-repository-context", func(c *gin.Context) {
		c.Set(pkg.ContextKeyTypeAuthUserID, actor)
		c.Set(pkg.ContextKeyTypeAuthUsername, "synthetic-actor")
		retained = middleware.AuthenticatedRequestContext(c)
		actual, ok := retained.Deadline()
		require.True(t, ok)
		require.Equal(t, deadline, actual)
		require.ErrorIs(t, retained.Err(), context.Canceled)
		select {
		case <-retained.Done():
		default:
			t.Error("request cancellation must reach repository work")
		}
		c.Status(http.StatusNoContent)
	})
	router.GET("/synthetic-next-context", func(c *gin.Context) {
		// An inherited account context must not replace absent authenticated
		// Gin values, nor change a context retained after the previous request.
		copy := middleware.AuthenticatedRequestContext(c)
		require.Equal(t, uuid.Nil, copy.Value(pkg.ContextKeyTypeAuthUserID))
		require.Equal(t, "", copy.Value(pkg.ContextKeyTypeAuthUsername))
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/synthetic-repository-context", nil).WithContext(parent)
	router.ServeHTTP(httptest.NewRecorder(), request)
	require.NotNil(t, retained)
	inherited := context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUserID, uuid.New())
	inherited = context.WithValue(inherited, pkg.ContextKeyTypeAuthUsername, "synthetic-untrusted-inheritance")
	for i := 0; i < 20; i++ {
		request = httptest.NewRequest(http.MethodGet, "/synthetic-next-context", nil).WithContext(inherited)
		router.ServeHTTP(httptest.NewRecorder(), request)
		require.Equal(t, actor, retained.Value(pkg.ContextKeyTypeAuthUserID))
		require.Equal(t, "synthetic-actor", retained.Value(pkg.ContextKeyTypeAuthUsername))
		require.ErrorIs(t, retained.Err(), context.Canceled)
	}
}
