package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

// maxRead bounds the text read_artifact returns.
const maxRead = 256 << 10

// artifactTools read and write the artifact tree, so an agent outside
// task-ui keeps its files beside the ones runs leave, where only runs
// wrote (docs/api.md).
func artifactTools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	return []Tool{
		{
			Name: "list_artifacts",
			Description: "List a directory of the artifact tree, the files runs leave and users write: the latest version of each file, " +
				"with its size, when it changed, and the run or the user that wrote it.",
			InputSchema: object(map[string]Property{"path": str("The directory, such as reports; the top of the tree unless given.")}),
			call:        listArtifacts,
		},
		{
			Name: "read_artifact",
			Description: "Read a file of the artifact tree: text as text, up to 256 KiB, and for anything else its type, size and download link. " +
				"The latest version unless a version is given.",
			InputSchema: object(map[string]Property{
				"path":    str("The file, such as reports/coverage.txt."),
				"version": str("A version, a commit list_artifacts or a reply named; the latest unless given."),
			}, "path"),
			call: readArtifact,
		},
		{
			Name: "write_artifact",
			Description: "Write a file into the artifact tree as you, at most 16 MiB, over the version there, which the history keeps. " +
				"A relative path with no .. and no dot file. Takes artifact/edit.",
			InputSchema: object(map[string]Property{
				"path":     str("The file, such as notes/plan.md."),
				"content":  str("The file's content."),
				"encoding": {Type: "string", Description: "How content is written: text unless given, or base64 for a binary file.", Enum: []string{"text", "base64"}},
			}, "path", "content"),
			call: writeArtifact,
		},
	}
}

// escape escapes each element of a slash path for a URL.
func escape(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

type pathArgs struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

func listArtifacts(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in pathArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	dir := strings.Trim(in.Path, "/")
	route := "/artifacts/tree"
	if dir != "" {
		route += "/" + escape(dir)
	}
	var view api.ArtifactTreeView
	if err := c.get(ctx, route, &view); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "/%s (%d)\n", view.Path, len(view.Entries))
	for _, entry := range view.Entries {
		if entry.Dir {
			fmt.Fprintf(&b, "  %s/  %s\n", entry.Name, size(entry.ByteSize))
			continue
		}
		by := "written by " + entry.UserName
		if entry.RunID != "" {
			by = fmt.Sprintf("run %s of %s/%s", entry.RunID, entry.Folder, entry.Job)
		}
		fmt.Fprintf(&b, "  %s  %s, %s, %s\n", entry.Name, size(entry.ByteSize), entry.Modified.UTC().Format("2006-01-02 15:04"), by)
	}
	if len(view.Entries) == 0 {
		b.WriteString("  empty\n")
	}
	return b.String(), nil
}

func readArtifact(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in pathArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	rel := strings.Trim(in.Path, "/")
	if rel == "" {
		return "", errors.New("a path is needed")
	}
	q := url.Values{}
	if in.Version != "" {
		q.Set("version", in.Version)
	}
	route := "/artifacts/files/" + escape(rel) + "?raw"
	if len(q) > 0 {
		route += "&" + q.Encode()
	}
	data, header, err := c.send(ctx, http.MethodGet, route, "", nil)
	if err != nil {
		return "", err
	}
	version := header.Get("X-Artifact-Version")
	kind := header.Get("Content-Type")
	head := fmt.Sprintf("%s, version %s, %s, %s", rel, short(version), kind, size(int64(len(data))))
	text, cut := textOf(data, kind)
	if text == nil {
		link := c.base + "/artifacts/files/" + escape(rel) + "?download"
		if version != "" {
			link += "&version=" + url.QueryEscape(version)
		}
		return head + ", not text.\nDownload: " + link + "\n", nil
	}
	out := head + "\n\n" + string(text)
	if cut {
		out += fmt.Sprintf("\n\n[cut at %s of %s]\n", size(int64(len(text))), size(int64(len(data))))
	}
	return out, nil
}

// textOf returns a file's text, up to maxRead and cut at a character, and
// whether it cut; nil for a file that is not text.
func textOf(data []byte, kind string) ([]byte, bool) {
	if strings.HasPrefix(kind, "image/") || strings.ContainsRune(string(data[:min(len(data), 8000)]), 0) {
		return nil, false
	}
	if len(data) <= maxRead {
		if !utf8.Valid(data) {
			return nil, false
		}
		return data, false
	}
	// The cut may split a character: back off to the start of it.
	text := data[:maxRead]
	for i := 0; i < utf8.UTFMax && !utf8.Valid(text); i++ {
		text = text[:len(text)-1]
	}
	if !utf8.Valid(text) {
		return nil, false
	}
	return text, true
}

// short is a commit's first seven characters.
func short(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	if commit == "" {
		return "none"
	}
	return commit
}

type writeArgs struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func writeArtifact(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in writeArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.Path == "" {
		return "", errors.New("a path is needed")
	}
	data := []byte(in.Content)
	switch in.Encoding {
	case "", "text":
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(in.Content)
		if err != nil {
			return "", fmt.Errorf("content is not base64: %w", err)
		}
		data = decoded
	default:
		return "", errors.New("encoding is text or base64")
	}
	out, _, err := c.send(ctx, http.MethodPut, "/artifacts/files/"+escape(in.Path), "application/octet-stream", data)
	if err != nil {
		return "", err
	}
	var artifact model.Artifact
	if err := json.Unmarshal(out, &artifact); err != nil {
		return "", err
	}
	return fmt.Sprintf("Wrote %s as %s, %s, version %s; the history keeps the version before.\n", artifact.Path, artifact.UserName, size(artifact.ByteSize), short(artifact.CommitID)), nil
}

// size formats a byte count, as the pages do.
func size(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, suffix := float64(n)/unit, "KB"
	for _, next := range []string{"MB", "GB"} {
		if value < unit {
			break
		}
		value, suffix = value/unit, next
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}
