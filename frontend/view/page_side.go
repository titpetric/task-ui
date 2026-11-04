package view

import (
	"github.com/titpetric/task-ui/model"
)

// FillSide fills the sidebar's top block from the rows a table holds,
// for the user the page is rendered to. The handler reads the rows and
// calls this, so the rule of what a user sees lives in one place.
func (p *Page) FillSide(rows []model.Sidebar) {
	p.SideItems = SideItems(rows, p.Can, p.CanSettings(), p.SideActive)
}

// SideActive reports whether a sidebar row is the page being read: the
// dashboard on its own path alone, every other row on anything under it,
// so /settings/keys lights Settings and / does not light everything.
func (p *Page) SideActive(href string) bool {
	if href == "/" {
		return p.IsPath("/")
	}
	return p.Under(href)
}
