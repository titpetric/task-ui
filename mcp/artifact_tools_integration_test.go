//go:build integration

package mcp_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A runner writes a file into the artifact tree, lists it and reads it
// back; writing again keeps the version before; a viewer reads and may
// not write; the tool refuses .. and dot files.
func TestArtifactTools_RoundTrip(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "ops", "write_artifact", map[string]any{"path": "notes/plan.md", "content": "# Plan\n"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Wrote notes/plan.md as ops, 7 B, version ")
	first := strings.Fields(strings.SplitAfter(text, "version ")[1])[0]
	first = strings.TrimSuffix(first, ";")

	text, failed = f.tool(t, "ops", "write_artifact", map[string]any{"path": "notes/plan.md", "content": "# Plan, revised\n"})
	assert.False(t, failed, text)

	text, failed = f.tool(t, "guest", "list_artifacts", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "  notes/  ")
	text, _ = f.tool(t, "guest", "list_artifacts", map[string]any{"path": "notes"})
	assert.Contains(t, text, "  plan.md  16 B, ")
	assert.Contains(t, text, "written by ops")

	text, failed = f.tool(t, "guest", "read_artifact", map[string]any{"path": "notes/plan.md"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "# Plan, revised")
	text, _ = f.tool(t, "guest", "read_artifact", map[string]any{"path": "notes/plan.md", "version": first})
	assert.Contains(t, text, "version "+first)
	assert.Contains(t, text, "\n\n# Plan\n")

	text, failed = f.tool(t, "guest", "write_artifact", map[string]any{"path": "x.txt", "content": "no"})
	assert.True(t, failed)
	assert.Contains(t, text, "artifact/edit permission (403)")
	for _, path := range []string{"../out.txt", "a/../../b", ".env", "notes/.secret"} {
		text, failed = f.tool(t, "ops", "write_artifact", map[string]any{"path": path, "content": "no"})
		assert.True(t, failed, path)
		assert.Contains(t, text, "(400)", path)
	}
	_, failed = f.tool(t, "guest", "read_artifact", map[string]any{"path": "nothing.txt"})
	assert.True(t, failed)
}

// A binary file reads as its type, size and download link; the tool
// cuts a long text file at 256 KiB and says so.
func TestArtifactTools_Read(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	png := base64.StdEncoding.EncodeToString([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))
	text, failed := f.tool(t, "admin", "write_artifact", map[string]any{"path": "plot.png", "content": png, "encoding": "base64"})
	assert.False(t, failed, text)
	text, failed = f.tool(t, "guest", "read_artifact", map[string]any{"path": "plot.png"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "plot.png, version ")
	assert.Contains(t, text, "image/png, 16 B, not text.")
	assert.Contains(t, text, "Download: http://")
	assert.Contains(t, text, "/artifacts/files/plot.png?download&version=")

	_, failed = f.tool(t, "admin", "write_artifact", map[string]any{"path": "bad.bin", "content": "not base64!", "encoding": "base64"})
	assert.True(t, failed)

	long := strings.Repeat("é", 200<<10)
	_, failed = f.tool(t, "admin", "write_artifact", map[string]any{"path": "long.txt", "content": long})
	assert.False(t, failed)
	text, _ = f.tool(t, "guest", "read_artifact", map[string]any{"path": "long.txt"})
	assert.Contains(t, text, "[cut at 256.0 KB of 400.0 KB]")
}
