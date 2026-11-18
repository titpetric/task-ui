package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
)

// The kinds a search document is: a markdown or text file, or a job target a folder declares.
const (
	SearchDocKind = "doc"
	SearchJobKind = "job"
)

// SearchDocument is one indexed file or job target.
type SearchDocument struct {
	ID         string     `db:"id" json:"id"`
	Kind       string     `db:"kind" json:"kind"`
	Path       string     `db:"path" json:"path"`
	File       string     `db:"file" json:"file"`
	FolderID   string     `db:"folder_id" json:"folder"`
	Workspace  string     `db:"workspace" json:"workspace"`
	Project    string     `db:"project" json:"project"`
	Target     string     `db:"target" json:"target"`
	Name       string     `db:"name" json:"name"`
	Dir        string     `db:"dir" json:"dir"`
	Title      string     `db:"title" json:"title"`
	Hash       string     `db:"hash" json:"hash"`
	Size       int64      `db:"size" json:"size"`
	ModifiedAt *time.Time `db:"modified_at" json:"modified_at"`
	IndexedAt  *time.Time `db:"indexed_at" json:"indexed_at"`
}

// SearchSection is one heading and the text under it.
type SearchSection struct {
	DocumentID string `db:"document_id" json:"document"`
	Ordinal    int64  `db:"ordinal" json:"ordinal"`
	Level      int64  `db:"level" json:"level"`
	Heading    string `db:"heading" json:"heading"`
	Breadcrumb string `db:"breadcrumb" json:"breadcrumb"`
	Line       int64  `db:"line" json:"line"`
	Body       string `db:"body" json:"body"`
}

// SearchTerm is one row of the inverted index.
type SearchTerm struct {
	DocumentID string `db:"document_id"`
	Ordinal    int64  `db:"ordinal"`
	Term       string `db:"term"`
	Weight     int64  `db:"weight"`
}

// SearchQuery holds the terms, the folders the reader may see, and the hit limit.
type SearchQuery struct {
	Terms   []string
	Folders []string
	Limit   int
}

// SearchHit is one matching section with the document it belongs to.
type SearchHit struct {
	SearchDocument
	Ordinal    int64  `db:"ordinal" json:"ordinal"`
	Heading    string `db:"heading" json:"heading"`
	Breadcrumb string `db:"breadcrumb" json:"breadcrumb"`
	Line       int64  `db:"line" json:"line"`
	Body       string `db:"body" json:"body"`
	Score      int64  `db:"score" json:"score"`
}

// SearchOutline is a document's sections, without their text.
type SearchOutline struct {
	Document SearchDocument  `json:"document"`
	Sections []SearchSection `json:"sections"`
}

// SearchProfile is what the index holds.
type SearchProfile struct {
	Documents int64  `db:"documents"`
	Sections  int64  `db:"sections"`
	Terms     int64  `db:"terms"`
	IndexedAt string `db:"indexed_at"`
}

// SearchStorage keeps the documentation and job target index.
type SearchStorage struct {
	db DB
}

// NewSearchStorage returns the storage over a pool.
func NewSearchStorage(db DB) *SearchStorage {
	return &SearchStorage{db: db}
}

type searchID struct {
	ID string `db:"id"`
}

