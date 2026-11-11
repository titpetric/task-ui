package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// SidebarRequest is the Sidebar settings page's form: the links in the
// order the table listed them, and whether each one is shown.
//
// The order is the array's order and no row carries a number, the way a
// workflow's steps post: a number per row and an array both saying the
// same thing is a pair that can disagree.
//
// Shown and not hidden, because the switch on the page reads "Shown". A
// link the map leaves out keeps whatever it had: the page names every
// row, and a caller that names one row changes one row rather than
// hiding the rest by saying nothing about them.
type SidebarRequest struct {
	Items []string        `json:"items"`
	Shown map[string]bool `json:"shown"`
}

// saveSidebar writes the order of the sidebar and what it leaves out, for
// admins alone, as the page's form sends it. A link the table does not
// hold is left alone rather than added. It answers the rows as they
// stand.
func (h *Handlers) saveSidebar(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin orders the sidebar")
		return
	}
	if h.opts.Sidebar == nil {
		fail(w, http.StatusNotFound, "this instance keeps no sidebar")
		return
	}
	var req SidebarRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Items) == 0 {
		fail(w, http.StatusBadRequest, "the sidebar takes the links in the order to list them")
		return
	}
	rows, err := h.opts.Sidebar.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	// The switch says what is shown; the table stores what is hidden. A
	// row the request did not name keeps the flag it had, so the rows
	// start from what the table holds and the request overlays it.
	hidden := make(map[string]bool, len(rows))
	for _, row := range rows {
		hidden[row.Href] = row.Hidden
	}
	for href, shown := range req.Shown {
		hidden[href] = !shown
	}
	if err := h.opts.Sidebar.Order(r.Context(), req.Items, hidden); err != nil {
		failErr(w, err)
		return
	}
	rows, err = h.opts.Sidebar.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sidebar": sidebarViews(rows)})
}

// SidebarView is one row of the sidebar as the API answers it.
type SidebarView struct {
	Kind     string `json:"kind"`
	Href     string `json:"href"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	Position int64  `json:"position"`
	Hidden   bool   `json:"hidden"`
	Rule     string `json:"rule,omitempty"`
}

// sidebarViews is the rows as the API answers them. The identifier and
// the dates are left out: a row is addressed by its link, which is what
// the page posts and what the table's unique index holds.
func sidebarViews(rows []model.Sidebar) []SidebarView {
	out := make([]SidebarView, 0, len(rows))
	for _, row := range rows {
		out = append(out, SidebarView{
			Kind: row.Kind, Href: row.Href, Title: row.Title, Icon: row.Icon,
			Position: row.Position, Hidden: row.Hidden, Rule: row.Rule,
		})
	}
	return out
}

// listSidebar answers the sidebar's rows, in its order, for a user who
// may see a step of the settings.
func (h *Handlers) listSidebar(w http.ResponseWriter, r *http.Request) {
	if h.opts.Sidebar == nil {
		writeJSON(w, http.StatusOK, map[string]any{"sidebar": []SidebarView{}})
		return
	}
	rows, err := h.opts.Sidebar.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sidebar": sidebarViews(rows)})
}
