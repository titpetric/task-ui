package skills_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// lintAnswer is what atkins prints for a pipeline its linter refuses.
const lintAnswer = `{
  "error": "pipeline 'Broken' has errors",
  "lint": [
    {"pipeline": "Broken", "path": "/w/app/atkins.yml", "job": "build",
     "issue": "missing task reference", "detail": "step references task 'go:fmt', but job \"go:fmt\" not found"}
  ]
}`

// A normal listing reads into its sections, and is not broken.
func TestParseListing(t *testing.T) {
	t.Parallel()
	listing, err := skills.ParseListing([]byte(`[{"id":"","kind":"pipeline","cmds":[{"id":"build","name":"build"}]}]`))
	assert.NoError(t, err)
	assert.Len(t, listing.Sections, 1)
	assert.Equal(t, "build", listing.Sections[0].Cmds[0].Name)
	assert.False(t, listing.Broken())
}

// The lint object reads into the error and its findings, whatever
// follows it; sections after it list too.
func TestParseListing_Lint(t *testing.T) {
	t.Parallel()
	listing, err := skills.ParseListing([]byte(lintAnswer + "\nError: EOF\n"))
	assert.NoError(t, err)
	assert.True(t, listing.Broken())
	assert.Equal(t, "pipeline 'Broken' has errors", listing.Error)
	assert.Len(t, listing.Lint, 1)
	assert.Equal(t, skills.Finding{Pipeline: "Broken", Path: "/w/app/atkins.yml", Job: "build", Issue: "missing task reference", Detail: `step references task 'go:fmt', but job "go:fmt" not found`}, listing.Lint[0])
	assert.Nil(t, listing.Sections)

	listing, err = skills.ParseListing([]byte(lintAnswer + "\n" + `[{"kind":"pipeline","cmds":[{"name":"test"}]}]` + "\n" + `{"error":"second","lint":[{"detail":"more"}]}`))
	assert.NoError(t, err)
	assert.Len(t, listing.Sections, 1)
	assert.Len(t, listing.Lint, 2)
	assert.Equal(t, "pipeline 'Broken' has errors", listing.Error)
}

// Garbage, nothing, and an object that is neither are errors.
func TestParseListing_Garbage(t *testing.T) {
	t.Parallel()
	for _, out := range []string{"", "not json", `{"id":"x"}`, `[{"cmds":"no"}]`, "Error: EOF"} {
		_, err := skills.ParseListing([]byte(out))
		assert.Error(t, err, out)
	}
}
