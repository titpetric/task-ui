package view

import (
	"github.com/titpetric/task-ui/config"
)

// MatrixRow is a section of a matrix with a cell for each rule.
type MatrixRow struct {
	config.Entry
	Cells []MatrixCell
}
