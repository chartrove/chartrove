package handler

import (
	"io"
	"net/http"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterDocumentImportRoutes(secure gin.IRoutes) {
	// Bound in-flight original buffers. OCR itself never runs on the server.
	slots := make(chan struct{}, 2)
	secure.GET("/document-import/sources", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		repo, ok := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DocumentImportRepository)
		if !ok {
			c.Status(503)
			return
		}
		data, err := repo.ListDocumentImportSources(middleware.AuthenticatedRequestContext(c))
		sourceAccessResponse(c, data, err)
	})
	secure.POST("/document-import/:sourceId", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		repo, ok := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DocumentImportRepository)
		if !ok {
			c.Status(503)
			return
		}
		source, err := uuid.Parse(c.Param("sourceId"))
		if err != nil {
			sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		if err = repo.AuthorizeDocumentImportSource(middleware.AuthenticatedRequestContext(c), source); err != nil {
			sourceAccessResponse(c, nil, err)
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			c.JSON(503, gin.H{"success": false, "error": "document import busy"})
			return
		}
		controller := http.NewResponseController(c.Writer)
		_ = controller.SetReadDeadline(time.Now().Add(30 * time.Second))
		defer func() { _ = controller.SetReadDeadline(time.Time{}) }()
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, models.DocumentMaxBytes+(256<<10))
		reader, err := c.Request.MultipartReader()
		if err != nil {
			sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		var input models.ReviewedDocument
		seen := map[string]bool{}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
				return
			}
			name := part.FormName()
			if seen[name] || len(seen) >= 4 {
				sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
				return
			}
			seen[name] = true
			limit := int64(0)
			switch name {
			case "file":
				limit = models.DocumentMaxBytes
				input.Filename = part.FileName()
			case "text":
				limit = models.DocumentMaxTranscriptBytes
			case "title":
				limit = 1024
			case "reviewed":
				limit = 5
			default:
				sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
				return
			}
			data, err := io.ReadAll(io.LimitReader(part, limit+1))
			if err != nil || int64(len(data)) > limit {
				sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
				return
			}
			switch name {
			case "file":
				input.Original = data
			case "text":
				input.Transcript = string(data)
			case "title":
				input.Title = string(data)
			case "reviewed":
				input.Reviewed = string(data) == "true"
			}
			if err = part.Close(); err != nil {
				sourceAccessResponse(c, nil, database.ErrSourceAccessInput)
				return
			}
		}
		data, err := repo.ImportReviewedDocument(middleware.AuthenticatedRequestContext(c), source, input)
		sourceAccessResponse(c, data, err)
	})
}
