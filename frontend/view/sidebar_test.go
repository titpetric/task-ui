package view_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// With no pins the sidebar lists the folders seen and every directory
// right inside a workspace, jobs or none; a home project, a project set
// not to show and the folders inside them stay out.
func TestSideDefaults(t *testing.T) {
	t.Parallel()
	folders := []*folder.Folder{jobs("malta", "deploy"), jobs("malta/dockerfiles", "build"), jobs("malta/workspace", "git:clone"), jobs("malta/workspace/sub", "x"), jobs("apps", "build")}
	dirs := []string{"malta/apps", "malta/dockerfiles", "malta/vhosts", "malta/workspace", "malta/secret"}
	hidden := map[string]bool{"malta/workspace": true, "apps": true}
	sees := func(id string) bool { return id != "malta/secret" }
	side := view.SideDefaults(folders, dirs, hidden, sees)
	assert.Equal(t, []string{"malta", "malta/dockerfiles", "malta/apps", "malta/vhosts"}, idsOf(side))
	assert.Equal(t, "apps", side[2].Name())
	assert.Len(t, side[2].Jobs, 0)

	// The home project never makes a workspace bar or a project of its own.
	groups := view.Groups(view.NewSidebar(side, nil, map[string]int{"malta": 0}, nil), map[string]int{"malta": 0})
	assert.Len(t, groups, 1)
	assert.True(t, groups[0].Workspace)
	assert.Equal(t, []string{"malta/apps", "malta/dockerfiles", "malta/vhosts"}, groupIDs(groups[0].Projects))
}

func groupIDs(groups []view.Group) []string {
	ids := []string{}
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	return ids
}

// A pin adds to the sidebar in its natural place: a deep folder with
// its workspace and project above it, in the catalogue's order. A job
// pin places its folder with no jobs on it, the catalogue's own folder
// untouched. A pin of a folder the sidebar lists already adds nothing,
// and a pin in a home project stays out.
func TestSidebar(t *testing.T) {
	t.Parallel()
	ws := map[string]int{"malta": 0}
	defaults := []*folder.Folder{jobs("apps", "build"), jobs("malta/dockerfiles", "default", "build")}
	blog := &folder.Folder{ID: "malta/apps/blog"}
	edge := jobs("infra/edge", "plan", "apply")
	docker := defaults[1]
	pin := func(f *folder.Folder, job, dir, args string) view.SidePin {
		p := view.SidePin{Pin: model.Pin{ID: f.ID + job + dir, Folder: f.ID, Job: job, Dir: dir, Args: args}, Folder: f}
		if job != "" {
			p.Job = f.Job(job)
		}
		return p
	}
	pins := []view.SidePin{
		pin(blog, "", "", ""),
		pin(edge, "plan", "", "[]"),
		pin(docker, "build", "", "[]"),
		pin(docker, "build", "", `["--push"]`),
		pin(&folder.Folder{ID: "malta/workspace"}, "", "", ""),
		pin(defaults[0], "", "", ""),
	}
	side := view.NewSidebar(defaults, pins, ws, map[string]bool{"malta/workspace": true})
	assert.Equal(t, []string{"apps", "infra", "infra/edge", "malta", "malta/apps", "malta/apps/blog", "malta/dockerfiles"}, idsOf(side))

	// infra/edge is there for its pinned plan alone.
	var listedEdge *folder.Folder
	for _, f := range side {
		if f.ID == "infra/edge" {
			listedEdge = f
		}
	}
	assert.Len(t, listedEdge.Jobs, 0)
	assert.Len(t, edge.Jobs, 2, "the catalogue's folder stays as it is")

	groups := view.Groups(side, ws)
	assert.Equal(t, []string{"apps", "infra", "malta"}, groupIDs(groups))
	assert.Equal(t, []string{"malta/apps", "malta/dockerfiles"}, groupIDs(groups[2].Projects))
	assert.Equal(t, []string{"malta/apps/blog"}, idsOf(groups[2].Projects[0].Nested))
}

// The sidebar is workspaces and projects: a project is one row and a
// link, with no folder under it, no job and no pin. A pin puts the
// project of the folder it names in the tree and draws nothing itself;
// the project's page is where it is highlighted.
func TestView_RenderSideRows(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := admin(t)
	p.Running = nil
	edge := jobs("infra/edge", "plan")
	plan := edge.Jobs[0]
	plan.Dir = "eu"
	p.SideFolders = view.NewSidebar(nil, []view.SidePin{{Pin: model.Pin{ID: "p1", Folder: "infra/edge", Job: "plan", Dir: "eu", Args: `["-x"]`}, Folder: edge, Job: plan}}, nil, nil)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", p))
	out := buf.String()
	side := out[strings.Index(out, `<nav class="folders">`):]
	side = side[:strings.Index(side, "</nav>")]
	assert.Contains(t, side, `<div class="project active" data-project="infra">`)
	assert.Contains(t, side, `<a class="section active" href="/f/internal/infra" data-folder="infra" title="infra/">`)
	for _, gone := range []string{"data-job-key", `data-folder="infra/edge"`, `class="jobs`, `<details class="folder`, "Nothing to run."} {
		assert.NotContains(t, side, gone)
	}
}

// jobs returns a folder of the ID with a job per name.
func jobs(id string, names ...string) *folder.Folder {
	f := &folder.Folder{ID: id}
	for _, name := range names {
		job := &folder.Job{Folder: id, ID: name}
		job.Name = name
		f.Jobs = append(f.Jobs, job)
	}
	return f
}

// sideMenu returns the top nav of the sidebar.
func sideMenu(t *testing.T, body string) string {
	t.Helper()
	at := strings.Index(body, `<nav class="menu">`)
	assert.True(t, at >= 0, "the sidebar menu")
	rest := body[at:]
	return rest[:strings.Index(rest, "</nav>")]
}

// Settings is in the sidebar, after Threads, for a user who may open any
// settings page; the link is active on every page under Settings.
func TestSidebar_Settings(t *testing.T) {
	t.Parallel()

	p := admin(t)
	menu := sideMenu(t, render(t, "home", p))
	threads, settings := strings.Index(menu, `href="/threads"`), strings.Index(menu, `<a href="/settings" class="">`)
	assert.True(t, threads > 0 && settings > threads, "Settings after Threads")

	p.Path = "/settings/keys"
	assert.Contains(t, sideMenu(t, render(t, "home", p)), `<a href="/settings" class="active">`)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	assert.NotContains(t, sideMenu(t, render(t, "home", p)), `href="/settings"`, "nothing to open, no entry")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow})
	assert.Contains(t, sideMenu(t, render(t, "home", p)), `href="/settings"`)
}

// With no folder the sidebar says so and links the form that adds a
// repository, for a user who may; others read where to add one.
func TestSidebar_EmptyLinksAddProject(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.SideFolders, p.Folders = nil, nil
	body := render(t, "home", p)
	assert.Contains(t, body, `No folders yet. <a href="/settings/projects/new">Add a git repository</a>, or put a folder with jobs under the root.`)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	body = render(t, "home", p)
	assert.Contains(t, body, `No folders yet. Add a git repository in Settings, or put a folder with jobs under the root.`)
	assert.NotContains(t, body, `href="/settings/projects/new"`)
}
