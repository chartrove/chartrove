package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/web/middleware"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	sourcePkg "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreatePractitioner(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	type CreatePractitionerRequest struct {
		Resource map[string]interface{} `json:"resource"`
	}

	var request CreatePractitionerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request payload"})
		return
	}

	if resourceType, ok := request.Resource["resourceType"].(string); !ok || resourceType != "Practitioner" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "resource must be of type Practitioner"})
		return
	}

	currentUser, err := databaseRepo.GetCurrentUser(middleware.AuthenticatedRequestContext(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "could not get current user"})
		return
	}

	practitionerSource := &models.SourceCredential{
		UserID:       currentUser.ID,
		PlatformType: sourcePkg.PlatformTypeManual,
		Patient:      "",
	}

	err = databaseRepo.CreateSource(middleware.AuthenticatedRequestContext(c), practitionerSource)
	if err != nil {
		logger.Errorln("Error creating practitioner source:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to create source"})
		return
	}

	resourceRaw, err := json.Marshal(request.Resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to serialize resource"})
		return
	}

	resourceId, ok := request.Resource["id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "practitioner must have an id field"})
		return
	}

	sortTitle := extractPractitionerSortTitle(request.Resource)

	resourceToUpsert := sourceModels.RawResourceFhir{
		SourceResourceType: "Practitioner",
		SourceResourceID:   resourceId,
		ResourceRaw:        resourceRaw,
		SortTitle:          &sortTitle,
	}

	createdResource, createError := databaseRepo.UpsertRawResource(middleware.AuthenticatedRequestContext(c), practitionerSource, resourceToUpsert)
	if createError != nil {
		logger.Errorln("Error upserting practitioner:", createError)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to create practitioner"})
		return
	}

	logger.Infof("Successfully created practitioner: %s with sort_title: %s", resourceId, sortTitle)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "practitioner created successfully", "id": resourceId, "resource": createdResource})
}

func GetPractitionerEncounterHistory(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	practitionerId := c.Param("practitionerId")
	if practitionerId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "practitioner ID is required"})
		return
	}

	relatedResources, err := databaseRepo.FindPractitionerEncounters(middleware.AuthenticatedRequestContext(c), practitionerId)
	if err != nil {
		logger.Errorln("Error fetching encounters:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to fetch encounters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"relatedResources": relatedResources,
	})
}

func UpdatePractitioner(c *gin.Context) {
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)
	practitionerId := c.Param("practitionerId")
	if practitionerId == "" {
		writeOwnedMutationError(c, database.ErrSourceAccessInput)
		return
	}
	var request struct {
		Resource json.RawMessage `json:"resource"`
		SourceID json.RawMessage `json:"source_id"`
	}
	if decodeOwnedMutationJSON(c, &request) != nil {
		writeOwnedMutationError(c, database.ErrSourceAccessInput)
		return
	}
	var resource map[string]interface{}
	if json.Unmarshal(request.Resource, &resource) != nil {
		writeOwnedMutationError(c, database.ErrSourceAccessInput)
		return
	}
	if resource["resourceType"] != "Practitioner" || resource["id"] != practitionerId {
		writeOwnedMutationError(c, database.ErrSourceAccessInput)
		return
	}
	source, err := parseOwnedJSONSource(request.SourceID)
	if err != nil {
		writeOwnedMutationError(c, err)
		return
	}
	mutations, ok := databaseRepo.(database.OwnedResourceMutationRepository)
	if !ok {
		writeOwnedMutationError(c, database.ErrSourceAccessDenied)
		return
	}
	existing, err := mutations.ResolveOwnedResource(middleware.AuthenticatedRequestContext(c), "Practitioner", practitionerId, source)
	if err != nil {
		writeOwnedMutationError(c, err)
		return
	}
	sortTitle := extractPractitionerSortTitle(resource)
	// Keep exact clinical bytes; decode only the fields needed for presentation.
	err = mutations.UpdateOwnedResource(middleware.AuthenticatedRequestContext(c), existing.SourceID, "Practitioner", practitionerId, database.OwnedResourceUpdate{ResourceRaw: request.Resource, SortTitle: &sortTitle})
	if err != nil {
		writeOwnedMutationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "practitioner updated successfully", "resource": true})
}

func extractPractitionerSortTitle(resource map[string]interface{}) string {

	if names, ok := resource["name"].([]interface{}); ok && len(names) > 0 {
		if nameObj, ok := names[0].(map[string]interface{}); ok {

			if text, ok := nameObj["text"].(string); ok && text != "" {
				return text
			}

			var family, given string
			if familyName, ok := nameObj["family"].(string); ok {
				family = familyName
			}
			if givenNames, ok := nameObj["given"].([]interface{}); ok && len(givenNames) > 0 {
				if givenName, ok := givenNames[0].(string); ok {
					given = givenName
				}
			}

			if family != "" && given != "" {
				return family + ", " + given
			} else if family != "" {
				return family
			}
		}
	}

	if name, ok := resource["name"].(string); ok && name != "" {
		return name
	}
	if fullName, ok := resource["full_name"].(string); ok && fullName != "" {
		return fullName
	}

	return ""
}

