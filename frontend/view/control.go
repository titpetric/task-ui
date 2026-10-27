package view

import (
	"sort"

	"github.com/titpetric/task-ui/folder"
)

// ControlRepeats is how many times a job name occurs, at least, for the
// control centre to show it as a column: more than once. The column's name
// stands on end, so a column is narrow.
const ControlRepeats = 2

// Control is the dashboard's control centre: a row per folder with
// jobs, a column per job name that repeats enough. A name counts across folders and within one, and
// the column of a job follows its name, so a Makefile's build and a pipeline's build share the
// column, and a folder's two builds share its cell. A folder's page lays its Jobs out
// as named rows: one per job source, its files, each skill and the built-ins.
type Control struct {
	// Columns are the job names with a column, default first, then the
	// most common, then by name.
	Columns []string
	Rows    []ControlRow
}

// ControlRow is a folder in the control centre, or on a folder's page
// a group of named buttons: a job source's targets, the files, a skill's
// jobs or the built-ins.
type ControlRow struct {
	Folder *folder.Folder
	// Label names the row where the folder's ID does not: the source's
	// file, below its directory for one inside, or the group's kind.
	Label string
	// Title says on a hover what the row holds: the source's kind and
	// command, the group's note.
	Title string
	// Cells hold the folder's jobs of each column, none where it has
	// none.
	Cells [][]*folder.Job
	// Extra are the folder's jobs whose names have no column, in their
	// order; a Named row's jobs, every one.
	Extra []*folder.Job
	// Named rows hold their jobs as named buttons beside their label:
	// every row of a folder's page.
	Named bool
	// Place marks a named row of a source's targets, whose label links to
	// the place they run in.
	Place bool
	// Empty is what a row without jobs says, as a source atkins cannot
	// read.
	Empty string
}

// Jobs returns the row's jobs, column by column, then the rest.
func (r ControlRow) Jobs() []*folder.Job {
	var list []*folder.Job
	for _, cell := range r.Cells {
		list = append(list, cell...)
	}
	return append(list, r.Extra...)
}

// NewControl lays the folders out, a column for each job name that occurs min times or more.
func NewControl(folders []*folder.Folder, min int) Control {
	var rows []ControlRow
	for _, f := range folders {
		if len(f.Jobs) > 0 {
			rows = append(rows, ControlRow{Folder: f, Extra: f.Jobs})
		}
	}
	return layout(rows, min)
}

// NewJobControl lays a page's Jobs out as named rows, one per group, a source's under its file.
func NewJobControl(groups []JobGroup) Control {
	c := Control{Columns: []string{}}
	for _, g := range groups {
		row := ControlRow{Folder: g.Folder, Label: g.Kind, Title: g.Note, Extra: g.Jobs, Named: true}
		if g.Targets {
			row.Label, row.Title, row.Place, row.Empty = g.Source, g.Kind+" "+g.Note, true, g.Empty
			if g.Dir != "" {
				row.Label = g.Dir + "/" + g.Source
			}
		}
		c.Rows = append(c.Rows, row)
	}
	return c
}

// layout returns a column for each name the rows' jobs repeat min times, and moves the jobs of a column from Extra to their cell.
func layout(rows []ControlRow, min int) Control {
	counts := map[string]int{}
	for _, row := range rows {
		if row.Named {
			continue
		}
		for _, job := range row.Extra {
			counts[job.Name]++
		}
	}
	columns := []string{}
	for name, count := range counts {
		if count >= min {
			columns = append(columns, name)
		}
	}
	sort.Slice(columns, func(i, j int) bool {
		if (columns[i] == "default") != (columns[j] == "default") {
			return columns[i] == "default"
		}
		if counts[columns[i]] != counts[columns[j]] {
			return counts[columns[i]] > counts[columns[j]]
		}
		return columns[i] < columns[j]
	})
	index := map[string]int{}
	for i, name := range columns {
		index[name] = i
	}
	c := Control{Columns: columns}
	for _, row := range rows {
		if row.Named {
			c.Rows = append(c.Rows, row)
			continue
		}
		jobs := row.Extra
		row.Cells, row.Extra = make([][]*folder.Job, len(columns)), nil
		for _, job := range jobs {
			if i, ok := index[job.Name]; ok {
				row.Cells[i] = append(row.Cells[i], job)
				continue
			}
			row.Extra = append(row.Extra, job)
		}
		c.Rows = append(c.Rows, row)
	}
	return c
}
