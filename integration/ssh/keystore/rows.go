package keystore

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// Rows are what the store knows of each key, by name: storage's key
// table, or a fake in a test.
type Rows interface {
	Create(ctx context.Context, key *model.KeyInfo) error
	Get(ctx context.Context, name string) (*model.KeyInfo, error)
	ByFingerprint(ctx context.Context, fingerprint string) (*model.KeyInfo, error)
	List(ctx context.Context) ([]model.KeyInfo, error)
	SetFingerprint(ctx context.Context, name, fingerprint string) error
	// SetBackup records a rotation: the key's new fingerprint and its
	// backup, by name and fingerprint; ClearBackup forgets the backup.
	SetBackup(ctx context.Context, name, fingerprint, backup, backupFingerprint string) error
	ClearBackup(ctx context.Context, name string) error
	Delete(ctx context.Context, name string) error
	Link(ctx context.Context, name string, roles, users []string) error
	Links(ctx context.Context) (map[string][]string, map[string][]string, error)
}

// Settings are the instance settings the store keeps the git key in.
type Settings interface {
	Values(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, values map[string]string) error
}
