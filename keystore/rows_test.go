package keystore_test

import (
	"context"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// memRows are the key table in memory.
type memRows struct {
	mu   sync.Mutex
	rows map[string]model.KeyInfo
}

var _ keystore.Rows = (*memRows)(nil)

func (m *memRows) Create(_ context.Context, key *model.KeyInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.rows[key.Name]; ok {
		return storage.ErrKeyExists
	}
	if key.Scope == "" {
		key.Scope = storage.ScopeInstance
	}
	now := time.Now()
	key.CreatedAt = &now
	m.rows[key.Name] = *key
	return nil
}

func (m *memRows) Get(_ context.Context, name string) (*model.KeyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[name]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return &row, nil
}

func (m *memRows) ByFingerprint(_ context.Context, fingerprint string) (*model.KeyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.rows {
		if row.Fingerprint == fingerprint {
			return &row, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (m *memRows) List(context.Context) ([]model.KeyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := make([]model.KeyInfo, 0, len(m.rows))
	for _, row := range m.rows {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows, nil
}

func (m *memRows) update(name string, change func(*model.KeyInfo)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[name]
	if !ok {
		return storage.ErrNotFound
	}
	change(&row)
	m.rows[name] = row
	return nil
}

func (m *memRows) SetFingerprint(_ context.Context, name, fingerprint string) error {
	return m.update(name, func(row *model.KeyInfo) { row.Fingerprint = fingerprint })
}

func (m *memRows) SetBackup(_ context.Context, name, fingerprint, backup, backupFingerprint string) error {
	return m.update(name, func(row *model.KeyInfo) {
		now := time.Now()
		row.Fingerprint, row.Backup, row.BackupFingerprint, row.RotatedAt = fingerprint, backup, backupFingerprint, &now
	})
}

func (m *memRows) ClearBackup(_ context.Context, name string) error {
	return m.update(name, func(row *model.KeyInfo) { row.Backup, row.BackupFingerprint = "", "" })
}

func (m *memRows) Delete(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rows, name)
	return nil
}

func (m *memRows) Link(context.Context, string, []string, []string) error { return nil }

func (m *memRows) Links(context.Context) (map[string][]string, map[string][]string, error) {
	return map[string][]string{}, map[string][]string{}, nil
}

// memSettings are the instance settings in memory.
type memSettings struct {
	mu     sync.Mutex
	values map[string]string
}

var _ keystore.Settings = (*memSettings)(nil)

func (s *memSettings) Values(context.Context) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	values := map[string]string{}
	for k, v := range s.values {
		values[k] = v
	}
	return values, nil
}

func (s *memSettings) Set(_ context.Context, values map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range values {
		s.values[k] = v
	}
	return nil
}

// newLocal returns a key folder in a temporary directory over rows in
// memory.
func newLocal(t *testing.T) *keystore.Local {
	t.Helper()
	return &keystore.Local{Dir: filepath.Join(t.TempDir(), "keys"), Rows: &memRows{rows: map[string]model.KeyInfo{}}, Settings: &memSettings{values: map[string]string{}}}
}

func TestRows(t *testing.T) {
	t.Parallel()
	l := newLocal(t)
	keys, err := l.List(context.Background())
	if err != nil || len(keys) != 0 {
		t.Fatalf("an empty folder lists no key: %v %v", keys, err)
	}
}

func TestSettings(t *testing.T) {
	t.Parallel()
	var s keystore.Settings = &memSettings{values: map[string]string{}}
	if err := s.Set(context.Background(), map[string]string{"keys.git": "deploy"}); err != nil {
		t.Fatal(err)
	}
	l := newLocal(t)
	l.Settings = s
	if git := l.Git(context.Background()); git != nil {
		t.Fatalf("a git setting naming no key is no git key: %v", git)
	}
}
