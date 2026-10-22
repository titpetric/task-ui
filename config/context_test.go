package config_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/mcp"
	"github.com/titpetric/task-ui/model"
)

// context.md keeps in step with task-ui: a heading names every section
// and permission of permissions.yml, and the text names every MCP tool
// and every built-in job. A feature added without a line there fails
// here.
func TestContext(t *testing.T) {
	t.Parallel()
	text := config.Context()
	assert.True(t, strings.HasPrefix(text, "# task-ui\n"))

	p := config.Default()
	for _, entry := range append(append([]config.Entry{}, p.Sections...), p.Permissions...) {
		heading := regexp.MustCompile("(?m)^## .*\\(.*`" + entry.Key + "`.*\\)$")
		assert.True(t, heading.MatchString(text), "a heading names "+entry.Key)
		assert.NotEqual(t, "", config.About(entry.Key), entry.Key)
	}
	for _, tool := range mcp.New(nil, "test").Tools() {
		assert.Contains(t, text, "`"+tool.Name+"`")
	}
	for _, job := range []string{folder.ShellID, model.ClaudeJob, model.PromptJob, model.NoteJob} {
		assert.Contains(t, text, "`"+job+"`")
	}
}

// About is the first sentence of the section a key's heading names.
func TestAbout(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Artifacts are the files runs leave.", config.About("artifact"))
	assert.Equal(t, config.About("inventory"), config.About("hosts"))
	assert.Equal(t, "", config.About("nothing"))
}