func AddPractitionerToFavorites(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	type AddFavoriteRequest struct {
		SourceID     string `json:"source_id" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		ResourceID   string `json:"resource_id" binding:"required"`
	}

	var request AddFavoriteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Errorln("Error binding request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request payload"})
		return
	}

	// Validate resource type
	if request.ResourceType != "Practitioner" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "only Practitioner resources are supported"})
		return
	}

	// Get current user
	currentUser, err := databaseRepo.GetCurrentUser(middleware.AuthenticatedRequestContext(c))
	if err != nil {
		logger.Errorln("Error getting current user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "could not get current user"})
		return
	}

	// Check if already favorited
	exists, err := databaseRepo.CheckFavoriteExists(
		middleware.AuthenticatedRequestContext(c),
		currentUser.ID.String(),
		request.SourceID,
		request.ResourceType,
		request.ResourceID,
	)
	if err != nil {
		logger.Errorln("Error checking if favorite exists:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to check favorite status"})
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "practitioner already in favorites"})
		return
	}

	// Add to favorites
	err = databaseRepo.AddFavorite(
		middleware.AuthenticatedRequestContext(c),
		currentUser.ID.String(),
		request.SourceID,
		request.ResourceType,
		request.ResourceID,
	)
	if err != nil {
		logger.Errorln("Error adding favorite:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to add to favorites"})
		return
	}

	logger.Infof("Successfully added practitioner %s (source %s) to favorites for user %s", request.ResourceID, request.SourceID, currentUser.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "practitioner added to favorites"})
}

// NEW: Remove practitioner from favorites
func RemovePractitionerFromFavorites(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	type RemoveFavoriteRequest struct {
		SourceID     string `json:"source_id" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		ResourceID   string `json:"resource_id" binding:"required"`
	}

	var request RemoveFavoriteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Errorln("Error binding request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request payload"})
		return
	}

	// Validate resource type
	if request.ResourceType != "Practitioner" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "only Practitioner resources are supported"})
		return
	}

	// Get current user
	currentUser, err := databaseRepo.GetCurrentUser(middleware.AuthenticatedRequestContext(c))
	if err != nil {
		logger.Errorln("Error getting current user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "could not get current user"})
		return
	}

	// Remove from favorites
	err = databaseRepo.RemoveFavorite(
		middleware.AuthenticatedRequestContext(c),
		currentUser.ID.String(),
		request.SourceID,
		request.ResourceType,
		request.ResourceID,
	)
	if err != nil {
		logger.Errorln("Error removing favorite:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to remove from favorites"})
		return
	}

	logger.Infof("Successfully removed practitioner %s (source %s) from favorites for user %s", request.ResourceID, request.SourceID, currentUser.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "practitioner removed from favorites"})
}

// NEW: Get user's favorite practitioners
func GetUserFavoritePractitioners(c *gin.Context) {
	logger := c.MustGet(pkg.ContextKeyTypeLogger).(*logrus.Entry)
	databaseRepo := c.MustGet(pkg.ContextKeyTypeDatabase).(database.DatabaseRepository)

	// Get query parameters
	resourceType := c.Query("resource_type")
	if resourceType == "" {
		resourceType = "Practitioner" // Default to Practitioner
	}

	// Validate resource type
	if resourceType != "Practitioner" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "only Practitioner resources are supported"})
		return
	}

	// Get current user
	currentUser, err := databaseRepo.GetCurrentUser(middleware.AuthenticatedRequestContext(c))
	if err != nil {
		logger.Errorln("Error getting current user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "could not get current user"})
		return
	}

	// Get user's favorites
	favorites, err := databaseRepo.GetUserFavorites(middleware.AuthenticatedRequestContext(c), currentUser.ID.String(), resourceType)
	if err != nil {
		logger.Errorln("Error getting user favorites:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get favorites"})
		return
	}

	// Transform to the expected format (array of objects with resource_id)
	favoriteObjects := make([]map[string]interface{}, len(favorites))
	for i, fav := range favorites {
		favoriteObjects[i] = map[string]interface{}{
			"source_id":     fav.SourceID,
			"resource_id":   fav.ResourceID,
			"resource_type": fav.ResourceType,
		}
	}

	logger.Infof("Retrieved %d favorite practitioners for user %s", len(favorites), currentUser.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": favoriteObjects})
}
