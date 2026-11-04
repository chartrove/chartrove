package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestReviewedDocumentImportOwnershipProvenanceAndAtomicity(t *testing.T) {
	cfg, err := config.Create()
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "synthetic-documents.db")
	cfg.Set("database.type", "sqlite")
	cfg.Set("database.location", path)
	cfg.Set("database.encryption.enabled", true)
	cfg.Set("database.encryption.key", "synthetic-document-key")
	db, err := newTestRepository(t, cfg, logrus.New(), event_bus.NewNoopEventBusServer())
	require.NoError(t, err)
	r := db.(*GormRepository)
	users := []*models.User{{Username: "synthetic-document-owner", Password: "synthetic-password"}, {Username: "synthetic-document-delegate", Password: "synthetic-password"}, {Username: "synthetic-document-admin", Password: "synthetic-password", Role: pkg.UserRoleAdmin}}
	contexts := []context.Context{}
	for _, u := range users {
		require.NoError(t, r.CreateUser(context.Background(), u))
		contexts = append(contexts, context.WithValue(context.Background(), pkg.ContextKeyTypeAuthUsername, u.Username))
	}
	sources, err := r.GetSources(contexts[0])
	require.NoError(t, err)
	source := sources[0].ID
	input := models.ReviewedDocument{Filename: "synthetic-scan.pdf", Title: "Synthetic reviewed document", Original: []byte("%PDF-1.7\n% Synthetic original document bytes\n%%EOF"), Transcript: "Synthetic private reviewed transcript café 100% follow_up! quoted \"text\" and slash \\ العربية", Reviewed: true}
	for _, ctx := range contexts[1:] {
		_, err = r.ImportReviewedDocument(ctx, source, input)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		_, err = r.SearchSourceResources(ctx, source, "DocumentReference", "reviewed transcript", 0)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
	}
	foreign := &models.SourceCredential{PlatformType: sp.PlatformType("synthetic-external"), Display: "Synthetic provider", Patient: "synthetic-external-patient"}
	require.NoError(t, r.CreateSource(contexts[0], foreign))
	require.ErrorIs(t, r.AuthorizeDocumentImportSource(contexts[0], foreign.ID), ErrSourceAccessDenied)
	available, err := r.ListDocumentImportSources(contexts[0])
	require.NoError(t, err)
	require.Len(t, available, 1)
	require.Equal(t, source, available[0].ID)
	var trace bytes.Buffer
	r.GormClient.Logger = logger.New(log.New(&trace, "", 0), logger.Config{LogLevel: logger.Info, Colorful: false})
	first, err := r.ImportReviewedDocument(contexts[0], source, input)
	require.NoError(t, err)
	require.True(t, first.Created)
	second, err := r.ImportReviewedDocument(contexts[0], source, input)
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Equal(t, first.ResourceID, second.ResourceID)
	require.Contains(t, trace.String(), "INSERT INTO")
	for _, private := range []string{input.Transcript, input.Filename, base64.StdEncoding.EncodeToString(input.Original)} {
		require.NotContains(t, trace.String(), private)
	}
	stored, err := r.GetSharedResource(contexts[0], source, "DocumentReference", first.ResourceID)
	require.NoError(t, err)
	require.False(t, stored.Editable)
	var resource struct {
		ID   string `json:"id"`
		Meta struct {
			Source string `json:"source"`
		} `json:"meta"`
		Content []struct {
			Attachment struct {
				ContentType string `json:"contentType"`
				Data        string `json:"data"`
			} `json:"attachment"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal(stored.ResourceRaw, &resource))
	require.Equal(t, first.ResourceID, resource.ID)
	require.Equal(t, documentImportOrigin, resource.Meta.Source)
	require.Len(t, resource.Content, 2)
	original, err := base64.StdEncoding.DecodeString(resource.Content[0].Attachment.Data)
	require.NoError(t, err)
	require.Equal(t, input.Original, original)
	text, err := base64.StdEncoding.DecodeString(resource.Content[1].Attachment.Data)
	require.NoError(t, err)
	require.Equal(t, input.Transcript, string(text))
	digest := sha256.Sum256(input.Original)
	require.Equal(t, hex.EncodeToString(digest[:]), first.OriginalSHA256)
	grant, err := r.SetSourceGrant(contexts[0], source, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	shared, err := r.GetSharedResource(contexts[1], source, "DocumentReference", first.ResourceID)
	require.NoError(t, err)
	require.False(t, shared.Editable)
	require.ErrorIs(t, r.UpdateSharedResource(contexts[1], source, "DocumentReference", first.ResourceID, json.RawMessage(shared.ResourceRaw)), ErrSourceAccessDenied)
	_, err = r.ImportReviewedDocument(contexts[1], source, input)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	meta, err := r.ListSourceMetadata(contexts[1], source, "DocumentReference", 0)
	require.NoError(t, err)
	require.Len(t, meta, 1)
	metadata, err := json.Marshal(meta)
	require.NoError(t, err)
	require.NotContains(t, string(metadata), "resource_raw")
	require.NotContains(t, string(metadata), input.Transcript)
	for _, term := range []string{"reviewed transcript", "café", "100%", "follow_up!", `"text"`, `\`, "العربية"} {
		hits, err := r.SearchSourceResources(contexts[1], source, "DocumentReference", term, 0)
		require.NoError(t, err)
		require.Len(t, hits, 1, term)
		require.Equal(t, first.ResourceID, hits[0].SourceResourceID)
	}
	hits, err := r.SearchSourceResources(contexts[1], source, "DocumentReference", "100units", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
	require.NotContains(t, trace.String(), "café")
	audit, err := r.ListSourceAccessAudit(contexts[0], source)
	require.NoError(t, err)
	imports := 0
	for _, event := range audit {
		if event.Action == "import_document" {
			imports++
		}
	}
	require.Equal(t, 1, imports)
	for _, change := range []func(*models.ReviewedDocument){func(d *models.ReviewedDocument) { d.Reviewed = false }, func(d *models.ReviewedDocument) { d.Original = []byte("unsupported synthetic text") }, func(d *models.ReviewedDocument) { d.Title = "" }, func(d *models.ReviewedDocument) {
		d.Transcript = strings.Repeat("x", models.DocumentMaxTranscriptBytes+1)
	}} {
		bad := input
		change(&bad)
		_, err = r.ImportReviewedDocument(contexts[0], source, bad)
		require.ErrorIs(t, err, ErrSourceAccessInput)
	}
	require.NoError(t, r.GormClient.Callback().Create().Before("gorm:create").Register("synthetic_document_audit_failure", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "source_access_audits" {
			_ = tx.AddError(errors.New("synthetic audit failure"))
		}
	}))
	changed := input
	changed.Title = "Synthetic rollback document"
	_, err = r.ImportReviewedDocument(contexts[0], source, changed)
	require.Error(t, err)
	require.NoError(t, r.GormClient.Callback().Create().Remove("synthetic_document_audit_failure"))
	meta, err = r.ListSourceMetadata(contexts[0], source, "DocumentReference", 0)
	require.NoError(t, err)
	require.Len(t, meta, 1)
	ciphertext, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, string(ciphertext), input.Transcript)
	require.NotContains(t, string(ciphertext), string(input.Original))
	// Earlier imports have no transcript projection. Keep title navigation
	// working without silently rewriting/backfilling their clinical payload.
	require.NoError(t, r.GormClient.Table("fhir_document_reference").Where("source_id = ? AND user_id = ? AND source_resource_id = ?", source, users[0].ID, first.ResourceID).Update("text", nil).Error)
	hits, err = r.SearchSourceResources(contexts[1], source, "DocumentReference", "reviewed transcript", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
	hits, err = r.SearchSourceResources(contexts[1], source, "DocumentReference", input.Title, 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.NoError(t, r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", grant.ID).Update("expires_at", time.Now().Add(-time.Minute)).Error)
	_, err = r.SearchSourceResources(contexts[1], source, "DocumentReference", "reviewed transcript", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
	require.NoError(t, r.RevokeSourceGrant(contexts[0], source, grant.ID))
	_, err = r.SearchSourceResources(contexts[1], source, "DocumentReference", "reviewed transcript", 0)
	require.ErrorIs(t, err, ErrSourceAccessDenied)
}
