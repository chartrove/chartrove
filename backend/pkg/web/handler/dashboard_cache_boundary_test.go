package handler

import (
	"encoding/json"
	"github.com/google/go-github/v93/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCustomDashboardGistBoundary(t *testing.T) {
	for _, tt := range []struct {
		name, id, content string
		wantErr           bool
	}{{"object", "0123456789abcdef", `{"title":"Synthetic Dashboard","widgets":[]}`, false}, {"null", "0123456789abcdef", `null`, true}, {"outside-id", "../outside", `{"title":"Synthetic Dashboard"}`, true}} {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			cache := filepath.Join(parent, "cache")
			require.NoError(t, os.Mkdir(cache, 0700))
			sentinel := filepath.Join(parent, "outside.json")
			prior := []byte(`{"synthetic":"keep-me"}`)
			require.NoError(t, os.WriteFile(sentinel, prior, 0600))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/gists/0123456789abcdef", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": tt.id, "files": map[string]interface{}{"dashboard.json": map[string]interface{}{"type": "application/json", "content": tt.content}}})
			}))
			defer server.Close()
			baseURL := server.URL + "/"
			client, err := github.NewClient(github.WithHTTPClient(server.Client()), github.WithURLs(&baseURL, nil))
			require.NoError(t, err)
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			var got string
			var gotErr error
			require.NotPanics(t, func() {
				got, gotErr = cacheCustomDashboard(logrus.NewEntry(logger), client, cache, "https://gist.github.com/synthetic/0123456789abcdef")
			})
			if tt.wantErr {
				require.Error(t, gotErr)
				require.Empty(t, got)
			} else {
				require.NoError(t, gotErr)
				require.Equal(t, filepath.Join(cache, "0123456789abcdef.json"), got)
				raw, err := os.ReadFile(got)
				require.NoError(t, err)
				var value map[string]interface{}
				require.NoError(t, json.Unmarshal(raw, &value))
				require.Equal(t, "Synthetic Dashboard", value["title"])
				require.Equal(t, "0123456789abcdef", value["id"])
				require.Equal(t, "https://gist.github.com/synthetic/0123456789abcdef", value["source"])
			}
			after, err := os.ReadFile(sentinel)
			require.NoError(t, err)
			require.Equal(t, prior, after, "response metadata must not overwrite files outside its dashboard cache")
		})
	}
}
