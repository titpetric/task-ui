package view

import (
	"fmt"
	"html/template"
)

// Head is the head of a block: its title and the note beside it, its
// buttons on the right, and the one line an empty list shows where its
// table goes, "No keys added.", so a list whose only actions are to add
// to it draws no empty table. The blockhead template draws it;
// docs/development.md, Lists and actions, says what goes where.
type Head struct {
	Title string
	Note  string
	// Empty is the line of an empty list, "" while the list has rows.
	Empty string
	// Buttons are the block's actions, a link or a button each, drawn
	// small, the danger ones in the bad colour.
	Buttons []MenuItem
}

// NewHead returns a head from its title, its note, the line of the empty list and its buttons.
// empty is a string, or a false value for a list with rows, so (and (not .Keys) "No keys added.") passes.
// A button is a MenuItem; head skips a false value, as menu does.
func NewHead(title, note string, empty any, buttons ...any) (*Head, error) {
	h := &Head{Title: title, Note: note}
	if line, ok := empty.(string); ok {
		h.Empty = line
	} else if truth, _ := template.IsTrue(empty); truth {
		return nil, fmt.Errorf("head %q: the empty line %T is not a string", title, empty)
	}
	for _, button := range buttons {
		if item, ok := button.(MenuItem); ok {
			h.Buttons = append(h.Buttons, item)
			continue
		}
		if truth, _ := template.IsTrue(button); truth {
			return nil, fmt.Errorf("head %q: %T is not a button", title, button)
		}
	}
	return h, nil
}