// ReplaceDocument writes a document, its sections and their terms, keeping the
// ID an earlier index gave the path.
func (s *SearchStorage) ReplaceDocument(ctx context.Context, doc *SearchDocument, sections []SearchSection, terms []SearchTerm) error {
	return inTx(ctx, s.db, func(c conn) error {
		existing, err := c.Get[searchID](ctx, "SELECT id FROM search_document WHERE path = ?", doc.Path)
		switch {
		case err == nil:
			doc.ID = existing.ID
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}
		if doc.ID == "" {
			doc.ID = model.NewID()
		}
		if doc.Kind == "" {
			doc.Kind = SearchDocKind
		}
		doc.IndexedAt = stamp()
		if err := c.Exec(ctx, "DELETE FROM search_term WHERE document_id = ?", doc.ID); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM search_section WHERE document_id = ?", doc.ID); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM search_document WHERE id = ? OR path = ?", doc.ID, doc.Path); err != nil {
			return err
		}
		if err := c.Insert(ctx, "search_document", doc); err != nil {
			return err
		}
		for i := range sections {
			sections[i].DocumentID = doc.ID
			if err := c.Insert(ctx, "search_section", &sections[i]); err != nil {
				return err
			}
		}
		for i := range terms {
			terms[i].DocumentID = doc.ID
			if err := c.Insert(ctx, "search_term", &terms[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// Documents returns the file documents of a folder, for the hash comparison and the sweep.
func (s *SearchStorage) Documents(ctx context.Context, folderID string) ([]SearchDocument, error) {
	return client(s.db).Select[SearchDocument](ctx, "SELECT * FROM search_document WHERE folder_id = ? AND kind = ? ORDER BY path", folderID, SearchDocKind)
}

// JobDocuments returns every job target document, for the sweep.
func (s *SearchStorage) JobDocuments(ctx context.Context) ([]SearchDocument, error) {
	return client(s.db).Select[SearchDocument](ctx, "SELECT * FROM search_document WHERE kind = ? ORDER BY path", SearchJobKind)
}

// DocumentFolders returns the folders the index holds documents of, for a folder that left the catalogue.
func (s *SearchStorage) DocumentFolders(ctx context.Context) ([]string, error) {
	rows, err := client(s.db).Select[searchFolder](ctx, "SELECT DISTINCT folder_id FROM search_document WHERE kind = ? ORDER BY folder_id", SearchDocKind)
	if err != nil {
		return nil, err
	}
	folders := make([]string, 0, len(rows))
	for _, row := range rows {
		folders = append(folders, row.FolderID)
	}
	return folders, nil
}

type searchFolder struct {
	FolderID string `db:"folder_id"`
}

// DeleteDocument removes a document, the sections under it and the terms.
func (s *SearchStorage) DeleteDocument(ctx context.Context, id string) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "DELETE FROM search_term WHERE document_id = ?", id); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM search_section WHERE document_id = ?", id); err != nil {
			return err
		}
		return c.Exec(ctx, "DELETE FROM search_document WHERE id = ?", id)
	})
}

// DeleteFolder removes every document of a folder, for a project that left the catalogue.
func (s *SearchStorage) DeleteFolder(ctx context.Context, folderID string) error {
	rows, err := client(s.db).Select[searchID](ctx, "SELECT id FROM search_document WHERE folder_id = ?", folderID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := s.DeleteDocument(ctx, row.ID); err != nil {
			return err
		}
	}
	return nil
}

