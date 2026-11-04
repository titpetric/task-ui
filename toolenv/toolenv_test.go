package toolenv

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The tooling reads as TASK_UI_TOOLING, one a line with the name and
// what the tool is used for.
func TestEnv(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", Env(nil))
	assert.Equal(t, "git: version control\ngo: Go toolchain", Env([]Tool{
		{Name: "git", Description: "version control"},
		{Name: "go", Description: "Go toolchain"},
	}))
}

// A spec is found by the URL-safe ID and by the name the page shows.
func TestSpecByID(t *testing.T) {
	t.Parallel()
	spec, ok := SpecByID("docker-compose")
	assert.True(t, ok)
	assert.Equal(t, "docker compose", spec.Name)

	spec, ok = SpecByName("docker compose")
	assert.True(t, ok)
	assert.Equal(t, "docker-compose", spec.ID)

	_, ok = SpecByID("nope")
	assert.False(t, ok)
}

// The version line drops the tool's name and the word version.
func TestVersionLine(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "go1.22 linux/amd64", versionLine("go version go1.22 linux/amd64\n", "go"))
	assert.Equal(t, "2.43.0", versionLine("  git version 2.43.0  \n", "git"))
	assert.Equal(t, "jq-1.7.1", versionLine("jq-1.7.1\n", "jq"))
	assert.Equal(t, "", versionLine("\n", "make"))
}
