package storage

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// HostStorage reads the host table, the inventory the Inventory page keeps:
// the machines a person entered and the ones the Beszel hub reported. The
// page owns the writes, in modules/inventory/storage/Hosts.php; a run on
// the hosts reads this to know which addresses are hosts at all.
type HostStorage struct {
	db DB
}

// NewHostStorage returns the storage over a pool.
func NewHostStorage(db DB) *HostStorage {
	return &HostStorage{db: db}
}

// List returns the hosts by name.
func (s *HostStorage) List(ctx context.Context) ([]model.Host, error) {
	hosts, err := client(s.db).Select[model.Host](ctx, "SELECT * FROM host ORDER BY name")
	if hosts == nil {
		hosts = []model.Host{}
	}
	return hosts, err
}

// Names returns the name of each host by the address it answers on, and
// nothing at all where the table cannot be read: a run on the hosts falls
// back to what the hub reports.
func (s *HostStorage) Names(ctx context.Context) map[string]string {
	hosts, err := s.List(ctx)
	if err != nil {
		return nil
	}
	out := make(map[string]string, len(hosts))
	for _, host := range hosts {
		out[host.Address] = host.Name
	}
	return out
}
