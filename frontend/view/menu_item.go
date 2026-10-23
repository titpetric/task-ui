package view

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
)

// attrName is an attribute a menu item may carry besides its own
// fields: data-* for what the script does on a click, title and aria-*.
var attrName = regexp.MustCompile(`^(data-[a-z0-9-]+|aria-[a-z]+|title)$`)

// MenuItem is one entry of a [...] menu: an icon and a label, a link
// when it has Href, else a button whose data-* attributes say what a
// click does, the same attributes a loose button takes (data-delete,
// data-api, data-copy, data-confirm, data-next...).
type MenuItem struct {
	// Icon names the symbol, #i-<Icon> of the layout's sprite.
	Icon  string
	Label string
	// Note is the faint text after the label, such as what turns a
	// disabled item on.
	Note string
	// Href makes the item a link; Download sets its download attribute.
	Href     string
	Download bool
	// Danger puts the item in the last group, in the bad colour.
	Danger   bool
	Disabled bool
	// Attrs are the item's data-*, aria-* and title attributes, their
	// values escaped.
	Attrs template.HTMLAttr
	// Sep ends a group: the menu draws a line before the next item.
	Sep bool
}

// NewMenuItem returns an item from its icon, its label and pairs of an attribute and its value.
// Href, download, note and disabled (its value the note) set fields; data-*, aria-* and title go into Attrs.
// Any other name, or a name without a value, is an error, so a template that misspells one fails to render.
func NewMenuItem(icon, label string, attrs ...string) (MenuItem, error) {
	item := MenuItem{Icon: icon, Label: label}
	if len(attrs)%2 != 0 {
		return item, fmt.Errorf("menu item %q: attribute %q without a value", label, attrs[len(attrs)-1])
	}
	var b strings.Builder
	for i := 0; i < len(attrs); i += 2 {
		name, value := attrs[i], attrs[i+1]
		switch name {
		case "href":
			item.Href = value
		case "download":
			item.Download = true
		case "note":
			item.Note = value
		case "disabled":
			item.Disabled = true
			if item.Note == "" {
				item.Note = value
			}
		default:
			if !attrName.MatchString(name) {
				return item, fmt.Errorf("menu item %q: attribute %q is not data-*, aria-* or title", label, name)
			}
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(name + `="` + html.EscapeString(value) + `"`)
		}
	}
	item.Attrs = template.HTMLAttr(b.String())
	return item, nil
}

// NewDangerItem returns an item as NewMenuItem does, in the danger group: Delete, Kill, Remove.
func NewDangerItem(icon, label string, attrs ...string) (MenuItem, error) {
	item, err := NewMenuItem(icon, label, attrs...)
	item.Danger = true
	return item, err
}
