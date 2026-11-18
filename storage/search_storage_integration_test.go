//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The index replaces a document whole, finds it by a prefix, filters by folder,
// and answers its outline, one section and its job targets.
func TestSearchStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	search := storage.NewSearchStorage(newDB(t))

	doc := &storage.SearchDocument{Kind: storage.SearchDocKind, Path: "src/acme/README.md", FolderID: "acme", Workspace: "internal", Project: "acme", Title: "Acme", Hash: "h1", Size: 12}
	sections := []storage.SearchSection{
		{Ordinal: 0, Level: 1, Heading: "Acme", Breadcrumb: "Acme", Line: 1, Body: "Install the server."},
		{Ordinal: 1, Level: 2, Heading: "Deploy", Breadcrumb: "Acme > Deploy", Line: 5, Body: "Run the daily deploy."},
	}
	terms := []storage.SearchTerm{
		{Ordinal: 0, Term: "acme", Weight: 4},
		{Ordinal: 0, Term: "install", Weight: 1},
		{Ordinal: 1, Term: "daily", Weight: 3},
		{Ordinal: 1, Term: "deploy", Weight: 1},
	}
	assert.NoError(t, search.ReplaceDocument(ctx, doc, sections, terms))
	assert.NotEmpty(t, doc.ID)

	docs, err := search.Documents(ctx, "acme")
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "h1", docs[0].Hash)

	folders, err := search.DocumentFolders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"acme"}, folders)

	hits, err := search.Search(ctx, storage.SearchQuery{Terms: []string{"dai"}})
	assert.NoError(t, err)
	assert.Len(t, hits, 1)
	assert.Equal(t, "Run the daily deploy.", hits[0].Body)
	assert.Equal(t, int64(1), hits[0].Ordinal)

	none, err := search.Search(ctx, storage.SearchQuery{Terms: []string{"daily"}, Folders: []string{"other"}})
	assert.NoError(t, err)
	assert.Len(t, none, 0)

	outline, err := search.Outline(ctx, doc.ID)
	assert.NoError(t, err)
	assert.Len(t, outline.Sections, 2)
	assert.Equal(t, "Acme > Deploy", outline.Sections[1].Breadcrumb)
	assert.Equal(t, "", outline.Sections[1].Body, "the outline carries no text")

	section, err := search.Section(ctx, doc.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Run the daily deploy.", section.Body)

	browsed, err := search.Browse(ctx, "src/acme", nil)
	assert.NoError(t, err)
	assert.Len(t, browsed, 1)

	// A changed file keeps the ID the path had.
	id := doc.ID
	assert.NoError(t, search.ReplaceDocument(ctx, doc, sections, terms))
	assert.Equal(t, id, doc.ID)

	job := &storage.SearchDocument{Kind: storage.SearchJobKind, Path: "src/acme/atkins.yml#ci:daily", File: "src/acme/atkins.yml", FolderID: "acme", Workspace: "internal", Project: "acme", Target: "ci:daily", Name: "ci:daily", Title: "ci:daily", Hash: "j1"}
	assert.NoError(t, search.ReplaceDocument(ctx, job, []storage.SearchSection{{Level: 1, Heading: "ci:daily", Breadcrumb: "ci:daily", Line: 1, Body: "Run the whole gate."}}, []storage.SearchTerm{{Term: "daily", Weight: 4}}))

	targets, err := search.Targets(ctx, "ci:daily")
	assert.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Equal(t, "acme", targets[0].FolderID)

	jobs, err := search.JobDocuments(ctx)
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)

	profile, err := search.Profile(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), profile.Documents)
	assert.Equal(t, int64(3), profile.Sections)

	assert.NoError(t, search.DeleteDocument(ctx, job.ID))
	targets, err = search.Targets(ctx, "ci:daily")
	assert.NoError(t, err)
	assert.Len(t, targets, 0)

	assert.NoError(t, search.DeleteFolder(ctx, "acme"))
	docs, err = search.Documents(ctx, "acme")
	assert.NoError(t, err)
	assert.Len(t, docs, 0)
}

// A closed pool reaches every error path of the storage.
func TestSearchStorage_Closed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	search := storage.NewSearchStorage(closedDB(t))
	doc := &storage.SearchDocument{Path: "src/acme/README.md"}

	assert.ErrorContains(t, search.ReplaceDocument(ctx, doc, nil, nil), errClosed)
	_, err := search.Documents(ctx, "acme")
	assert.ErrorContains(t, err, errClosed)
	_, err = search.JobDocuments(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = search.DocumentFolders(ctx)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, search.DeleteDocument(ctx, "d1"), errClosed)
	assert.ErrorContains(t, search.DeleteFolder(ctx, "acme"), errClosed)
	_, err = search.Search(ctx, storage.SearchQuery{Terms: []string{"daily"}})
	assert.ErrorContains(t, err, errClosed)
	_, err = search.Outline(ctx, "d1")
	assert.ErrorContains(t, err, errClosed)
	_, err = search.Section(ctx, "d1", 0)
	assert.ErrorContains(t, err, errClosed)
	_, err = search.Browse(ctx, "src/acme", nil)
	assert.ErrorContains(t, err, errClosed)
	_, err = search.Targets(ctx, "ci:daily")
	assert.ErrorContains(t, err, errClosed)
	_, err = search.Profile(ctx)
	assert.ErrorContains(t, err, errClosed)
}
