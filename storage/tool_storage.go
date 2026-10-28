package storage

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// ToolStorage keeps the tools the instance found in its PATH at start,
// one a name, replaced at each start.
type ToolStorage struct {
	db DB
}

// NewToolStorage returns the storage over a pool.
func NewToolStorage(db DB) *ToolStorage {
	return &ToolStorage{db: db}
}

// List returns the tools, by category and then name.
func (s *ToolStorage) List(ctx context.Context) ([]model.Tool, error) {
	return client(s.db).Select[model.Tool](ctx, "SELECT * FROM tool ORDER BY category, name")
}

// Save stores one tool in place of its row.
func (s *ToolStorage) Save(ctx context.Context, tool model.Tool) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "DELETE FROM tool WHERE name = ?", tool.Name); err != nil {
			return err
		}
		tool.UpdatedAt = stamp()
		return c.Insert(ctx, model.ToolTable, &tool)
	})
}

// Remove deletes a tool's row, for a tool the PATH no longer has.
func (s *ToolStorage) Remove(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM tool WHERE name = ?", name)
	return err
}

// Replace stores the tools in place of every row there is, at once.
func (s *ToolStorage) Replace(ctx context.Context, tools []model.Tool) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "DELETE FROM tool"); err != nil {
			return err
		}
		now := stamp()
		for _, tool := range tools {
			row := tool
			row.UpdatedAt = now
			if err := c.Insert(ctx, model.ToolTable, &row); err != nil {
				return err
			}
		}
		return nil
	})
}
