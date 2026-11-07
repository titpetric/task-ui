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
	groups := view.Groups(view.NewSidebar(side, nil, map[string]int{"malta": 0}, nil).Folders, map[string]int{"malta": 0})
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
// its workspace and project above it, in the catalogue's order, and a
// job pin under its folder. Every job pin is listed, the folder's row
// having nothing else under it. A pin of a folder the sidebar lists
// already adds nothing, and a pin in a home project stays out.
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
	s := view.NewSidebar(defaults, pins, ws, map[string]bool{"malta/workspace": true})
	assert.Equal(t, []string{"apps", "infra", "infra/edge", "malta", "malta/apps", "malta/apps/blog", "malta/dockerfiles"}, idsOf(s.Folders))

	// infra/edge is there for its pinned plan alone.
	var listedEdge *folder.Folder
	for _, f := range s.Folders {
		if f.ID == "infra/edge" {
			listedEdge = f
		}
	}
	assert.Len(t, listedEdge.Jobs, 0)
	assert.Len(t, edge.Jobs, 2, "the catalogue's folder stays as it is")
	assert.Len(t, s.Pins["infra/edge"], 1)
	assert.Equal(t, "plan", s.Pins["infra/edge"][0].Label())

	// Both pins of build are listed, the bare one and the one with args.
	assert.Len(t, s.Pins["malta/dockerfiles"], 2)
	assert.Equal(t, "build", s.Pins["malta/dockerfiles"][0].Label())
	assert.Equal(t, "build --push", s.Pins["malta/dockerfiles"][1].Label())

	groups := view.Groups(s.Folders, ws)
	assert.Equal(t, []string{"apps", "infra", "malta"}, groupIDs(groups))
	assert.Equal(t, []string{"malta/apps", "malta/dockerfiles"}, groupIDs(groups[2].Projects))
	assert.Equal(t, []string{"malta/apps/blog"}, idsOf(groups[2].Projects[0].Nested))
}

// The sidebar draws a pinned job as a run item under its folder: its
// play runs the job with the pin's directory and args.
func TestView_RenderSidePin(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := admin(t)
	p.Running = nil
	edge := jobs("infra/edge", "plan")
	plan := edge.Jobs[0]
	plan.Dir = "eu"
	s := view.NewSidebar(nil, []view.SidePin{{Pin: model.Pin{ID: "p1", Folder: "infra/edge", Job: "plan", Dir: "eu", Args: `["-x"]`}, Folder: edge, Job: plan}}, nil, nil)
	p.SideFolders, p.SidePins = s.Folders, s.Pins
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", p))
	out := buf.String()
	side := out[strings.Index(out, `<nav class="folders">`):]
	side = side[:strings.Index(side, "</nav>")]
	assert.Contains(t, side, `data-project="infra"`)
	assert.Contains(t, side, `<div class="job pinned" data-job-key="infra/edge/eu/plan">`)
	assert.Contains(t, side, `data-run data-folder="infra/edge" data-dir="eu" data-job="plan" data-run-args="[&#34;-x&#34;]"`)
	assert.Contains(t, side, `>eu/ plan -x</a>`)
	assert.NotContains(t, side, "Nothing to run.")
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

// Settings is in the sidebar, after Notes, for a user who may open any
// settings page; the link is active on every page under Settings.
func TestSidebar_Settings(t *testing.T) {
	t.Parallel()

	p := admin(t)
	menu := sideMenu(t, render(t, "home", p))
	notes, settings := strings.Index(menu, `href="/notes"`), strings.Index(menu, `<a href="/settings" class="">`)
	assert.True(t, notes > 0 && settings > notes, "Settings after Notes")

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
