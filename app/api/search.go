package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/storage"
)

// searchDocument is a document as the search routes name it, without its hash.
type searchDocument struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	Folder    string `json:"folder"`
	Workspace string `json:"workspace"`
	Project   string `json:"project"`
	Target    string `json:"target"`
	Name      string `json:"name"`
	Dir       string `json:"dir"`
	Title     string `json:"title"`
}

// searchHit is one matching section with the text around the match.
type searchHit struct {
	Document   string `json:"document"`
	Ordinal    int64  `json:"ordinal"`
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	Title      string `json:"title"`
	Folder     string `json:"folder"`
	Target     string `json:"target"`
	Name       string `json:"name"`
	Heading    string `json:"heading"`
	Breadcrumb string `json:"breadcrumb"`
	Line       int64  `json:"line"`
	Snippet    string `json:"snippet"`
	Score      int64  `json:"score"`
}

// searchSection is one section of a document, without its text.
type searchSection struct {
	Ordinal    int64  `json:"ordinal"`
	Level      int64  `json:"level"`
	Heading    string `json:"heading"`
	Breadcrumb string `json:"breadcrumb"`
	Line       int64  `json:"line"`
}

func documentView(doc storage.SearchDocument) searchDocument {
	return searchDocument{
		ID: doc.ID, Kind: doc.Kind, Path: doc.Path, Folder: doc.FolderID, Workspace: doc.Workspace,
		Project: doc.Project, Target: doc.Target, Name: doc.Name, Dir: doc.Dir, Title: doc.Title,
	}
}

// searchDocs returns the sections holding the query, best score first, in the folders the user sees.
func (h *Handlers) searchDocs(w http.ResponseWriter, r *http.Request) {
	if h.opts.Search == nil {
		fail(w, http.StatusNotFound, "search isn't available on this instance")
		return
	}
	if !h.require(w, r, "prompt/view") {
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	terms := search.Unique(search.Tokenize(query))
	hits := []searchHit{}
	if len(terms) > 0 {
		found, err := h.opts.Search.Search(r.Context(), storage.SearchQuery{
			Terms: terms, Folders: access.UserFrom(r.Context()).FolderList(), Limit: queryInt(r, "limit", 20),
		})
		if err != nil {
			failErr(w, err)
			return
		}
		for _, hit := range found {
			hits = append(hits, searchHit{
				Document: hit.ID, Ordinal: hit.Ordinal, Kind: hit.Kind, Path: hit.Path, Title: hit.Title,
				Folder: hit.FolderID, Target: hit.Target, Name: hit.Name, Heading: hit.Heading,
				Breadcrumb: hit.Breadcrumb, Line: hit.Line, Score: hit.Score,
				Snippet: search.Snippet(hit.Body, terms, 160),
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"query": query, "hits": hits, "count": len(hits)})
}

// getSearchDocument returns a document's outline, its headings and no text.
func (h *Handlers) getSearchDocument(w http.ResponseWriter, r *http.Request) {
	if h.opts.Search == nil {
		fail(w, http.StatusNotFound, "search isn't available on this instance")
		return
	}
	if !h.require(w, r, "prompt/view") {
		return
	}
	outline, err := h.opts.Search.Outline(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.seesSearchDocument(r, outline.Document) {
		failErr(w, storage.ErrNotFound)
		return
	}
	sections := make([]searchSection, 0, len(outline.Sections))
	for _, section := range outline.Sections {
		sections = append(sections, searchSection{Ordinal: section.Ordinal, Level: section.Level, Heading: section.Heading, Breadcrumb: section.Breadcrumb, Line: section.Line})
	}
	writeJSON(w, http.StatusOK, map[string]any{"document": documentView(outline.Document), "sections": sections})
}

// getSearchSection returns one section of a document with its text.
func (h *Handlers) getSearchSection(w http.ResponseWriter, r *http.Request) {
	if h.opts.Search == nil {
		fail(w, http.StatusNotFound, "search isn't available on this instance")
		return
	}
	if !h.require(w, r, "prompt/view") {
		return
	}
	ordinal, err := strconv.ParseInt(param(r, "ordinal"), 10, 64)
	if err != nil {
		fail(w, http.StatusBadRequest, "the section is a number")
		return
	}
	section, err := h.opts.Search.Section(r.Context(), param(r, "id"), ordinal)
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.seesSearchDocument(r, section.SearchDocument) {
		failErr(w, storage.ErrNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"document": documentView(section.SearchDocument),
		"section":  searchSection{Ordinal: section.Ordinal, Level: 0, Heading: section.Heading, Breadcrumb: section.Breadcrumb, Line: section.Line},
		"body":     section.Body,
	})
}

// browseSearch returns the folders and files directly under a path.
func (h *Handlers) browseSearch(w http.ResponseWriter, r *http.Request) {
	if h.opts.Search == nil {
		fail(w, http.StatusNotFound, "search isn't available on this instance")
		return
	}
	if !h.require(w, r, "prompt/view") {
		return
	}
	prefix := strings.Trim(strings.TrimSpace(r.URL.Query().Get("path")), "/")
	docs, err := h.opts.Search.Browse(r.Context(), prefix, access.UserFrom(r.Context()).FolderList())
	if err != nil {
		failErr(w, err)
		return
	}
	folders := []string{}
	files := []searchDocument{}
	for _, doc := range docs {
		rel := strings.TrimPrefix(doc.Path, prefix)
		rel = strings.TrimPrefix(rel, "/")
		if name, _, found := strings.Cut(rel, "/"); found {
			folders = append(folders, joinPath(prefix, name))
			continue
		}
		if rel != "" {
			files = append(files, documentView(doc))
		}
	}
	sort.Strings(folders)
	folders = uniqueStrings(folders)
	writeJSON(w, http.StatusOK, map[string]any{"path": prefix, "folders": folders, "documents": files})
}

// reindexSearch starts the index, for an admin.
func (h *Handlers) reindexSearch(w http.ResponseWriter, r *http.Request) {
	if user := access.UserFrom(r.Context()); user == nil || !user.CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin reindexes the search")
		return
	}
	if h.opts.Index == nil {
		fail(w, http.StatusServiceUnavailable, "this instance has no index")
		return
	}
	if err := h.opts.Index(r.Context()); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"started": true})
}

// seesSearchDocument reports whether the user may see the folder the document belongs to.
func (h *Handlers) seesSearchDocument(r *http.Request, doc storage.SearchDocument) bool {
	user := access.UserFrom(r.Context())
	if user == nil {
		return false
	}
	return user.CanSee(doc.FolderID)
}

// joinPath joins a browse prefix and a child directory.
func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

// uniqueStrings keeps one of each string, in order.
func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
