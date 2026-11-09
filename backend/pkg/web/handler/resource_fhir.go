package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/utils"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func QueryResourceFhir(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	var query models.QueryResource
	if err := c.ShouldBindJSON(&query); err != nil {
		logger.Errorln("An error occurred while parsing queries", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	queryResults, err := databaseRepo.QueryResources(middleware.AuthenticatedRequestContext(c), query)
	if err != nil {
		logger.Errorln("An error occurred while querying resources", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	//sort by date
	//queryResults = utils.SortResourceListByDate(queryResults)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": queryResults})
}

func ListResourceFhir(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	listResourceQueryOptions := models.ListResourceQueryOptions{}
	if len(c.Query("sourceResourceType")) > 0 {
		listResourceQueryOptions.SourceResourceType = c.Query("sourceResourceType")
	}
	if len(c.Query("sourceID")) > 0 {
		listResourceQueryOptions.SourceID = c.Query("sourceID")
	}
	if len(c.Query("sourceResourceID")) > 0 {
		listResourceQueryOptions.SourceResourceID = c.Query("sourceResourceID")
	}
	if len(c.Query("page")) > 0 {
		listResourceQueryOptions.Limit = pkg.ResourceListPageSize //hardcoded number of resources per page
		pageNumb, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			logger.Errorln("An error occurred while calculating page number", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		listResourceQueryOptions.Offset = pageNumb * listResourceQueryOptions.Limit
	}
	wrappedResourceModels, err := databaseRepo.ListResources(middleware.AuthenticatedRequestContext(c), listResourceQueryOptions)

	if c.Query("sortBy") == "title" {
		wrappedResourceModels = utils.SortResourceListByTitle(wrappedResourceModels)
	} else {
		wrappedResourceModels = utils.SortResourceListByDate(wrappedResourceModels)
	}

	if err != nil {
		logger.Errorln("An error occurred while retrieving resources", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": wrappedResourceModels})
}

// this endpoint retrieves a specific resource by its ID
func GetResourceFhir(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	resourceId := strings.Trim(c.Param("resourceId"), "/")
	sourceId := strings.Trim(c.Param("sourceId"), "/")
	wrappedResourceModel, err := databaseRepo.GetResourceBySourceId(middleware.AuthenticatedRequestContext(c), sourceId, resourceId)

	if err != nil {
		logger.Errorln("An error occurred while retrieving resource", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": wrappedResourceModel})
}

// deprecated - using Manual Resource Wizard instead
func CreateResourceComposition(c *gin.Context) {

	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	type jsonPayload struct {
		Resources []*models.ResourceBase `json:"resources"`
		Title     string                 `json:"title"`
	}
	var payload jsonPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		logger.Errorln("An error occurred while parsing posted resources & title", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	err := databaseRepo.AddResourceComposition(middleware.AuthenticatedRequestContext(c), payload.Title, payload.Resources)
	if err != nil {
		logger.Errorln("An error occurred while creating resource group (composition)", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetResourceFhirGraph
// Retrieve a list of all fhir resources (vertex), and a list of all associations (edge)
// Generate a graph
// find the PredecessorMap
// - filter to only vertices that are "Condition" or "Encounter" and are "root" nodes (have no edges directed to this node)
func GetResourceFhirGraph(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	graphType := strings.Trim(c.Param("graphType"), "/")

	var graphOptions models.ResourceGraphOptions
	if err := c.ShouldBindJSON(&graphOptions); err != nil {
		logger.Errorln("An error occurred while parsing resource graph query options", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	if len(graphOptions.ResourcesIds) == 0 {
		logger.Errorln("No resource ids specified")
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	resourceListDictionary, err := databaseRepo.GetFlattenedResourceGraph(middleware.AuthenticatedRequestContext(c), pkg.ResourceGraphType(graphType), graphOptions)
	if err != nil {
		logger.Errorln("An error occurred while retrieving list of resources", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": map[string]interface{}{
		"results": resourceListDictionary,
	}})
}

func UpdateResourceFhir(c *gin.Context) {
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	resourceType := strings.Trim(c.Param("resourceType"), "/")
	resourceId := strings.Trim(c.Param("resourceId"), "/")

	type UpdatePayload struct {
		ResourceRaw json.RawMessage `json:"resource_raw"`
		SortTitle   json.RawMessage `json:"sort_title"`
		SortDate    json.RawMessage `json:"sort_date"`
		SourceID    json.RawMessage `json:"source_id"`
	}
	var payload UpdatePayload
	err := decodeOwnedMutationJSON(c, &payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request payload"})
		return
	}
	var sortTitle, sortDate *string
	for _, field := range []struct {
		raw    json.RawMessage
		target **string
	}{{payload.SortTitle, &sortTitle}, {payload.SortDate, &sortDate}} {
		if field.raw == nil {
			continue
		}
		var text string
		if json.Unmarshal(field.raw, &text) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request payload"})
			return
		}
		*field.target = &text // Explicit null/empty retains legacy clearing behavior.
	}

	source, err := parseOwnedJSONSource(payload.SourceID)
	if err != nil {
		writeOwnedMutationError(c, err)
		return
	}
	updates, ok := databaseRepo.(database.OwnedResourceMutationRepository)
	if !ok {
		writeOwnedMutationError(c, errors.New("resource editing unavailable"))
		return
	}
	resource, err := updates.ResolveOwnedResource(middleware.AuthenticatedRequestContext(c), resourceType, resourceId, source)
	if err != nil {
		writeOwnedMutationError(c, err)
		return
	}
	updateError := updates.UpdateOwnedResource(middleware.AuthenticatedRequestContext(c), resource.SourceID, resourceType, resourceId, database.OwnedResourceUpdate{
		ResourceRaw: payload.ResourceRaw, SortTitle: sortTitle,
		SortDate: parseDateTimeWithFallback(sortDate), SetSortDate: sortDate != nil,
	})
	if updateError != nil {
		writeOwnedMutationError(c, updateError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteResourceFhir(c *gin.Context) {
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	resourceType := strings.Trim(c.Param("resourceType"), "/")
	resourceId := strings.Trim(c.Param("resourceId"), "/")

	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		writeOwnedMutationError(c, database.ErrSourceAccessInput)
		return
	}
	var source *uuid.UUID
	if values, present := query["source_id"]; present {
		if len(values) != 1 {
			writeOwnedMutationError(c, database.ErrSourceAccessInput)
			return
		}
		var err error
		source, err = parseOwnedMutationSource(values[0])
		if err != nil {
			writeOwnedMutationError(c, err)
			return
		}
	}
	mutations, ok := databaseRepo.(database.OwnedResourceMutationRepository)
	if !ok {
		writeOwnedMutationError(c, errors.New("resource deletion unavailable"))
		return
	}
	resource, err := mutations.ResolveOwnedResource(middleware.AuthenticatedRequestContext(c), resourceType, resourceId, source)
	if err != nil {
		writeOwnedMutationError(c, err)
		return
	}
	if err := mutations.DeleteOwnedResource(middleware.AuthenticatedRequestContext(c), resource.SourceID, resourceType, resourceId); err != nil {
		writeOwnedMutationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "resource deleted successfully"})
}

func parseOwnedMutationSource(text string) (*uuid.UUID, error) {
	id, err := uuid.Parse(text)
	if err != nil || id == uuid.Nil {
		return nil, database.ErrSourceAccessInput
	}
	return &id, nil
}

func parseOwnedJSONSource(raw json.RawMessage) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return nil, database.ErrSourceAccessInput
	}
	return parseOwnedMutationSource(text)
}

func decodeOwnedMutationJSON(c *gin.Context, payload interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	if decoder.Decode(payload) != nil {
		return database.ErrSourceAccessInput
	}
	var extra interface{}
	if decoder.Decode(&extra) != io.EOF {
		return database.ErrSourceAccessInput
	}
	return nil
}

func writeOwnedMutationError(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "Resource operation failed"
	switch {
	case errors.Is(err, database.ErrOwnedResourceAmbiguous):
		status, message = http.StatusConflict, "Resource ID is ambiguous; select its source"
	case errors.Is(err, database.ErrSourceAccessDenied):
		status, message = http.StatusNotFound, "Resource unavailable"
	case errors.Is(err, database.ErrSourceAccessInput):
		status, message = http.StatusBadRequest, "Invalid resource request"
	}
	c.JSON(status, gin.H{"success": false, "error": message})
}

func parseDateTimeWithFallback(dateTime *string) *time.Time {
	if dateTime == nil {
		return nil
	}
	var parsedDateTime time.Time
	var err error
	parsedDateTime, err = time.Parse(time.RFC3339, *dateTime)
	if err != nil {
		parsedDateTime, err = time.Parse("2006-01-02T15:04", *dateTime)
		if err != nil {
			parsedDateTime, err = time.Parse("2006-01-02", *dateTime)
			if err != nil {
				return nil
			}
		}
	}
	return &parsedDateTime
}
