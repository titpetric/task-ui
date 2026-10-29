package app

import (
	"context"

	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/toolenv"
)

// toolSummary returns what the stored tooling reads as TASK_UI_TOOLING:
// the tools, one a line, or nothing when there are none or the table
// cannot be read. The rows come from the last Check of the Instance
// page.
func toolSummary(tools *storage.ToolStorage) func(ctx context.Context) string {
	return func(ctx context.Context) string {
		rows, err := tools.List(ctx)
		if err != nil {
			return ""
		}
		detected := make([]toolenv.Tool, 0, len(rows))
		for _, row := range rows {
			detected = append(detected, toolenv.Tool{Name: row.Name, Category: row.Category, Version: row.Version, Path: row.Path, Description: row.Description})
		}
		return toolenv.Env(detected)
	}
}
