package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The chrome every page shares: the user, the folders they see, and the
// recent runs with who started them.
func TestPage_Chrome(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.run(t, "infra", "plan", "ops", model.StateOK)
	f.run(t, "private", "deploy", "admin", model.StateFailed)

	body := f.page(t, asAdmin, "/")
	assert.Contains(t, body, `class="avatar">AD<`)
	assert.Contains(t, body, "Settings")
	assert.Contains(t, body, "Logout")
	assert.Contains(t, body, `data-folder="infra"`)
	assert.Contains(t, body, `data-folder="private"`)
	assert.Contains(t, body, "Recents")
	assert.Contains(t, body, `class="avatar sm" title="ops">OP<`)
	assert.Contains(t, body, `class="avatar sm" title="admin">AD<`)

	limited := f.page(t, asLimited, "/")
	assert.NotContains(t, limited, "Settings</a>")
	assert.NotContains(t, limited, `data-folder="private"`)
	assert.NotContains(t, limited, `title="admin">AD<`)
	assert.Contains(t, limited, `title="ops">OP<`)
}

func mustProject(t *testing.T, f *fixture, folder string) *model.Project {
	t.Helper()
	project, err := f.opts.Projects.GetByFolder(context.Background(), folder)
	assert.NoError(t, err)
	return project
}

// A project in the sidebar folds: a disclosure with the folders inside
// it, open by default, the chevron on its row.
func TestPage_ProjectsFold(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/")
	infra := strings.Index(body, `<details class="project" data-project="infra" open>`)
	edge := strings.Index(body, `<details class="folder" data-folder="infra/edge"`)
	private := strings.Index(body, `<details class="project" data-project="private" open>`)
	assert.True(t, infra >= 0 && edge > infra && private > edge, "infra holds infra/edge, then private")
	assert.True(t, strings.Index(body[infra:], "</details>") > edge-infra, "infra closes after infra/edge")
	chev := body[infra : strings.Index(body[infra:], "</summary>")+infra]
	assert.Contains(t, chev, `<a class="section" href="/f/infra" data-folder="infra"`)
	assert.Contains(t, chev, `<svg class="ic chev"><use href="#i-chev"/></svg>`)

	// The sidebar leaves out a project set not to show there, with the
	// folders inside it; its pages stay.
	assert.NoError(t, f.opts.Projects.Create(context.Background(), &model.Project{Folder: "infra", SidebarHidden: true}))
	hidden := f.page(t, asAdmin, "/")
	assert.NotContains(t, hidden, `data-project="infra"`)
	assert.NotContains(t, hidden, `<details class="folder" data-folder="infra/edge"`)
	assert.Contains(t, hidden, `data-project="private"`)
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/f/infra").Status)
	assert.NoError(t, f.opts.Projects.Delete(context.Background(), mustProject(t, f, "infra").ID))

	// The sidebar marks the open project, on the project and its whole row.
	page := f.page(t, asAdmin, "/f/infra")
	assert.Contains(t, page, `<details class="project active" data-project="infra" open>`)
	assert.Contains(t, page, `<summary class="active" title="infra/`)
	assert.Contains(t, page, `<a class="section active" href="/f/infra"`)
	// A folder inside it keeps the project open, and lights its own row only.
	edgePage := f.page(t, asAdmin, "/f/infra/edge")
	assert.Contains(t, edgePage, `<details class="project active" data-project="infra" open>`)
	assert.NotContains(t, edgePage, `<summary class="active"`)
}

func TestPage_NoticeAndError(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/runs?notice=Saved&error=Oops")
	assert.Contains(t, body, `<div class="alert ok" role="status">Saved</div>`)
	assert.Contains(t, body, `<div class="alert bad" role="alert">Oops</div>`)
}

func TestPage_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, path := range []string{"/settings", "/settings/projects", "/settings/keys", "/settings/secrets", "/settings/users", "/f/infra/env/.env"} {
		res := f.get(t, asOps, path)
		assert.Equal(t, http.StatusForbidden, res.Status, path)
		assert.Contains(t, res.Body, "Settings are for admins.")
	}
}
