package search_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// fakeStore keeps the index in a map, keyed by path.
type fakeStore struct {
	docs     map[string]storage.SearchDocument
	sections map[string][]storage.SearchSection
	terms    map[string][]storage.SearchTerm
	deleted  []string
	folders  []string
}

func newFakeStore() *fakeStore {
	return &fakeStore{docs: map[string]storage.SearchDocument{}, sections: map[string][]storage.SearchSection{}, terms: map[string][]storage.SearchTerm{}}
}

func (f *fakeStore) Documents(_ context.Context, folderID string) ([]storage.SearchDocument, error) {
	var out []storage.SearchDocument
	for _, doc := range f.docs {
		if doc.Kind == storage.SearchDocKind && doc.FolderID == folderID {
			out = append(out, doc)
		}
	}
	return out, nil
}

func (f *fakeStore) DocumentFolders(_ context.Context) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, doc := range f.docs {
		if doc.Kind == storage.SearchDocKind && !seen[doc.FolderID] {
			seen[doc.FolderID] = true
			out = append(out, doc.FolderID)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (f *fakeStore) JobDocuments(_ context.Context) ([]storage.SearchDocument, error) {
	var out []storage.SearchDocument
	for _, doc := range f.docs {
		if doc.Kind == storage.SearchJobKind {
			out = append(out, doc)
		}
	}
	return out, nil
}

func (f *fakeStore) ReplaceDocument(_ context.Context, doc *storage.SearchDocument, sections []storage.SearchSection, terms []storage.SearchTerm) error {
	type termKey struct {
		ordinal int64
		term    string
	}
	seen := map[termKey]bool{}
	for _, term := range terms {
		key := termKey{term.Ordinal, term.Term}
		if seen[key] {
			return fmt.Errorf("duplicate term %d %q", term.Ordinal, term.Term)
		}
		seen[key] = true
	}
	if old, ok := f.docs[doc.Path]; ok {
		doc.ID = old.ID
	}
	if doc.ID == "" {
		doc.ID = "id-" + doc.Path
	}
	f.docs[doc.Path] = *doc
	f.sections[doc.Path] = sections
	f.terms[doc.Path] = terms
	return nil
}

func (f *fakeStore) DeleteDocument(_ context.Context, id string) error {
	for path, doc := range f.docs {
		if doc.ID == id {
			delete(f.docs, path)
			delete(f.sections, path)
			delete(f.terms, path)
			f.deleted = append(f.deleted, id)
		}
	}
	return nil
}

func (f *fakeStore) DeleteFolder(_ context.Context, folderID string) error {
	f.folders = append(f.folders, folderID)
	for path, doc := range f.docs {
		if doc.FolderID == folderID && doc.Kind == storage.SearchDocKind {
			delete(f.docs, path)
			delete(f.sections, path)
			delete(f.terms, path)
		}
	}
	return nil
}

// A run indexes markdown and text, skips an unchanged file, replaces a changed
// one, deletes one that is gone, and indexes a job target with its words.
func TestIndexer_Run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := t.TempDir()
	project := filepath.Join(root, "src", "acme")
	write(t, filepath.Join(project, "README.md"), "# Acme\n\nIntro.\n\n## Install\n\nRun it.\n")
	write(t, filepath.Join(project, "docs", "guide.txt"), "Daily deploy.\n")
	write(t, filepath.Join(project, ".git", "hidden.md"), "# Hidden\n")
	write(t, filepath.Join(project, "node_modules", "pkg", "README.md"), "# Dependency\n")

	store := newFakeStore()
	index := search.NewIndexer(store, root)
	index.Sources = func(context.Context) ([]search.Source, error) {
		return []search.Source{{Workspace: "internal", Project: "acme", FolderID: "acme", Dir: project}}, nil
	}
	index.Targets = func(context.Context) ([]search.Target, error) {
		return []search.Target{{
			FolderID: "acme", FolderTitle: "Acme", Workspace: "internal", Project: "acme",
			ID: "ci:daily", Name: "ci:daily", Kind: "atkins", SourceFile: "src/acme/atkins.yml",
			Description: "Run the whole gate.", Steps: []string{"build", "test"}, Command: []string{"atkins", "ci:daily"},
		}}, nil
	}
	after := 0
	index.After = func(context.Context) error { after++; return nil }

	stats, err := index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, stats.Documents)
	assert.Equal(t, 2, stats.Indexed)
	assert.Equal(t, 1, stats.Jobs)
	assert.Equal(t, 1, after)

	readme := store.docs["src/acme/README.md"]
	assert.Equal(t, "Acme", readme.Title)
	sections := store.sections["src/acme/README.md"]
	assert.Len(t, sections, 2)
	assert.Equal(t, "Acme > Install", sections[1].Breadcrumb)
	terms := store.terms["src/acme/README.md"]
	assert.True(t, hasTerm(terms, 1, "install"))

	job := store.docs["src/acme/atkins.yml#ci:daily"]
	assert.Equal(t, storage.SearchJobKind, job.Kind)
	assert.Equal(t, "ci:daily", job.Target)
	assert.Contains(t, store.sections["src/acme/atkins.yml#ci:daily"][0].Body, "atkins ci:daily")
	type termKey struct {
		ordinal int64
		term    string
	}
	seen := map[termKey]bool{}
	for _, row := range store.terms["src/acme/atkins.yml#ci:daily"] {
		key := termKey{row.Ordinal, row.Term}
		assert.True(t, !seen[key], "a word is one row of a section")
		seen[key] = true
	}

	stats, err = index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.Indexed)
	assert.Equal(t, 3, stats.Skipped)

	write(t, filepath.Join(project, "docs", "guide.txt"), "Daily deploy, and more.\n")
	stats, err = index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Indexed)

	assert.NoError(t, os.Remove(filepath.Join(project, "README.md")))
	stats, err = index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Deleted)

	index.Sources = nil
	stats, err = index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Deleted)
	assert.Len(t, store.docs, 1, "only the job remains")

	index.Targets = nil
	stats, err = index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.JobsDeleted)
	assert.Len(t, store.docs, 0)
}

