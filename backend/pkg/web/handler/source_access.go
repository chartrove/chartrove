package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sourceAccessOperation func(*gin.Context, database.SourceAccessRepository, uuid.UUID) (interface{}, error)

// RegisterSourceAccessRoutes must receive the existing authenticated secure
// group. These routes never substitute an owner identity into the auth context.
func RegisterSourceAccessRoutes(secure *gin.RouterGroup) {
	routes := secure.Group("/source-access")
	routes.GET("/shared", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, _ uuid.UUID) (interface{}, error) {
		return repo.ListSharedSources(c)
	}))
	routes.GET("/:sourceId/grants", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.ListSourceGrants(c, source)
	}))
	routes.POST("/:sourceId/grants", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		var payload struct {
			DelegateUsername string                   `json:"delegate_username"`
			AccessLevel      models.SourceAccessLevel `json:"access_level"`
			ExpiresAt        *time.Time               `json:"expires_at"`
		}
		if err := decodeSourceAccessBody(c, &payload); err != nil {
			return nil, err
		}
		return repo.SetSourceGrant(c, source, payload.DelegateUsername, payload.AccessLevel, payload.ExpiresAt)
	}))
	routes.DELETE("/:sourceId/grants/:grantId", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		id, err := uuid.Parse(c.Param("grantId"))
		if err != nil {
			return nil, database.ErrSourceAccessInput
		}
		return nil, repo.RevokeSourceGrant(c, source, id)
	}))
	routes.GET("/:sourceId/audit", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.ListSourceAccessAudit(c, source)
	}))
	routes.GET("/:sourceId/summary", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.GetSharedSourceSummary(c, source)
	}))
	routes.GET("/:sourceId/resources/:resourceType", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		page := 0
		if value, ok := c.GetQuery("page"); ok {
			var err error
			page, err = strconv.Atoi(value)
			if err != nil {
				return nil, database.ErrSourceAccessInput
			}
		}
		return repo.ListSharedResources(c, source, c.Param("resourceType"), page)
	}))
	routes.GET("/:sourceId/resources/:resourceType/:resourceId", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.GetSharedResource(c, source, c.Param("resourceType"), c.Param("resourceId"))
	}))
	routes.POST("/:sourceId/resources", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		var payload struct {
			ResourceRaw json.RawMessage `json:"resource_raw"`
		}
		if err := decodeSourceAccessBody(c, &payload); err != nil {
			return nil, err
		}
		return repo.CreateLocalResource(c, source, payload.ResourceRaw)
	}))
	routes.PATCH("/:sourceId/resources/:resourceType/:resourceId", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		var payload struct {
			ResourceRaw json.RawMessage `json:"resource_raw"`
		}
		if err := decodeSourceAccessBody(c, &payload); err != nil {
			return nil, err
		}
		return nil, repo.UpdateSharedResource(c, source, c.Param("resourceType"), c.Param("resourceId"), payload.ResourceRaw)
	}))
}

func decodeSourceAccessBody(c *gin.Context, payload interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(payload); err != nil {
		return database.ErrSourceAccessInput
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return database.ErrSourceAccessInput
	}
	return nil
}

func sourceAccessEndpoint(operation sourceAccessOperation) gin.HandlerFunc {
	return func(c *gin.Context) {
		repo, ok := c.MustGet(pkg.ContextKeyTypeDatabase).(database.SourceAccessRepository)
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "source access unavailable"})
			return
		}
		var source uuid.UUID
		if value := c.Param("sourceId"); value != "" {
			var err error
			source, err = uuid.Parse(value)
			if err != nil {
				sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
				return
			}
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
		data, err := operation(c, repo, source)
		sourceAccessResponse(c, data, err)
	}
}

func sourceAccessResponse(c *gin.Context, data interface{}, err error) {
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
		return
	}
	status, message := http.StatusInternalServerError, "source access operation failed"
	switch {
	case errors.Is(err, database.ErrSourceAccessDenied), errors.Is(err, gorm.ErrRecordNotFound):
		// Do not reveal whether a foreign source/resource exists. A permission
		// denial also does not require signing out an otherwise valid session.
		status, message = http.StatusNotFound, "source or resource unavailable"
	case errors.Is(err, database.ErrSourceAccessInput):
		status, message = http.StatusBadRequest, database.ErrSourceAccessInput.Error()
	}
	c.JSON(status, gin.H{"success": false, "error": message})
}
