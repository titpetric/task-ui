package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// The prompt box of a folder posts its ID; of a directory inside it, the
// directory too.
func TestPromptIn(t *testing.T) {
	t.Parallel()
	assert.Equal(t, &view.Prompt{API: "/api/prompts", Folder: "malta", Where: "malta/", Place: "malta"}, view.PromptIn("malta", ""))
	assert.Equal(t, &view.Prompt{API: "/api/prompts", Folder: "malta", Dir: "dockerfiles/task", Where: "malta/dockerfiles/task/", Place: "malta/dockerfiles/task"}, view.PromptIn("malta", "dockerfiles/task"))
}

// The prompt box of a run's page posts the run, and stands open.
func TestPromptAbout(t *testing.T) {
	t.Parallel()
	assert.Equal(t, &view.Prompt{API: "/api/prompts", Run: "01RUN", Where: "malta/dockerfiles/", Place: "malta/dockerfiles", Open: true}, view.PromptAbout("01RUN", "malta/dockerfiles"))
}

// The Workflows page's box asks in the area, in the root unless its
// form picks a folder, open.
func TestPromptOfArea(t *testing.T) {
	t.Parallel()
	p := view.PromptOfArea("workflows")
	assert.Equal(t, "workflows", p.Area)
	assert.Equal(t, "/api/prompts", p.API)
	assert.Equal(t, "_", p.Place)
	assert.True(t, p.Open)
}

// A box says who the prompt runs as: an admin picks a user, the login
// selected, which the form posts as user; anyone else reads "as you".
// The header's prompt gets the same pick.
func TestPrompt_RunsAs(t *testing.T) {
	t.Parallel()
	users := []model.User{{Name: "admin", Role: model.RoleAdmin}, {Name: "ops", Role: model.RoleRunner}}
	const pick = `<select name="user" aria-label="Runs as">`

	p := admin(t)
	p.Prompt = view.PromptIn("infra", "")
	p.PromptUsers = users
	p.UserKeys = map[string][]string{"ops": {"deploy"}}
	body := render(t, "folder", p)
	assert.Contains(t, body, `Runs @claude-prompt as `+pick)
	assert.Contains(t, body, `<option value="admin" selected data-keys="">admin · no keys</option>`)
	assert.Contains(t, body, `<option value="ops" data-keys="deploy">ops · keys: deploy</option>`)
	assert.Contains(t, body, `</select> in infra/, one prompt at a time.`)
	assert.Contains(t, body, `<select class="header-prompt-user" id="header-prompt-user" tabindex="-1" aria-label="Runs as">`)
	assert.NotContains(t, body, "as you in")

	p = admin(t)
	p.Prompt = view.PromptAbout(p.Run.ID, "infra")
	p.PromptUsers = users
	body = render(t, "run", p)
	assert.Contains(t, body, `Runs @claude-prompt as `+pick)
	assert.Contains(t, body, `</select> in infra/, with this run's output.`)

	p = admin(t)
	p.Prompt = view.PromptIn("infra", "")
	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"prompt/edit": permission.Allow})
	body = render(t, "folder", p)
	assert.Contains(t, body, `Runs @claude-prompt as you in infra/, one prompt at a time.`)
	assert.Contains(t, body, `id="header-prompt"`)
	assert.NotContains(t, body, pick)
	assert.NotContains(t, body, "header-prompt-user")
}
