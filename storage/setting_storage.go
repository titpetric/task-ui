package storage

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// SettingStorage keeps the instance settings a page edits, a value by
// name, such as the data retention rules.
type SettingStorage struct {
	db DB
}

// NewSettingStorage returns the storage over a pool.
func NewSettingStorage(db DB) *SettingStorage {
	return &SettingStorage{db: db}
}

// Values returns every setting, by name.
func (s *SettingStorage) Values(ctx context.Context) (map[string]string, error) {
	rows, err := client(s.db).Select[model.Setting](ctx, "SELECT * FROM setting")
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Name] = row.Value
	}
	return values, nil
}

// Set stores settings by name, replacing the values they had, at once.
func (s *SettingStorage) Set(ctx context.Context, values map[string]string) error {
	return inTx(ctx, s.db, func(c conn) error {
		now := stamp()
		for _, name := range model.SortedKeys(values) {
			if err := c.Exec(ctx, "INSERT INTO setting (name, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(name) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at", name, values[name], now); err != nil {
				return err
			}
		}
		return nil
	})
}
