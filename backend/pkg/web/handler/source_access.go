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
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sourceAccessOperation func(*gin.Context, database.SourceAccessRepository, uuid.UUID) (interface{}, error)

// RegisterSourceAccessRoutes must receive the existing authenticated secure
// group. These routes never substitute an owner identity into the auth context.
func RegisterSourceAccessRoutes(secure gin.IRoutes) {
	const prefix = "/source-access"
	secure.GET(prefix+"/:sourceId/metadata/:resourceType", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		c.Header("Cache-Control", "no-store")
		metadata, ok := repo.(database.SourceMetadataRepository)
		if !ok {
			return nil, errors.New("source metadata unavailable")
		}
		page := 0
		if value, ok := c.GetQuery("page"); ok {
			var err error
			page, err = strconv.Atoi(value)
			if err != nil {
				return nil, database.ErrSourceAccessInput
			}
		}
		return metadata.ListSourceMetadata(middleware.AuthenticatedRequestContext(c), source, c.Param("resourceType"), page)
	}))
	secure.POST(prefix+"/:sourceId/search/:resourceType", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		c.Header("Cache-Control", "no-store")
		search, ok := repo.(database.SourceSearchRepository)
		if !ok {
			return nil, errors.New("source search unavailable")
		}
		var body struct {
			Text string          `json:"text"`
			Page int             `json:"page"`
			Mode json.RawMessage `json:"mode"`
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
		if err := decodeSourceAccessBody(c, &body); err != nil {
			return nil, err
		}
		mode := "literal"
		if body.Mode != nil {
			var selected *string
			if err := json.Unmarshal(body.Mode, &selected); err != nil || selected == nil {
				return nil, database.ErrSourceAccessInput
			}
			mode = *selected
		}
		switch mode {
		case "literal":
			return search.SearchSourceResources(middleware.AuthenticatedRequestContext(c), source, c.Param("resourceType"), body.Text, body.Page)
		case "exact_id":
			modes, ok := repo.(database.SourceExactIDSearchRepository)
			if !ok {
				return nil, errors.New("source search mode unavailable")
			}
			return modes.SearchSourceResourceID(middleware.AuthenticatedRequestContext(c), source, c.Param("resourceType"), body.Text, body.Page)
		default:
			return nil, database.ErrSourceAccessInput
		}
	}))
	secure.GET(prefix+"/shared", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, _ uuid.UUID) (interface{}, error) {
		return repo.ListSharedSources(middleware.AuthenticatedRequestContext(c))
	}))
	secure.GET(prefix+"/:sourceId/grants", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.ListSourceGrants(middleware.AuthenticatedRequestContext(c), source)
	}))
	secure.POST(prefix+"/:sourceId/grants", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		var payload struct {
			DelegateUsername string                   `json:"delegate_username"`
			AccessLevel      models.SourceAccessLevel `json:"access_level"`
			ExpiresAt        *time.Time               `json:"expires_at"`
		}
		if err := decodeSourceAccessBody(c, &payload); err != nil {
			return nil, err
		}
		return repo.SetSourceGrant(middleware.AuthenticatedRequestContext(c), source, payload.DelegateUsername, payload.AccessLevel, payload.ExpiresAt)
	}))
	secure.DELETE(prefix+"/:sourceId/grants/:grantId", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		id, err := uuid.Parse(c.Param("grantId"))
		if err != nil {
			return nil, database.ErrSourceAccessInput
		}
		return nil, repo.RevokeSourceGrant(middleware.AuthenticatedRequestContext(c), source, id)
	}))
	secure.GET(prefix+"/:sourceId/audit", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.ListSourceAccessAudit(middleware.AuthenticatedRequestContext(c), source)
	}))
	secure.GET(prefix+"/:sourceId/summary", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.GetSharedSourceSummary(middleware.AuthenticatedRequestContext(c), source)
	}))
	secure.GET(prefix+"/:sourceId/resources/:resourceType", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		page := 0
		if value, ok := c.GetQuery("page"); ok {
			var err error
			page, err = strconv.Atoi(value)
			if err != nil {
				return nil, database.ErrSourceAccessInput
			}
		}
		return repo.ListSharedResources(middleware.AuthenticatedRequestContext(c), source, c.Param("resourceType"), page)
	}))
	secure.GET(prefix+"/:sourceId/resources/:resourceType/:resourceId", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		return repo.GetSharedResource(middleware.AuthenticatedRequestContext(c), source, c.Param("resourceType"), c.Param("resourceId"))
	}))
	secure.POST(prefix+"/:sourceId/resources", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		var payload struct {
			ResourceRaw json.RawMessage `json:"resource_raw"`
		}
		if err := decodeSourceAccessBody(c, &payload); err != nil {
			return nil, err
		}
		return repo.CreateLocalResource(middleware.AuthenticatedRequestContext(c), source, payload.ResourceRaw)
	}))
	secure.PATCH(prefix+"/:sourceId/resources/:resourceType/:resourceId", sourceAccessEndpoint(func(c *gin.Context, repo database.SourceAccessRepository, source uuid.UUID) (interface{}, error) {
		var payload struct {
			ResourceRaw json.RawMessage `json:"resource_raw"`
		}
		if err := decodeSourceAccessBody(c, &payload); err != nil {
			return nil, err
		}
		return nil, repo.UpdateSharedResource(middleware.AuthenticatedRequestContext(c), source, c.Param("resourceType"), c.Param("resourceId"), payload.ResourceRaw)
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
