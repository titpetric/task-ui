package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/titpetric/task-ui/storage"
)

// maxSection bounds the text search_read returns for one section.
const maxSection = 32 << 10

// searchTools read the documentation and job target index, a section at a time.
func searchTools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	num := func(text string) Property { return Property{Type: "integer", Description: text} }
	return []Tool{
		{
			Name: "search_docs",
			Description: "Search the documentation and job targets of every project the user sees. Returns matching sections with a short excerpt, best first. " +
				"Read one with search_read, or list a document's headings with search_outline, which carries no text.",
			InputSchema: object(map[string]Property{
				"query": str("Words to find, such as daily deploy or ci:daily."),
				"limit": num("How many sections, 20 unless given."),
			}, "query"),
			call: searchDocs,
		},
		{
			Name:        "search_outline",
			Description: "List a document's headings in order, with the number search_read takes. It carries no paragraph text, so it is how a document is read before a section is chosen.",
			InputSchema: object(map[string]Property{
				"id": str("The document ID a search_docs hit names."),
			}, "id"),
			call: searchOutline,
		},
		{
			Name:        "search_read",
			Description: "Read one section of a document, named by the document and the section number search_outline lists. Returns the section's text and nothing around it.",
			InputSchema: object(map[string]Property{
				"id":      str("The document ID a search_docs hit names."),
				"section": num("The section number search_outline lists; 0 when omitted."),
			}, "id"),
			call: searchRead,
		},
		{
			Name:        "search_browse",
			Description: "List the folders and documents indexed under a path, with job targets marked. It is how the folder structure is read when the query names a place rather than a topic.",
			InputSchema: object(map[string]Property{
				"path": str("A path under the root, such as src/acme; the top of the index unless given."),
			}),
			call: searchBrowse,
		},
	}
}

type searchHitView struct {
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
}

type searchDocumentView struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Folder  string `json:"folder"`
	Target  string `json:"target"`
	Name    string `json:"name"`
	Dir     string `json:"dir"`
	Title   string `json:"title"`
	Ordinal int64  `json:"ordinal"`
	Level   int64  `json:"level"`
	Heading string `json:"heading"`
}

type searchSectionView struct {
	Ordinal    int64  `json:"ordinal"`
	Heading    string `json:"heading"`
	Breadcrumb string `json:"breadcrumb"`
}

type searchArgs struct {
	Query   string `json:"query"`
	Limit   int    `json:"limit"`
	ID      string `json:"id"`
	Section int64  `json:"section"`
	Path    string `json:"path"`
}

func searchDocs(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in searchArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return "", fmt.Errorf("search_docs needs a query")
	}
	values := url.Values{"q": {query}}
	if in.Limit > 0 {
		values.Set("limit", fmt.Sprint(in.Limit))
	}
	var out struct {
		Hits []searchHitView `json:"hits"`
	}
	if err := c.get(ctx, "/api/search?"+values.Encode(), &out); err != nil {
		return "", err
	}
	if len(out.Hits) == 0 {
		return "Nothing in the index matches " + query + ".", nil
	}
	var text strings.Builder
	fmt.Fprintf(&text, "%d sections for %s:\n", len(out.Hits), query)
	for _, hit := range out.Hits {
		if hit.Kind == storage.SearchJobKind {
			fmt.Fprintf(&text, "\n[job] %s %s\n", hit.Folder, hit.Name)
			if hit.Title != "" && hit.Title != hit.Name {
				fmt.Fprintf(&text, "  %s\n", hit.Title)
			}
			fmt.Fprintf(&text, "  document %s section %d\n", hit.Document, hit.Ordinal)
		} else {
			place := hit.Path
			if hit.Breadcrumb != "" {
				place += " > " + hit.Breadcrumb
			}
			fmt.Fprintf(&text, "\n[doc] %s (line %d)\n", place, hit.Line)
			fmt.Fprintf(&text, "  document %s section %d\n", hit.Document, hit.Ordinal)
		}
		if hit.Snippet != "" {
			fmt.Fprintf(&text, "  %s\n", hit.Snippet)
		}
	}
	return cut(text.String(), maxSection), nil
}

func searchOutline(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in searchArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", fmt.Errorf("search_outline needs a document ID")
	}
	var out struct {
		Document searchDocumentView   `json:"document"`
		Sections []searchDocumentView `json:"sections"`
	}
	if err := c.get(ctx, "/api/search/documents/"+url.PathEscape(in.ID), &out); err != nil {
		return "", err
	}
	var text strings.Builder
	fmt.Fprintf(&text, "%s (%s)\n", out.Document.Path, out.Document.Kind)
	for _, section := range out.Sections {
		fmt.Fprintf(&text, "\n%d %s%s", section.Ordinal, strings.Repeat("  ", int(section.Level)), section.Heading)
	}
	text.WriteByte('\n')
	return text.String(), nil
}

func searchRead(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in searchArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", fmt.Errorf("search_read needs a document ID")
	}
	var out struct {
		Document searchDocumentView `json:"document"`
		Section  searchSectionView  `json:"section"`
		Body     string             `json:"body"`
	}
	path := fmt.Sprintf("/api/search/documents/%s/sections/%d", url.PathEscape(in.ID), in.Section)
	if err := c.get(ctx, path, &out); err != nil {
		return "", err
	}
	head := out.Document.Path
	if out.Section.Breadcrumb != "" {
		head += " > " + out.Section.Breadcrumb
	}
	return cut(head+"\n\n"+out.Body+"\n", maxSection), nil
}

func searchBrowse(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in searchArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	path := strings.Trim(in.Path, "/")
	values := url.Values{}
	if path != "" {
		values.Set("path", path)
	}
	var out struct {
		Folders   []string             `json:"folders"`
		Documents []searchDocumentView `json:"documents"`
	}
	if err := c.get(ctx, "/api/search/browse?"+values.Encode(), &out); err != nil {
		return "", err
	}
	var text strings.Builder
	label := "/" + path
	if path == "" {
		label = "/"
	}
	fmt.Fprintf(&text, "%s (%d folders, %d documents)\n", label, len(out.Folders), len(out.Documents))
	for _, folder := range out.Folders {
		fmt.Fprintf(&text, "  %s/\n", folder)
	}
	for _, doc := range out.Documents {
		if doc.Kind == storage.SearchJobKind {
			fmt.Fprintf(&text, "  [job] %s - %s\n", doc.Name, doc.Title)
			continue
		}
		fmt.Fprintf(&text, "  %s\n", doc.Path)
	}
	return text.String(), nil
}

// cut bounds a tool result at size bytes on a line boundary.
func cut(text string, size int) string {
	if len(text) <= size {
		return text
	}
	text = text[:size]
	if at := strings.LastIndexByte(text, '\n'); at > 0 {
		text = text[:at]
	}
	return text + "\n... (cut)"
}
