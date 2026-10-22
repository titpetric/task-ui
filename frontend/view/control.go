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
// jobs, a column per job name that repeats enough.
type Control struct {
	// Columns are the job names with a column, default first, then the
	// most common.
	Columns []string
	Rows    []ControlRow
}

// ControlRow is a folder in the control centre.
type ControlRow struct {
	Folder *folder.Folder
	// Cells hold the folder's jobs of each column, none where it has
	// none.
	Cells [][]*folder.Job
	// Extra are the folder's jobs whose names have no column, in their
	// order.
	Extra []*folder.Job
}

// NewControl lays the folders out: every job name that occurs min times
// or more, across folders or within one, gets a column, default first,
// then the most common and then by name. The column of a job follows its name, so a Makefile's
// build and a pipeline's build share the column, and a folder's two
// builds share its cell.
func NewControl(folders []*folder.Folder, min int) Control {
	counts := map[string]int{}
	for _, f := range folders {
		for _, job := range f.Jobs {
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
	for _, f := range folders {
		if len(f.Jobs) == 0 {
			continue
		}
		row := ControlRow{Folder: f, Cells: make([][]*folder.Job, len(columns))}
		for _, job := range f.Jobs {
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
