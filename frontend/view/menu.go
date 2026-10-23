package view

import (
	"fmt"
	"html/template"
)

// Menu is the [...] menu of a row, or of a block, the one way task-ui
// shows the actions a row seldom needs: an icon button that opens a
// list on a click, its details on top, then its items in groups, the
// danger group last. The menu template draws it; docs/development.md
// says when an action goes in it.
type Menu struct {
	// Label names the button for a screen reader and on a hover:
	// "Actions on deploy_ed25519".
	Label   string
	Details []MenuDetail
	// Items are in the order given; an item with Sep ends a group.
	Items []MenuItem
}

// NewMenu returns a menu from its label and parts: each a MenuItem, a MenuDetail, or a false value it skips.
// A false value is one a template counts as false, so (and $cond (item ...)) leaves the item out.
func NewMenu(label string, parts ...any) (*Menu, error) {
	m := &Menu{Label: label}
	for _, part := range parts {
		switch part := part.(type) {
		case MenuItem:
			m.Items = append(m.Items, part)
		case MenuDetail:
			m.Details = append(m.Details, part)
		default:
			if truth, _ := template.IsTrue(part); truth {
				return nil, fmt.Errorf("menu %q: %T is not an item", label, part)
			}
		}
	}
	return m, nil
}

// Empty reports whether the menu has nothing to show, an item or a
// detail; the template draws no button then.
func (m *Menu) Empty() bool {
	if len(m.Details) > 0 {
		return false
	}
	for _, item := range m.Items {
		if !item.Sep {
			return false
		}
	}
	return true
}

// Groups returns the items between separators, empty groups left out, then the danger items as the last group.
func (m *Menu) Groups() [][]MenuItem {
	var groups [][]MenuItem
	var group, danger []MenuItem
	for _, item := range m.Items {
		switch {
		case item.Sep:
			if len(group) > 0 {
				groups = append(groups, group)
			}
			group = nil
		case item.Danger:
			danger = append(danger, item)
		default:
			group = append(group, item)
		}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	if len(danger) > 0 {
		groups = append(groups, danger)
	}
	return groups
}
