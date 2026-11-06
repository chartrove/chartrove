package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	sp "github.com/fastenhealth/fasten-sources/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestEncryptedSQLiteBackupIndependentRecovery(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()
	source := filepath.Join(sourceDir, "synthetic.db")
	key := "synthetic-independent-key ?&#='"
	open := func(path, secret string, readOnly bool) (*GormRepository, error) {
		cfg, err := config.Create()
		require.NoError(t, err)
		cfg.Set("database.type", "sqlite")
		cfg.Set("database.location", path)
		cfg.Set("database.encryption.enabled", true)
		cfg.Set("database.encryption.key", secret)
		cfg.Set("database.validation_mode", readOnly)
		log := logrus.New()
		log.SetOutput(io.Discard)
		r, err := NewRepository(cfg, log, event_bus.NewNoopEventBusServer())
		if err != nil {
			return nil, err
		}
		return r.(*GormRepository), nil
	}
	r, err := open(source, key, false)
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	require.NoError(t, r.GormClient.Exec("PRAGMA wal_autocheckpoint=0").Error)
	require.NoError(t, r.GormClient.Exec("PRAGMA wal_checkpoint(TRUNCATE)").Error)
	users := []*models.User{
		{Username: "synthetic-backup-owner", Password: "independent-synthetic-password"},
		{Username: "synthetic-backup-delegate", Password: "delegate-synthetic-password"},
		{Username: "synthetic-backup-recovery", Password: "recovery-synthetic-password", Role: pkg.UserRoleAdmin},
	}
	actor := func(u *models.User) context.Context {
		c := context.WithValue(ctx, pkg.ContextKeyTypeAuthUsername, u.Username)
		return context.WithValue(c, pkg.ContextKeyTypeAuthUserID, u.ID)
	}
	for _, u := range users {
		require.NoError(t, r.CreateUser(ctx, u))
	}
	owner, delegate, admin := actor(users[0]), actor(users[1]), actor(users[2])
	sources, err := r.GetSources(owner)
	require.NoError(t, err)
	sourceID := sources[0].ID
	local, err := r.CreateLocalResource(owner, sourceID, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic original"},"effectiveDateTime":"2020-01-02T03:04:05+05:30","valueQuantity":{"value":0.1234567890123456789}}`))
	require.NoError(t, err)
	_, err = r.SetSourceGrant(owner, sourceID, users[1].Username, models.SourceAccessEdit, nil)
	require.NoError(t, err)
	deleted, err := r.CreateLocalResource(owner, sourceID, json.RawMessage(`{"resourceType":"Observation","status":"final","code":{"text":"Synthetic deleted"}}`))
	require.NoError(t, err)
	require.NoError(t, r.DeleteOwnedResource(owner, sourceID, "Observation", deleted.SourceResourceID))
	deniedSources := []uuid.UUID{}
	for _, state := range []string{"revoked", "expired", "deleted"} {
		s := &models.SourceCredential{Patient: "synthetic-backup-" + state, Display: "Synthetic " + state, PlatformType: sp.PlatformTypeFasten}
		require.NoError(t, r.CreateSource(owner, s))
		grant, err := r.SetSourceGrant(owner, s.ID, users[1].Username, models.SourceAccessView, nil)
		require.NoError(t, err)
		switch state {
		case "revoked":
			require.NoError(t, r.RevokeSourceGrant(owner, s.ID, grant.ID))
		case "expired":
			require.NoError(t, r.GormClient.Model(&models.SourceGrant{}).Where("id = ?", grant.ID).Update("expires_at", time.Now().UTC().Add(-time.Minute)).Error)
		case "deleted":
			_, err = r.DeleteSource(owner, s.ID.String())
			require.NoError(t, err)
		}
		deniedSources = append(deniedSources, s.ID)
	}
	_, session, err := r.CompleteOIDC(ctx, "https://synthetic-issuer.example.invalid/", "synthetic-subject", users[1].ID, users[1].Password, time.Now().UTC(), time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, r.RevokeOIDCSession(ctx, users[1].ID, session.ID))
	original := append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("Synthetic attachment bytes\n"), 12000)...)
	doc, err := r.ImportReviewedDocument(owner, sourceID, models.ReviewedDocument{Filename: "synthetic.pdf", Title: "Synthetic document", Original: original, Transcript: "Synthetic reviewed transcript", Reviewed: true})
	require.NoError(t, err)
	pool, err := r.GormClient.DB()
	require.NoError(t, err)
	want := backupTableSnapshot(t, pool)
	wal, err := os.Stat(source + "-wal")
	require.NoError(t, err)
	require.Positive(t, wal.Size(), "fixture must contain committed WAL state")
	output := filepath.Join(privateBackupParent(t), "private backup?#")
	receipt, err := BackupEncryptedSQLite(ctx, source, output, key)
	require.NoError(t, err)
	backup := filepath.Join(output, "database.sqlite")
	data, err := os.ReadFile(backup)
	require.NoError(t, err)
	require.False(t, bytes.HasPrefix(data, []byte("SQLite format 3")))
	require.NotContains(t, string(data), "Synthetic original")
	digest := sha256.Sum256(data)
	require.Equal(t, hex.EncodeToString(digest[:]), receipt.SHA256)
	for _, name := range []string{"database.sqlite", "receipt.json"} {
		info, err := os.Stat(filepath.Join(output, name))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
	entries, err := os.ReadDir(output)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	info, err := os.Stat(output)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0700), info.Mode().Perm())
	safe, err := os.ReadFile(filepath.Join(output, "receipt.json"))
	require.NoError(t, err)
	require.NotContains(t, string(safe), key)
	require.NotContains(t, string(safe), source)
	require.Equal(t, want, backupTableSnapshot(t, pool), "backup must not migrate or change source rows")
	// Make the original installation unavailable. Recovery gets only the encrypted
	// artifact and an independently supplied key, never source WAL/configuration.
	require.NoError(t, r.Close())
	require.NoError(t, os.RemoveAll(sourceDir))
	// Use a plain destination name for the existing constructor's path contract.
	restoredPath := filepath.Join(t.TempDir(), "restored.db")
	require.NoError(t, os.WriteFile(restoredPath, data, 0600))
	for _, bad := range []string{"wrong-synthetic-key", ""} {
		_, err := open(restoredPath, bad, true)
		require.Error(t, err)
	}
	for restart := 0; restart < 2; restart++ {
		restored, err := open(restoredPath, key, false)
		require.NoError(t, err)
		p, err := restored.GormClient.DB()
		require.NoError(t, err)
		require.Equal(t, want, backupTableSnapshot(t, p), "all application tables must survive restart")
		for _, u := range users {
			found, err := restored.GetUserByUsername(ctx, u.Username)
			require.NoError(t, err)
			require.Equal(t, u.ID, found.ID)
			require.Equal(t, u.Password, found.Password)
		}
		u, err := restored.GetUserByUsername(ctx, users[2].Username)
		require.NoError(t, err)
		require.NoError(t, u.CheckPassword("recovery-synthetic-password"))
		require.Error(t, u.CheckPassword("wrong-password"))
		row, err := restored.GetSharedResource(delegate, sourceID, "Observation", local.SourceResourceID)
		require.NoError(t, err)
		require.True(t, row.Editable)
		require.Equal(t, local.ID, row.ID)
		require.Contains(t, string(row.ResourceRaw), "0.1234567890123456789")
		require.Contains(t, string(row.ResourceRaw), "2020-01-02T03:04:05+05:30")
		_, err = restored.GetSharedResource(admin, sourceID, "Observation", local.SourceResourceID)
		require.ErrorIs(t, err, ErrSourceAccessDenied)
		_, err = restored.GetSharedResource(owner, sourceID, "Observation", deleted.SourceResourceID)
		require.Error(t, err)
		for _, id := range deniedSources {
			_, err = restored.GetSharedSourceSummary(delegate, id)
			require.ErrorIs(t, err, ErrSourceAccessDenied)
		}
		require.ErrorIs(t, restored.ValidateOIDCSession(ctx, session.ID, users[1].ID, "https://synthetic-issuer.example.invalid/"), ErrOIDCDenied)
		attachment, err := restored.GetSharedResource(owner, sourceID, "DocumentReference", doc.ResourceID)
		require.NoError(t, err)
		var parsed struct {
			Content []struct{ Attachment struct{ Data string } }
		}
		require.NoError(t, json.Unmarshal(attachment.ResourceRaw, &parsed))
		decoded, err := base64.StdEncoding.DecodeString(parsed.Content[0].Attachment.Data)
		require.NoError(t, err)
		require.Equal(t, original, decoded)
		require.NoError(t, restored.Close())
	}
}

// Compare every application table, including audit/identity/provenance/index
// tables, without decoding clinical JSON numbers through float64.
func backupTableSnapshot(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	require.NoError(t, err)
	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	result := map[string]string{}
	for _, table := range tables {
		r, err := db.Query(`SELECT * FROM "` + strings.ReplaceAll(table, `"`, `""`) + `"`)
		require.NoError(t, err)
		cols, err := r.Columns()
		require.NoError(t, err)
		var records []string
		for r.Next() {
			values := make([]interface{}, len(cols))
			dest := make([]interface{}, len(cols))
			for i := range values {
				dest[i] = &values[i]
			}
			require.NoError(t, r.Scan(dest...))
			raw, err := json.Marshal(values)
			require.NoError(t, err)
			records = append(records, string(raw))
		}
		require.NoError(t, r.Err())
		require.NoError(t, r.Close())
		sort.Strings(records)
		result[table] = strings.Join(records, "\n")
	}
	return result
}
