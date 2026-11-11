package storage

import (
	"context"

	"github.com/titpetric/task-ui/app/model"
)

// SidebarStorage keeps the sidebar's top block, one row per item, in the
// order the sidebar lists them.
type SidebarStorage struct {
	db DB
}

// NewSidebarStorage returns the storage over a pool.
func NewSidebarStorage(db DB) *SidebarStorage {
	return &SidebarStorage{db: db}
}

// List returns every item by position, title breaking a tie so the order is stable.
func (s *SidebarStorage) List(ctx context.Context) ([]model.Sidebar, error) {
	return client(s.db).Select[model.Sidebar](ctx, "SELECT * FROM sidebar ORDER BY position, title")
}

// Get returns the item of a link.
func (s *SidebarStorage) Get(ctx context.Context, href string) (*model.Sidebar, error) {
	row, err := client(s.db).Get[model.Sidebar](ctx, "SELECT * FROM sidebar WHERE href = ?", href)
	return row, notFound(err)
}

// EnsureDefaults fills an empty table with items, in the order they are
// given, which is the first start of an instance. A table with a row in
// it stays as it is, so an order the user saved is never overwritten.
func (s *SidebarStorage) EnsureDefaults(ctx context.Context, items []model.Sidebar) error {
	rows, err := s.List(ctx)
	if err != nil || len(rows) > 0 {
		return err
	}
	for i := range items {
		items[i].Position = int64(i)
		if err := s.Save(ctx, &items[i]); err != nil {
			return err
		}
	}
	return nil
}

// Save stores an item over the row its link holds, keeping that row's id,
// position and hidden flag: a module registering again stays where the user
// put it and stays hidden. It reads the row back into item.
func (s *SidebarStorage) Save(ctx context.Context, item *model.Sidebar) error {
	if item.ID == "" {
		item.ID = model.NewID()
	}
	if item.Kind == "" {
		item.Kind = model.SidebarView
	}
	now := stamp()
	err := client(s.db).Exec(ctx,
		"INSERT INTO sidebar (id, kind, href, title, icon, position, hidden, rule, created_at, updated_at)"+
			" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"+
			" ON CONFLICT(href) DO UPDATE SET kind = excluded.kind, title = excluded.title,"+
			" icon = excluded.icon, rule = excluded.rule, updated_at = excluded.updated_at",
		item.ID, item.Kind, item.Href, item.Title, item.Icon,
		item.Position, item.Hidden, item.Rule, now, now)
	if err != nil {
		return err
	}
	stored, err := s.Get(ctx, item.Href)
	if err != nil {
		return err
	}
	*item = *stored
	return nil
}

// Order writes the position and the hidden flag of every row the page posted,
// in one transaction: the position is the index in hrefs, and hidden names the
// links to leave out of the sidebar. A link the table does not hold is skipped,
// because the page lists what the table holds and anything else is a stale form.
func (s *SidebarStorage) Order(ctx context.Context, hrefs []string, hidden map[string]bool) error {
	return inTx(ctx, s.db, func(c conn) error {
		now := stamp()
		for i, href := range hrefs {
			err := c.Exec(ctx, "UPDATE sidebar SET position = ?, hidden = ?, updated_at = ? WHERE href = ?",
				int64(i), hidden[href], now, href)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Keep deletes every module row whose link is not in hrefs, and returns
// how many it dropped. A view's row is never dropped: the instance owns
// those, and a module's goes when its directory does.
func (s *SidebarStorage) Keep(ctx context.Context, hrefs []string) (int, error) {
	db := client(s.db)
	args := append([]any{model.SidebarModule}, anyList(hrefs)...)
	if err := db.Exec(ctx, "DELETE FROM sidebar WHERE kind = ?"+notIn("AND", "href", len(hrefs)), args...); err != nil {
		return 0, err
	}
	return int(db.RowsAffected()), nil
}
