//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/permission"
	"github.com/titpetric/task-ui/storage"
)

// The search routes narrow to the sections and the folders a user sees, and
// the reindex is the admin's.
func TestSearchAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	users := storage.NewUserStorage(db)
	alice := &model.User{Name: "alice", Role: model.RoleAdmin, Admin: true}
	bob := &model.User{Name: "bob", Role: model.RoleViewer, Folders: model.EncodeList([]string{"acme"})}
	for _, user := range []*model.User{alice, bob} {
		assert.NoError(t, users.Create(ctx, user))
	}
	policies := storage.NewPermissionPolicyStorage(db)
	assert.NoError(t, policies.Put(ctx, permission.KindRole, model.RoleViewer, "prompt/view", permission.Allow))

	search := storage.NewSearchStorage(db)
	doc := &storage.SearchDocument{Kind: storage.SearchDocKind, Path: "src/acme/README.md", FolderID: "acme", Title: "Acme", Hash: "h1"}
	assert.NoError(t, search.ReplaceDocument(ctx, doc, []storage.SearchSection{{Ordinal: 0, Level: 1, Heading: "Acme", Breadcrumb: "Acme", Line: 1, Body: "Install the server."}, {Ordinal: 1, Level: 2, Heading: "Deploy", Breadcrumb: "Acme > Deploy", Line: 5, Body: "Run the daily deploy."}}, []storage.SearchTerm{{Ordinal: 1, Term: "daily", Weight: 4}, {Ordinal: 1, Term: "deploy", Weight: 1}}))
	other := &storage.SearchDocument{Kind: storage.SearchDocKind, Path: "src/other/README.md", FolderID: "other", Title: "Other", Hash: "h2"}
	assert.NoError(t, search.ReplaceDocument(ctx, other, []storage.SearchSection{{Ordinal: 0, Level: 1, Heading: "Other", Breadcrumb: "Other", Line: 1, Body: "Run the daily deploy."}}, []storage.SearchTerm{{Ordinal: 0, Term: "daily", Weight: 4}}))

	indexed := false
	router := api.New(api.Options{Users: users, Search: search, Index: func(context.Context) error { indexed = true; return nil }}).Routes()

	rec := ask(t, bob, router, http.MethodGet, "/search?q=daily", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var found struct {
		Hits []struct {
			Document string `json:"document"`
			Section  int64  `json:"ordinal"`
			Path     string `json:"path"`
			Snippet  string `json:"snippet"`
		} `json:"hits"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &found))
	assert.Len(t, found.Hits, 1, "the other folder is not bob's")
	assert.Equal(t, "src/acme/README.md", found.Hits[0].Path)
	assert.Contains(t, found.Hits[0].Snippet, "daily deploy")

	rec = ask(t, bob, router, http.MethodGet, "/search/documents/"+doc.ID, "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Deploy")

	rec = ask(t, bob, router, http.MethodGet, "/search/documents/"+doc.ID+"/sections/1", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Run the daily deploy.")

	rec = ask(t, bob, router, http.MethodGet, "/search/browse?path=src", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "src/acme")

	rec = ask(t, bob, router, http.MethodGet, "/search/documents/"+other.ID, "")
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = ask(t, bob, router, http.MethodPost, "/search/reindex", "")
	assert.Equal(t, http.StatusForbidden, rec.Code)
	rec = ask(t, alice, router, http.MethodPost, "/search/reindex", "")
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.True(t, indexed)
}
