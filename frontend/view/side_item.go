package view

import (
	"strings"

	"github.com/titpetric/task-ui/model"
)

// SideItem is one row of the sidebar's top block, as the sidebar table
// holds it and the page may see it.
//
// Two rows carry more than a link, and the template keys on these rather
// than on the href, so renaming a page does not quietly empty either:
// Workflows lists the instance's workflows under it, and Runs carries the
// dot that pulses while anything is running.
type SideItem struct {
	Title  string
	Href   string
	Icon   string
	Active bool
	Flows  bool
	Pulse  bool
}

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

// SideItems returns the rows of the sidebar a user sees: the table's
// order, less the hidden ones, less the ones whose rule the user lacks.
//
// allowed answers a permission path, settings answers whether the user may
// open any step of the settings, and active reports whether a row is the
// page being read.
func SideItems(rows []model.Sidebar, allowed func(string) bool, settings bool, active func(href string) bool) []SideItem {
	items := make([]SideItem, 0, len(rows))
	for _, row := range rows {
		if row.Hidden || !seesSide(row, allowed, settings) {
			continue
		}
		items = append(items, SideItem{
			Title:  row.Title,
			Href:   row.Href,
			Icon:   sideIcon(row.Icon),
			Active: active(row.Href),
			Flows:  row.Href == "/workflows",
			Pulse:  row.Href == "/runs",
		})
	}
	return items
}

// seesSide reports whether a user may see a row at all.
//
// An empty rule is a row everyone signed in sees. model.RuleSettings is
// the one value that is not a permission: the settings open for whoever
// holds the view rule of any of their sections, which is a question only
// the page can answer, so it is answered before this is called.
func seesSide(row model.Sidebar, allowed func(string) bool, settings bool) bool {
	switch row.Rule {
	case "":
		return true
	case model.RuleSettings:
		return settings
	default:
		return allowed != nil && allowed(row.Rule)
	}
}

// sideIcon keeps an icon to the characters a sprite symbol is named with,
// so a value from the database cannot reach the markup as anything else.
// A name with nothing left falls back to the dot every list already uses.
func sideIcon(icon string) string {
	kept := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			return r
		default:
			return -1
		}
	}, strings.ToLower(strings.TrimSpace(icon)))
	if kept == "" {
		return "box"
	}
	return kept
}
