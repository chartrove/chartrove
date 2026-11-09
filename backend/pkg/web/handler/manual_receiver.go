package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// manualImportReceiver checks optional caller intent against the already verified,
// immutable actor. The form assertion is not authentication or delegated authority.
// Missing assertions preserve legacy clients; present malformed assertions fail closed.
func manualImportReceiver(c *gin.Context, actor uuid.UUID) (int, string) {
	invalid := func() (int, string) {
		return http.StatusBadRequest, "invalid receiving profile confirmation"
	}
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		return invalid()
	}
	for key := range query {
		if strings.EqualFold(key, "expected_owner_id") {
			return invalid() // An assertion belongs in exactly one multipart value, not the URL.
		}
	}
	form := c.Request.MultipartForm // storeFileLocally has already parsed this form.
	if form == nil {
		return invalid()
	}
	for key := range form.File {
		if strings.EqualFold(key, "expected_owner_id") {
			return invalid()
		}
	}
	for key := range form.Value {
		if key != "expected_owner_id" && strings.EqualFold(key, "expected_owner_id") {
			return invalid()
		}
	}
	values, present := form.Value["expected_owner_id"]
	if !present {
		return 0, ""
	}
	if len(values) != 1 || len(values[0]) != 36 {
		return invalid()
	}
	expected, err := uuid.Parse(values[0])
	if err != nil || expected == uuid.Nil || !strings.EqualFold(values[0], expected.String()) {
		return invalid()
	}
	if actor == uuid.Nil || expected != actor {
		return http.StatusConflict, "receiving profile no longer matches this import"
	}
	return 0, ""
}
