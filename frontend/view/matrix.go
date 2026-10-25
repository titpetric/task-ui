package view

import (
	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/permission"
)

// Matrix is a policy as the permissions pages draw it: a row for each
// section with a cell for each rule, then a row for each rule a section
// adds and each application permission. A cell cycles inherit, allow and deny; Save puts the
// cells that changed to API. A fixed matrix shows every cell as allowed.
type Matrix struct {
	// ID anchors the matrix on its page.
	ID string
	// Title names the role or the user the policy belongs to.
	Title string
	// API is where Save puts the changed cells.
	API string
	// Rules head the columns.
	Rules       []config.Entry
	Sections    []MatrixRow
	Permissions []MatrixCell
	// Fixed is the admin's: every cell allows, none changes.
	Fixed bool
	// Editable lets the cells change and shows Save.
	Editable bool
	// Effective shows, next to each cell, whether the user may do it
	// by the role's policy and their own together.
	Effective bool
}

// NewMatrix lays a policy out on the rules of p, effective what the policy and the role allow together.
func NewMatrix(p *config.Permissions, policy permission.Policy, effective map[string]bool, fixed bool) Matrix {
	m := Matrix{Rules: p.Rules, Fixed: fixed}
	cell := func(path, title string) MatrixCell {
		c := MatrixCell{Path: path, Title: title, Effect: policy[path], Allowed: effective[path] || fixed, Checked: p.Checkable(path)}
		if fixed {
			c.Effect = permission.Allow
		}
		return c
	}
	for _, section := range p.Sections {
		row := MatrixRow{Entry: section}
		for _, rule := range p.Rules {
			row.Cells = append(row.Cells, cell(section.Key+"/"+rule.Key, section.Title+", "+rule.Title))
		}
		m.Sections = append(m.Sections, row)
	}
	// A rule a section adds is a row of its own: no column of every
	// section holds it.
	for _, extra := range p.Extra {
		title := extra.Title
		for _, section := range p.Sections {
			if section.Key == extra.Section {
				title = section.Title + ", " + extra.Title
			}
		}
		m.Permissions = append(m.Permissions, cell(extra.Path(), title))
	}
	for _, entry := range p.Permissions {
		m.Permissions = append(m.Permissions, cell(entry.Key, entry.Title))
	}
	return m
}

// Columns is how many columns a matrix row spans: the section and a cell for each rule.
func (m Matrix) Columns() int {
	return len(m.Rules) + 1
}
