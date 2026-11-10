package handler

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestManualImportReceiverConfirmation(t *testing.T) {
	owner := uuid.MustParse("abcdef01-2345-4678-9123-abcdef012345")
	for _, test := range []struct {
		name   string
		actor  uuid.UUID
		form   *multipart.Form
		query  string
		status int
	}{
		{name: "legacy assertion absent", form: &multipart.Form{}},
		{name: "verified owner", actor: owner, form: &multipart.Form{Value: map[string][]string{"expected_owner_id": {owner.String()}}}},
		{name: "upper case UUID is the same identity", actor: owner, form: &multipart.Form{Value: map[string][]string{"expected_owner_id": {strings.ToUpper(owner.String())}}}},
		{name: "missing immutable actor", form: &multipart.Form{Value: map[string][]string{"expected_owner_id": {owner.String()}}}, status: http.StatusConflict},
		{name: "nil parsed form", actor: owner, status: http.StatusBadRequest},
		{name: "query alias", actor: owner, form: &multipart.Form{}, query: "?EXPECTED_OWNER_ID=" + owner.String(), status: http.StatusBadRequest},
		{name: "invalid query encoding", actor: owner, form: &multipart.Form{}, query: "?expected_owner_id=%ZZ", status: http.StatusBadRequest},
		{name: "value plus alias", actor: owner, form: &multipart.Form{Value: map[string][]string{"expected_owner_id": {owner.String()}, "EXPECTED_OWNER_ID": {owner.String()}}}, status: http.StatusBadRequest},
		{name: "file alias", actor: owner, form: &multipart.Form{File: map[string][]*multipart.FileHeader{"Expected_Owner_ID": {{Filename: "synthetic-private-file"}}}}, status: http.StatusBadRequest},
		{name: "noncanonical wrapper", actor: owner, form: &multipart.Form{Value: map[string][]string{"expected_owner_id": {"urn:uuid:" + owner.String()}}}, status: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/api/secure/source/manual"+test.query, nil)
			c.Request.MultipartForm = test.form
			status, message := manualImportReceiver(c, test.actor)
			require.Equal(t, test.status, status)
			if status == 0 {
				require.Empty(t, message)
			} else if status == http.StatusConflict {
				require.Equal(t, "receiving profile no longer matches this import", message)
			} else {
				require.Equal(t, "invalid receiving profile confirmation", message)
			}
		})
	}
}
