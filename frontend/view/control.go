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
// column, and a folder's two builds share its cell. A folder's page lays its Jobs out the same
// way, a row for the folder and one per directory inside with targets, then a row of named
// buttons for its files, each skill and the built-ins.
type Control struct {
	// Columns are the job names with a column, default first, then the
	// most common, then by name.
	Columns []string
	Rows    []ControlRow
}

// ControlRow is a folder in the control centre, or on a folder's page
// the folder, a directory inside it, or a group of named buttons.
type ControlRow struct {
	Folder *folder.Folder
	// Label names the row where the folder's ID does not: "./" for the
	// folder of the page, the directory, the group's kind.
	Label string
	// Title says on a hover what the row holds: its sources, the
	// group's note.
	Title string
	// Cells hold the folder's jobs of each column, none where it has
	// none.
	Cells [][]*folder.Job
	// Extra are the folder's jobs whose names have no column, in their
	// order; a Named row's jobs, every one.
	Extra []*folder.Job
	// Named rows hold their jobs as named buttons across the columns:
	// the files that run, a skill's jobs, the built-ins.
	Named bool
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

// NewJobControl lays a page's Jobs out, a row per place of the groups' targets then a Named row per group; with one place every name has a column.
func NewJobControl(groups []JobGroup) Control {
	var places, named []ControlRow
	index := map[*folder.Folder]int{}
	for _, g := range groups {
		if !g.Targets {
			named = append(named, ControlRow{Folder: g.Folder, Label: g.Kind, Title: g.Note, Extra: g.Jobs, Named: true})
			continue
		}
		i, ok := index[g.Folder]
		if !ok {
			label := "./"
			if g.Dir != "" {
				label = g.Dir + "/"
			}
			places = append(places, ControlRow{Folder: g.Folder, Label: label, Title: g.Kind + " " + g.Note})
			i = len(places) - 1
			index[g.Folder] = i
		} else {
			places[i].Title += "\n" + g.Kind + " " + g.Note
		}
		places[i].Extra = append(places[i].Extra, g.Jobs...)
		if len(places[i].Extra) == 0 {
			places[i].Empty = g.Empty
		}
	}
	min, targets := ControlRepeats, 0
	for _, row := range places {
		if len(row.Extra) > 0 {
			targets++
		}
	}
	if targets < 2 {
		min = 1
	}
	return layout(append(places, named...), min)
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
