package search_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/assert"
)

// Headings start sections with their ancestors in the breadcrumb, a fenced
// block hides its hashes, setext headings count, and front matter is skipped.
func TestSplit(t *testing.T) {
	t.Parallel()
	src := []byte("---\ntitle: ignored\n---\n\nIntro text.\n\n# Acme\n\nFirst paragraph.\n\n## Install\n\n```sh\n# not a heading\n```\n\nStill install.\n\nDeploy\n------\n\nRun it.\n\n# After\n")

	title, sections := search.Split(src, "README.md")
	assert.Equal(t, "Acme", title)
	assert.Len(t, sections, 5)

	assert.Equal(t, "Intro text.", sections[0].Body)
	assert.Equal(t, "", sections[0].Heading)
	assert.Equal(t, 5, sections[0].Line)

	assert.Equal(t, "Acme", sections[1].Heading)
	assert.Equal(t, 1, sections[1].Level)
	assert.Equal(t, "Acme", sections[1].Breadcrumb)
	assert.Equal(t, 7, sections[1].Line)

	assert.Equal(t, "Install", sections[2].Heading)
	assert.Equal(t, "Acme > Install", sections[2].Breadcrumb)
	assert.Contains(t, sections[2].Body, "# not a heading")
	assert.Contains(t, sections[2].Body, "Still install.")

	assert.Equal(t, "Deploy", sections[3].Heading)
	assert.Equal(t, 2, sections[3].Level)
	assert.Equal(t, "Acme > Deploy", sections[3].Breadcrumb)
	assert.Equal(t, "Run it.", sections[3].Body)

	assert.Equal(t, "After", sections[4].Heading)
	assert.Equal(t, 4, sections[4].Ordinal)
}

// A file with no heading is one section named after the file.
func TestSplit_NoHeading(t *testing.T) {
	t.Parallel()
	title, sections := search.Split([]byte("Just words.\n"), "notes.md")
	assert.Equal(t, "notes", title)
	assert.Len(t, sections, 1)
	assert.Equal(t, "notes", sections[0].Heading)
	assert.Equal(t, "Just words.", sections[0].Body)
}

// A text file is one section named by its file.
func TestSplitText(t *testing.T) {
	t.Parallel()
	title, sections := search.SplitText([]byte("plain text"), "todo.txt")
	assert.Equal(t, "todo", title)
	assert.Len(t, sections, 1)
	assert.Equal(t, "plain text", sections[0].Body)

	title, sections = search.SplitText(nil, "empty.txt")
	assert.Equal(t, "empty", title)
	assert.Len(t, sections, 0)
}

// A section keeps its place in the document.
func TestSection(t *testing.T) {
	t.Parallel()
	section := search.Section{Ordinal: 2, Level: 2, Heading: "Install", Breadcrumb: "Acme > Install", Line: 9, Body: "Run it."}
	assert.Equal(t, 2, section.Ordinal)
	assert.Equal(t, "Acme > Install", section.Breadcrumb)
}
