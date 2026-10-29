package view

import (
	"github.com/titpetric/task-ui/model"
)

// SidebarRow is one row of the Sidebar settings table: the item as the
// table holds it, with the module behind it when there is one.
type SidebarRow struct {
	Item model.Sidebar

	// Module is the module a row of that kind links, for the description
	// the table shows beside the title. Nil for a view, and nil for a
	// module whose row outlived its table entry.
	Module *model.Module
}

// Title is what the row shows, falling back to the link so a row with no
// title is still something to drag.
func (r SidebarRow) Title() string {
	if r.Item.Title != "" {
		return r.Item.Title
	}
	return r.Item.Href
}

// Note is the line under the title: a module's description, or what the
// instance calls the row's kind.
func (r SidebarRow) Note() string {
	if r.Module != nil && r.Module.Description != "" {
		return r.Module.Description
	}
	if r.Item.IsModule() {
		return "a module"
	}
	return ""
}

// SidebarRows pairs the sidebar's rows with the modules that registered
// them, by link, so the table reads both in one pass.
func SidebarRows(items []model.Sidebar, modules []model.Module) []SidebarRow {
	byHref := make(map[string]*model.Module, len(modules))
	for i := range modules {
		byHref["/m/"+modules[i].Name] = &modules[i]
	}
	rows := make([]SidebarRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, SidebarRow{Item: item, Module: byHref[item.Href]})
	}
	return rows
}
