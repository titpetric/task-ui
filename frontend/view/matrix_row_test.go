package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A row is a section of the configuration with a cell for each rule.
func TestMatrixRow(t *testing.T) {
	t.Parallel()
	p := config.Default()
	var row view.MatrixRow = view.NewMatrix(p, nil, nil, false).Sections[0]
	assert.Equal(t, p.Sections[0], row.Entry)
	assert.Len(t, row.Cells, len(p.Rules))
	assert.Equal(t, p.Sections[0].Key+"/"+p.Rules[0].Key, row.Cells[0].Path)
}
