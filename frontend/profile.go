package frontend

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/storage"
)

// profileCount is how many of each kind of activity a profile lists.
const profileCount = 10

// profile shows a user to read: who they are and what they did lately, as far as the one who looks sees it.
// Edit, for your own page or with user/edit, opens the forms of the roles, folders, permissions, keys and MCP access.
func (h *Handlers) profile(w http.ResponseWriter, r *http.Request) {
	user, own, ok := h.userPage(w, r)
	if !ok {
		return
	}
	p := h.page(r, user.Name)
	p.Crumbs = view.Crumbs(user.Name)
	if p.Can("user/view") {
		p.Crumbs = view.Crumbs("Settings", "/settings", "Users", "/settings/users", user.Name)
	}
	p.Account = user
	p.Profile = &view.Profile{Own: own, CanEdit: own || p.Can("user/edit")}
	if own || p.Can("key/view") {
		p.Keys = h.userKeys(r.Context(), user)
	}
	h.activityOf(r, p, user)
	h.render(w, http.StatusOK, "profile", p)
}

// userPage returns the user a page under /users/{name} shows and whether the page is the signed in user's own.
// Another user's page takes user/view, and the page answers forbidden itself.
// A name no user holds answers not found.
func (h *Handlers) userPage(w http.ResponseWriter, r *http.Request) (*model.User, bool, bool) {
	me := access.UserFrom(r.Context())
	name := param(r, "name")
	own := me != nil && me.Name == name
	if !own && !h.require(w, r, "user/view") {
		return nil, own, false
	}
	user, err := h.opts.Users.GetByName(r.Context(), name)
	if err != nil {
		h.missing(w, r, "No user by that name.")
		return nil, own, false
	}
	return user, own, true
}

// activityOf fills a profile with the latest runs, workflow runs, notes
// and prompts of a user that the page's user may see: runs and
// workflow runs in folders they see, notes with note/view, prompts with
// prompt/view in places they see.
func (h *Handlers) activityOf(r *http.Request, p *view.Page, user *model.User) {
	ctx, viewer, pr := r.Context(), p.User, p.Profile
	pr.Runs = h.visibleRuns(ctx, h.runSeer(r), storage.RunFilter{UserName: user.Name, Folders: viewer.FolderList(), Limit: profileCount})
	if h.opts.WorkflowRuns != nil {
		runs, _ := h.opts.WorkflowRuns.List(ctx, 50*profileCount)
		for _, run := range runs {
			if len(pr.FlowRuns) == profileCount {
				break
			}
			if run.UserName == user.Name && h.seesFlowRun(ctx, viewer, run.ID, 0) {
				pr.FlowRuns = append(pr.FlowRuns, run)
			}
		}
	}
	if h.opts.Board == nil {
		return
	}
	if p.Can("note/view") {
		pr.Notes, _ = h.opts.Board.Notes.List(ctx, storage.NoteFilter{Channel: model.ChannelNotes, UserName: user.Name, Limit: profileCount})
	}
	if p.Can("prompt/view") {
		prompts, _ := h.opts.Board.Notes.List(ctx, storage.NoteFilter{Prompts: true, UserName: user.Name, Limit: 5 * profileCount})
		for _, prompt := range prompts {
			if len(pr.Prompts) < profileCount && h.seesPlace(viewer, prompt.Dir) {
				pr.Prompts = append(pr.Prompts, prompt)
			}
		}
	}
}

// seesPlace reports whether a user sees the place a prompt ran in, as a
// note's Dir holds it: the root, internal, or a directory of a folder they see,
// the deepest folder that holds it.
func (h *Handlers) seesPlace(user *model.User, place string) bool {
	if place == model.Default {
		return true
	}
	best := ""
	for _, f := range h.opts.Catalogue.List() {
		if (place == f.ID || strings.HasPrefix(place, f.ID+"/")) && len(f.ID) > len(best) {
			best = f.ID
		}
	}
	return best != "" && user.CanSee(best)
}

// userMatrix returns a user's policy with what they may do. It changes
// with user/edit, never an admin's, and your own only with user/edit
// through your role, as the API takes it.
func (h *Handlers) userMatrix(r *http.Request, p *view.Page, user *model.User, own bool) *view.Matrix {
	set := h.opts.Users.Permissions(r.Context(), user)
	m := view.NewMatrix(config.Default(), set.Own, set.Effective(), user.CanAdmin())
	m.ID, m.Title, m.API, m.Effective = "user-"+user.Name, user.Name, "/api/permissions/users/"+user.Name, true
	m.Editable = !m.Fixed && p.Can("user/edit") && (!own || editsByRole(p))
	return &m
}

// editsByRole reports whether the page's user holds user/edit through
// their role; a policy of their own alone reports false.
func editsByRole(p *view.Page) bool {
	return p.User.CanAdmin() || p.Permissions.Role.Effect("user/edit") == permission.Allow
}
