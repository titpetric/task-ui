package view

import (
	"sort"

	"github.com/titpetric/task-ui/folder"
)

// ControlColumns is how many job names the control centre gives a
// column of their own.
const ControlColumns = 7

// Control is the dashboard's control centre: a row per folder with
// jobs, a column per job name that repeats across folders.
type Control struct {
	// Columns are the job names with a column, the most common first.
	Columns []string
	Rows    []ControlRow
}

// ControlRow is a folder in the control centre.
type ControlRow struct {
	Folder *folder.Folder
	// Cells hold the folder's job of each column, nil where it has none.
	Cells []*folder.Job
	// Extra are the folder's jobs without a column, in their order.
	Extra []*folder.Job
}

// NewControl lays the folders out: the job names that more than one
// folder has get a column, the most common first and then by name, up
// to n of them. A job is placed by its name, so a Makefile's build and
// a pipeline's build share the column.
func NewControl(folders []*folder.Folder, n int) Control {
	counts := map[string]int{}
	for _, f := range folders {
		seen := map[string]bool{}
		for _, job := range f.Jobs {
			if !seen[job.Name] {
				seen[job.Name] = true
				counts[job.Name]++
			}
		}
	}
	columns := []string{}
	for name, count := range counts {
		if count > 1 {
			columns = append(columns, name)
		}
	}
	sort.Slice(columns, func(i, j int) bool {
		if counts[columns[i]] != counts[columns[j]] {
			return counts[columns[i]] > counts[columns[j]]
		}
		return columns[i] < columns[j]
	})
	if len(columns) > n {
		columns = columns[:n]
	}
	index := map[string]int{}
	for i, name := range columns {
		index[name] = i
	}

	c := Control{Columns: columns}
	for _, f := range folders {
		if len(f.Jobs) == 0 {
			continue
		}
		row := ControlRow{Folder: f, Cells: make([]*folder.Job, len(columns))}
		for _, job := range f.Jobs {
			if i, ok := index[job.Name]; ok && row.Cells[i] == nil {
				row.Cells[i] = job
				continue
			}
			row.Extra = append(row.Extra, job)
		}
		c.Rows = append(c.Rows, row)
	}
	return c
}
