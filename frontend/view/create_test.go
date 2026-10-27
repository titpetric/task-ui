package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// Add project names the workspace to clone into, escaped; without one
// the form chooses.
func TestAddProjectHref(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/settings/projects/new", view.AddProjectHref(""))
	assert.Equal(t, "/settings/projects/new?workspace=malta", view.AddProjectHref("malta"))
	assert.Equal(t, "/settings/projects/new?workspace=internal", view.AddProjectHref("internal"))
	assert.Equal(t, "/settings/projects/new?workspace=a+b%26c", view.AddProjectHref("a b&c"))
}

// The pages a workspace owns hang off its settings page.
func TestWorkspaceHrefs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/settings/workspaces/malta", view.WorkspaceHref("malta"))
	assert.Equal(t, "/settings/workspaces/malta/keys", view.WorkspaceKeysHref("malta"))
	assert.Equal(t, "/settings/secrets?workspace=malta", view.WorkspaceSecretsHref("malta"))
}

// A key of a user starts on the key forms with the user preset.
func TestUserKeyHref(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/settings/keys/generate?for=ops", view.UserKeyHref("generate", "ops"))
	assert.Equal(t, "/settings/keys/upload?for=ops", view.UserKeyHref("upload", "ops"))
	assert.Equal(t, "/settings/keys/generate?for=a+b", view.UserKeyHref("generate", "a b"))
}

func createLabels(p *view.Page) []string {
	var labels []string
	for _, item := range p.CreateItems() {
		labels = append(labels, item.Label)
	}
	return labels
}

// An admin may add everything; a user adding users needs the server to
// write a password file.
func TestPage_CreateItems_Admin(t *testing.T) {
	t.Parallel()

	p := &view.Page{}
	signIn(p, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	assert.Equal(t, []string{"Add workspace", "Add project", "Generate key", "Upload key", "Add variable", "Add workflow"}, createLabels(p))

	p.CanAddUsers = true
	assert.Equal(t, []string{"Add workspace", "Add project", "Add user", "Generate key", "Upload key", "Add variable", "Add workflow"}, createLabels(p))

	items := p.CreateItems()
	assert.Equal(t, "/settings/workspaces/new", items[0].Href)
	assert.Equal(t, "/settings/projects/new", items[1].Href)
	assert.Equal(t, "/settings/users/new", items[2].Href)
	assert.Equal(t, "/settings/keys/generate", items[3].Href)
	assert.Equal(t, "/settings/keys/upload", items[4].Href)
	assert.Equal(t, "/settings/secrets/new", items[5].Href)
	assert.Equal(t, "/workflows/new", items[6].Href)
}

// Each item takes the rule of the page it links.
func TestPage_CreateItems_Rules(t *testing.T) {
	t.Parallel()

	p := &view.Page{CanAddUsers: true}
	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	assert.Equal(t, 0, len(p.CreateItems()))

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"key/edit": permission.Allow, "project/edit": permission.Allow})
	assert.Equal(t, []string{"Add project", "Generate key", "Upload key"}, createLabels(p))

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"user/edit": permission.Allow})
	assert.Equal(t, []string{"Add user"}, createLabels(p))
	p.CanAddUsers = false
	assert.Equal(t, 0, len(p.CreateItems()), "no password file, no Add user")
}
