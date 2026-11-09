package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils/clinicalsnapshot"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// The secure group supplies real authentication. Expected receiving identity and
// explicit confirmation prevent a stale preview being imported after account
// switching; neither is provider authorization or a native-origin attestation.
func RegisterClinicalSnapshotRoutes(secure gin.IRoutes) {
	imports := make(chan struct{}, 1)
	downloads := make(chan struct{}, 2)
	secure.POST("/source/clinical-snapshot", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		db := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)
		repo, ok := db.(database.ClinicalSnapshotRepository)
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "clinical snapshot import unavailable"})
			return
		}
		requestContext := middleware.AuthenticatedRequestContext(c)
		actor, err := db.GetCurrentUser(requestContext)
		if err != nil || actor == nil || actor.ID == uuid.Nil || actor.DeletedAt != nil {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessDenied)
			return
		}
		select {
		case imports <- struct{}{}:
			defer func() { <-imports }()
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "clinical snapshot import busy; retry later"})
			return
		}
		mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || mediaType != "application/json" {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		// Raw JSON travels as a string, not a JS-number-parsed object. Quoting
		// can double its wire size; decoded collections still have the 16 MiB cap.
		controller := http.NewResponseController(c.Writer)
		_ = controller.SetReadDeadline(time.Now().Add(30 * time.Second))
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*clinicalsnapshot.MaxBytes+(8<<10))
		wire, readErr := io.ReadAll(c.Request.Body)
		_ = controller.SetReadDeadline(time.Time{})
		if readErr != nil || !utf8.Valid(wire) {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		fields, err := clinicalRequestObject(wire, "expected_owner_id", "confirmed", "origin", "snapshot_json")
		var expectedOwner, raw string
		var confirmed bool
		if err != nil || json.Unmarshal(fields["expected_owner_id"], &expectedOwner) != nil || json.Unmarshal(fields["confirmed"], &confirmed) != nil || !confirmed || json.Unmarshal(fields["snapshot_json"], &raw) != nil || len(raw) == 0 || len(raw) > clinicalsnapshot.MaxBytes {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		owner, err := uuid.Parse(expectedOwner)
		if err != nil || owner != actor.ID {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessDenied)
			return
		}
		originFields, err := clinicalRequestObject(fields["origin"], "fhir_base_url", "display")
		var base, display string
		if err != nil || json.Unmarshal(originFields["fhir_base_url"], &base) != nil || json.Unmarshal(originFields["display"], &display) != nil {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		result, err := repo.ImportClinicalSnapshot(requestContext, base, display, []byte(raw))
		if err != nil {
			clinicalSnapshotResponse(c, nil, err)
			return
		}
		if !result.AlreadyImported {
			// Persistence already committed. Event failure is not import failure,
			// and must never cause a duplicate write or expose a clinical payload.
			if value, exists := c.Get(pkg.ContextKeyTypeEventBusServer); exists {
				bus, supported := value.(event_bus.BestEffortPublisher)
				if !supported || bus.TryPublishMessage(models.NewEventSourceComplete(actor.ID.String(), result.SourceID.String())) != nil {
					c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry).Warn("clinical snapshot committed; refresh notification unavailable")
				}
			}
		}
		clinicalSnapshotResponse(c, result, nil)
	})
	secure.GET("/source/:sourceId/clinical-snapshots", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		repo, source, ok := clinicalSnapshotSource(c)
		if !ok {
			return
		}
		limit, offset := 50, 0
		var err error
		if value, exists := c.GetQuery("limit"); exists {
			limit, err = strconv.Atoi(value)
		}
		if value, exists := c.GetQuery("offset"); exists && err == nil {
			offset, err = strconv.Atoi(value)
		}
		if err != nil {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		ctx, cancel := context.WithTimeout(middleware.AuthenticatedRequestContext(c), 30*time.Second)
		defer cancel()
		rows, err := repo.ListClinicalSnapshots(ctx, source, limit, offset)
		clinicalSnapshotResponse(c, rows, err)
	})
	secure.GET("/source/:sourceId/clinical-snapshots/:snapshotId/original", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		repo, source, ok := clinicalSnapshotSource(c)
		if !ok {
			return
		}
		snapshot, err := uuid.Parse(c.Param("snapshotId"))
		if err != nil {
			clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
			return
		}
		select {
		case downloads <- struct{}{}:
			defer func() { <-downloads }()
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "clinical original download busy; retry later"})
			return
		}
		ctx, cancel := context.WithTimeout(middleware.AuthenticatedRequestContext(c), 30*time.Second)
		defer cancel()
		raw, err := repo.GetClinicalSnapshotOriginal(ctx, source, snapshot)
		if err != nil {
			clinicalSnapshotResponse(c, nil, err)
			return
		}
		// A repository context does not bound a blocked socket write. Do not let
		// connected non-reading clients retain both shared download slots forever.
		transport := http.NewResponseController(c.Writer)
		if err := transport.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil && !errors.Is(err, http.ErrNotSupported) {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		defer func() { _ = transport.SetWriteDeadline(time.Time{}) }()
		// In-process recording writers have no transport deadline to configure.
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Disposition", `attachment; filename="clinical-snapshot-`+snapshot.String()+`.json"`)
		c.Data(http.StatusOK, "application/fhir+json", raw)
	})
}

func clinicalSnapshotSource(c *gin.Context) (database.ClinicalSnapshotRepository, uuid.UUID, bool) {
	repo, ok := c.MustGet(pkg.ContextKeyTypeDatabase).(database.ClinicalSnapshotRepository)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "clinical snapshot access unavailable"})
		return nil, uuid.Nil, false
	}
	source, err := uuid.Parse(c.Param("sourceId"))
	if err != nil {
		clinicalSnapshotResponse(c, nil, database.ErrSourceAccessInput)
		return nil, uuid.Nil, false
	}
	return repo, source, true
}

func clinicalSnapshotResponse(c *gin.Context, data interface{}, err error) {
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
		return
	}
	status, message := http.StatusInternalServerError, "clinical snapshot operation failed"
	switch {
	case errors.Is(err, database.ErrSourceAccessDenied):
		status, message = http.StatusForbidden, "clinical snapshot access denied"
	case errors.Is(err, database.ErrSourceAccessInput):
		status, message = http.StatusBadRequest, "invalid clinical snapshot request"
	case errors.Is(err, database.ErrClinicalSnapshotConflict):
		status, message = http.StatusConflict, "clinical snapshot identity requires review"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		status, message = http.StatusRequestTimeout, "clinical snapshot operation did not complete"
	}
	// Errors may wrap database/decoder text. Never return or log those values.
	c.JSON(status, gin.H{"success": false, "error": message})
}

func clinicalRequestObject(raw []byte, keys ...string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return nil, database.ErrSourceAccessInput
	}
	allowed := map[string]bool{}
	for _, key := range keys {
		allowed[key] = true
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || !allowed[key] || fields[key] != nil {
			return nil, database.ErrSourceAccessInput
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return nil, database.ErrSourceAccessInput
		}
		fields[key] = value
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') || len(fields) != len(keys) || decoder.Decode(new(interface{})) != io.EOF {
		return nil, database.ErrSourceAccessInput
	}
	return fields, nil
}
