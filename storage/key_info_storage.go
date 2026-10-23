package storage

import (
	"context"
	"errors"
	"slices"

	"github.com/titpetric/task-ui/model"
)

// The purposes of a key: it encrypts secrets or it authenticates SSH
// connections, never both.
const (
	KeyEncrypt = "encrypt"
	KeySSH     = "ssh"
)

// ScopeInstance is the scope of a key of the instance; a workspace's key
// is ScopeWorkspace and the workspace's ID.
const ScopeInstance = "instance"

// KeyInfoStorage keeps what each key of the key folder is, by name: its
// purpose, algorithm, fingerprint and scope, and the roles and users
// whose runs get an SSH key. The files are the key folder's; a row
// names them.
type KeyInfoStorage struct {
	db DB
}

// NewKeyInfoStorage returns the storage over a pool.
func NewKeyInfoStorage(db DB) *KeyInfoStorage {
	return &KeyInfoStorage{db: db}
}

// ErrKeyExists reports a key whose name or material the storage holds.
var ErrKeyExists = errors.New("a key of that name or that material is there already: a key has one name and one purpose")

// Create adds a key's row.
func (s *KeyInfoStorage) Create(ctx context.Context, key *model.KeyInfo) error {
	if key.Name == "" || (key.Purpose != KeyEncrypt && key.Purpose != KeySSH) {
		return errors.New("a key needs a name and a purpose, encrypt or ssh")
	}
	if key.Scope == "" {
		key.Scope = ScopeInstance
	}
	if _, err := s.Get(ctx, key.Name); err == nil {
		return ErrKeyExists
	}
	if _, err := s.ByFingerprint(ctx, key.Fingerprint); err == nil {
		return ErrKeyExists
	}
	key.CreatedAt = stamp()
	key.UpdatedAt = key.CreatedAt
	return client(s.db).Insert(ctx, model.KeyInfoTable, key)
}

// Get returns a key's row by name.
func (s *KeyInfoStorage) Get(ctx context.Context, name string) (*model.KeyInfo, error) {
	key, err := client(s.db).Get[model.KeyInfo](ctx, "SELECT * FROM key_info WHERE name = ?", name)
	return key, notFound(err)
}

// ByFingerprint returns the key with a fingerprint.
func (s *KeyInfoStorage) ByFingerprint(ctx context.Context, fingerprint string) (*model.KeyInfo, error) {
	key, err := client(s.db).Get[model.KeyInfo](ctx, "SELECT * FROM key_info WHERE fingerprint = ?", fingerprint)
	return key, notFound(err)
}

// List returns every key, by name.
func (s *KeyInfoStorage) List(ctx context.Context) ([]model.KeyInfo, error) {
	return client(s.db).Select[model.KeyInfo](ctx, "SELECT * FROM key_info ORDER BY name")
}

// SetFingerprint records a key's new material, after a rotation.
func (s *KeyInfoStorage) SetFingerprint(ctx context.Context, name, fingerprint string) error {
	return client(s.db).Exec(ctx, "UPDATE key_info SET fingerprint = ?, updated_at = ? WHERE name = ?", fingerprint, stamp(), name)
}

// SetBackup records a rotation: the key's new fingerprint, and the
// backup of its previous material by name and fingerprint.
func (s *KeyInfoStorage) SetBackup(ctx context.Context, name, fingerprint, backup, backupFingerprint string) error {
	now := stamp()
	return client(s.db).Exec(ctx, "UPDATE key_info SET fingerprint = ?, backup = ?, backup_fingerprint = ?, rotated_at = ?, updated_at = ? WHERE name = ?", fingerprint, backup, backupFingerprint, now, now, name)
}

// ClearBackup forgets a key's backup, once its files went.
func (s *KeyInfoStorage) ClearBackup(ctx context.Context, name string) error {
	return client(s.db).Exec(ctx, "UPDATE key_info SET backup = '', backup_fingerprint = '', updated_at = ? WHERE name = ?", stamp(), name)
}

// Delete removes a key's row and its links.
func (s *KeyInfoStorage) Delete(ctx context.Context, name string) error {
	return inTx(ctx, s.db, func(c conn) error {
		for _, query := range []string{"DELETE FROM key_role WHERE key_name = ?", "DELETE FROM key_user WHERE key_name = ?", "DELETE FROM key_info WHERE name = ?"} {
			if err := c.Exec(ctx, query, name); err != nil {
				return err
			}
		}
		return nil
	})
}

// Link replaces the roles and the users whose runs get a key.
func (s *KeyInfoStorage) Link(ctx context.Context, name string, roles, users []string) error {
	for _, role := range roles {
		if !model.ValidRole(role) {
			return errors.New("unknown role " + role)
		}
	}
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "DELETE FROM key_role WHERE key_name = ?", name); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM key_user WHERE key_name = ?", name); err != nil {
			return err
		}
		for _, role := range unique(roles) {
			if err := c.Insert(ctx, model.KeyRoleTable, &model.KeyRole{ID: model.NewID(), KeyName: name, Role: role, CreatedAt: stamp()}); err != nil {
				return err
			}
		}
		for _, user := range unique(users) {
			if err := c.Insert(ctx, model.KeyUserTable, &model.KeyUser{ID: model.NewID(), KeyName: name, UserName: user, CreatedAt: stamp()}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Links returns every key's roles, most permissive first, and users,
// by key name.
func (s *KeyInfoStorage) Links(ctx context.Context) (map[string][]string, map[string][]string, error) {
	roleRows, err := client(s.db).Select[model.KeyRole](ctx, "SELECT * FROM key_role")
	if err != nil {
		return nil, nil, err
	}
	userRows, err := client(s.db).Select[model.KeyUser](ctx, "SELECT * FROM key_user ORDER BY user_name")
	if err != nil {
		return nil, nil, err
	}
	roles, users := map[string][]string{}, map[string][]string{}
	for _, row := range roleRows {
		roles[row.KeyName] = append(roles[row.KeyName], row.Role)
	}
	for name, list := range roles {
		roles[name] = byRole(list)
	}
	for _, row := range userRows {
		users[row.KeyName] = append(users[row.KeyName], row.UserName)
	}
	return roles, users, nil
}

// unique returns the items without repeats or empties, in order.
func unique(items []string) []string {
	var result []string
	for _, item := range items {
		if item != "" && !slices.Contains(result, item) {
			result = append(result, item)
		}
	}
	return result
}