// A document emptied of its text leaves the index.
func TestIndexer_Emptied(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := t.TempDir()
	project := filepath.Join(root, "src", "acme")
	write(t, filepath.Join(project, "README.md"), "# Acme\n\nIntro.\n")

	store := newFakeStore()
	index := search.NewIndexer(store, root)
	index.Sources = func(context.Context) ([]search.Source, error) {
		return []search.Source{{FolderID: "acme", Dir: project}}, nil
	}
	_, err := index.Run(ctx)
	assert.NoError(t, err)
	assert.Len(t, store.docs, 1)

	write(t, filepath.Join(project, "README.md"), "")
	stats, err := index.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Deleted)
	assert.Len(t, store.docs, 0)
}

// Trigger starts a run in the background and Wait returns when it ends.
func TestIndexer_Trigger(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, "src", "acme", "README.md"), "# Acme\n")
	store := newFakeStore()
	index := search.NewIndexer(store, root)
	index.Sources = func(context.Context) ([]search.Source, error) {
		return []search.Source{{FolderID: "acme", Dir: filepath.Join(root, "src", "acme")}}, nil
	}
	assert.NoError(t, index.Trigger(context.Background()))
	index.Wait()
	assert.Len(t, store.docs, 1)
}

// A nil indexer, and one with no store, index nothing.
func TestIndexer_None(t *testing.T) {
	t.Parallel()
	stats, err := (*search.Indexer)(nil).Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.Documents)
	assert.NoError(t, (*search.Indexer)(nil).Trigger(context.Background()))
	assert.NotNil(t, search.NewIndexer(nil, ""))
}

// The store interface is what the indexer writes through.
func TestStore(t *testing.T) {
	t.Parallel()
	var _ search.Store = (*fakeStore)(nil)
}

func write(t *testing.T, path, body string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, []byte(body), 0o600))
}

func hasTerm(terms []storage.SearchTerm, ordinal int64, term string) bool {
	for _, row := range terms {
		if row.Ordinal == ordinal && row.Term == term {
			return true
		}
	}
	return false
}
