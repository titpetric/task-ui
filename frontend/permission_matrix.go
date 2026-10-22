package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// permissionMatrix shows each role's policy, to change with user/edit:
// the admin role's never changes, and your own role's changes only with
// user/edit through it. Takes user/view.
func (h *Handlers) permissionMatrix(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
		return
	}
	p := h.page(r, "Permissions")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Permissions")
	p.Step = view.StepPermissions
	for _, role := range model.Roles() {
		fixed := role == model.RoleAdmin
		policy := permission.Policy{}
		if !fixed {
			policy, _ = h.opts.Users.Policies().Policy(r.Context(), permission.KindRole, role)
		}
		effective := permission.Set{User: &model.User{Role: role}, Role: policy}.Effective()
		m := view.NewMatrix(config.Default(), policy, effective, fixed)
		m.ID, m.Title, m.API = "role-"+role, role, "/api/permissions/roles/"+role
		m.Editable = !fixed && p.Can("user/edit") && (role != p.User.Role || editsByRole(p))
		p.Matrices = append(p.Matrices, m)
	}
	h.render(w, http.StatusOK, "permissions", p)
}