// Search returns the sections that hold the terms, best score first, in the folders given.
func (s *SearchStorage) Search(ctx context.Context, q SearchQuery) ([]SearchHit, error) {
	terms := normalizeTerms(q.Terms)
	if len(terms) == 0 {
		return nil, nil
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	where := make([]string, 0, len(terms))
	args := make([]any, 0, len(terms)*2+len(q.Folders)+1)
	for _, term := range terms {
		where = append(where, "(t.term = ? OR t.term LIKE ? || '%')")
		args = append(args, term, term)
	}
	query := `SELECT d.id AS id, d.kind AS kind, d.path AS path, d.file AS file, d.folder_id AS folder_id,
		d.workspace AS workspace, d.project AS project, d.target AS target, d.name AS name, d.dir AS dir,
		d.title AS title, d.hash AS hash, d.size AS size, d.modified_at AS modified_at, d.indexed_at AS indexed_at,
		s.ordinal AS ordinal, s.heading AS heading, s.breadcrumb AS breadcrumb, s.line AS line, s.body AS body,
		SUM(t.weight) AS score
		FROM search_term t
		JOIN search_section s ON s.document_id = t.document_id AND s.ordinal = t.ordinal
		JOIN search_document d ON d.id = t.document_id
		WHERE (` + strings.Join(where, " OR ") + `)`
	if len(q.Folders) > 0 {
		query += " AND d.folder_id IN (?" + strings.Repeat(", ?", len(q.Folders)-1) + ")"
		args = append(args, anyList(q.Folders)...)
	}
	query += " GROUP BY d.id, s.ordinal ORDER BY score DESC, d.path, s.ordinal LIMIT ?"
	args = append(args, q.Limit)
	return client(s.db).Select[SearchHit](ctx, query, args...)
}

// Outline returns a document and its sections without their text.
func (s *SearchStorage) Outline(ctx context.Context, id string) (*SearchOutline, error) {
	doc, err := client(s.db).Get[SearchDocument](ctx, "SELECT * FROM search_document WHERE id = ?", id)
	if err != nil {
		return nil, notFound(err)
	}
	sections, err := client(s.db).Select[SearchSection](ctx, "SELECT document_id, ordinal, level, heading, breadcrumb, line FROM search_section WHERE document_id = ? ORDER BY ordinal", id)
	if err != nil {
		return nil, err
	}
	return &SearchOutline{Document: *doc, Sections: sections}, nil
}

// Section returns one section of a document with the document it belongs to.
func (s *SearchStorage) Section(ctx context.Context, id string, ordinal int64) (*SearchHit, error) {
	row, err := client(s.db).Get[SearchHit](ctx, `SELECT d.id AS id, d.kind AS kind, d.path AS path, d.file AS file, d.folder_id AS folder_id,
		d.workspace AS workspace, d.project AS project, d.target AS target, d.name AS name, d.dir AS dir,
		d.title AS title, d.hash AS hash, d.size AS size, d.modified_at AS modified_at, d.indexed_at AS indexed_at,
		s.ordinal AS ordinal, s.heading AS heading, s.breadcrumb AS breadcrumb, s.line AS line, s.body AS body, 0 AS score
		FROM search_section s JOIN search_document d ON d.id = s.document_id
		WHERE s.document_id = ? AND s.ordinal = ?`, id, ordinal)
	return row, notFound(err)
}

// Browse returns the documents under a path prefix, in the folders given.
func (s *SearchStorage) Browse(ctx context.Context, prefix string, folders []string) ([]SearchDocument, error) {
	query := "SELECT * FROM search_document WHERE path LIKE ? || '%'"
	args := []any{prefix}
	if len(folders) > 0 {
		query += " AND folder_id IN (?" + strings.Repeat(", ?", len(folders)-1) + ")"
		args = append(args, anyList(folders)...)
	}
	query += " ORDER BY path LIMIT 500"
	return client(s.db).Select[SearchDocument](ctx, query, args...)
}

// Targets returns the job documents of a target name.
func (s *SearchStorage) Targets(ctx context.Context, name string) ([]SearchDocument, error) {
	return client(s.db).Select[SearchDocument](ctx, "SELECT * FROM search_document WHERE kind = ? AND (target = ? OR name = ?) ORDER BY folder_id, dir, target", SearchJobKind, name, name)
}

// Profile counts what the index holds and when it last changed.
func (s *SearchStorage) Profile(ctx context.Context) (SearchProfile, error) {
	row, err := client(s.db).Get[SearchProfile](ctx, `SELECT
		(SELECT COUNT(*) FROM search_document) AS documents,
		(SELECT COUNT(*) FROM search_section) AS sections,
		(SELECT COUNT(*) FROM search_term) AS terms,
		(SELECT COALESCE(MAX(indexed_at), '') FROM search_document) AS indexed_at`)
	if err != nil {
		return SearchProfile{}, err
	}
	return *row, nil
}

// normalizeTerms lowercases, trims and deduplicates the terms of a query.
func normalizeTerms(terms []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" || seen[term] {
			continue
		}
		seen[term] = true
		out = append(out, term)
	}
	return out
}
