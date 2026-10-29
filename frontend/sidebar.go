package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// sidebarSettings shows the sidebar as a table: a row per item, in the
// order the sidebar lists them, with a grip to drag it, the link it
// opens and a switch that leaves it out.
//
// Every row is here, the hidden ones too, because this is the page that
// unhides one. The sidebar beside the table shows the same rows less the
// hidden, so the two disagree until Save, which is the point: the table
// is what the sidebar would be.
func (h *Handlers) sidebarSettings(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, view.StepSidebar)
	if !p.CanSettings() {
		h.refuse(w, r)
		return
	}
	p.Crumbs = view.Crumbs("Settings", "/settings", view.StepSidebar)
	p.Step = view.StepSidebar

	items := model.SidebarDefaults()
	if h.opts.Sidebar != nil {
		stored, err := h.opts.Sidebar.List(r.Context())
		if err != nil {
			http.Error(w, "reading the sidebar failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(stored) > 0 {
			items = stored
		}
	}
	var modules []model.Module
	if h.opts.Modules != nil {
		for _, found := range h.opts.Modules.Modules() {
			modules = append(modules, model.Module{
				Name: found.Name(), Title: found.Label(), Description: found.Description, Icon: found.Icon,
			})
		}
	}
	p.SidebarRows = view.SidebarRows(items, modules)
	h.render(w, http.StatusOK, "sidebar", p)
}
