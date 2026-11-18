package storage_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The search types keep their fields, and every query is a method of the storage that owns it.
func TestSearchDocument(t *testing.T) {
	t.Parallel()
	doc := storage.SearchDocument{ID: "d1", Kind: storage.SearchJobKind, Path: "src/acme/atkins.yml#ci:daily", FolderID: "acme", Target: "ci:daily", Name: "ci:daily"}
	section := storage.SearchSection{DocumentID: "d1", Ordinal: 1, Level: 1, Heading: "ci:daily", Breadcrumb: "ci:daily", Line: 1, Body: "Run the whole gate."}
	term := storage.SearchTerm{DocumentID: "d1", Ordinal: 0, Term: "daily", Weight: 4}
	query := storage.SearchQuery{Terms: []string{"daily"}, Folders: []string{"acme"}, Limit: 5}
	hit := storage.SearchHit{SearchDocument: doc, Ordinal: 0, Heading: "ci:daily", Score: 4}
	outline := storage.SearchOutline{Document: doc, Sections: []storage.SearchSection{section}}
	profile := storage.SearchProfile{Documents: 1, Sections: 1, Terms: 1}

	assert.Equal(t, "ci:daily", doc.Target)
	assert.Equal(t, int64(1), section.Ordinal)
	assert.Equal(t, "daily", term.Term)
	assert.Equal(t, 5, query.Limit)
	assert.Equal(t, "d1", hit.ID)
	assert.Len(t, outline.Sections, 1)
	assert.Equal(t, int64(1), profile.Documents)
}
