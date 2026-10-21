package secret

import (
	"context"
	"fmt"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Store is the secrets as the rest of task-ui reads them: plain on the
// way in and out, encrypted in storage. The workspace's key encrypts a
// workspace's secrets, those it shares with its folders and each
// folder's own; the instance key encrypts a service's, such as an
// integration's.
type Store struct {
	Secrets *storage.SecretStorage
	// Vault is the instance vault: it encrypts the services' secrets,
	// and every secret when Workspaces is nil.
	Vault *Vault
	// Workspaces finds the workspace of every other scope and the
	// vault of its key. Nil keeps every secret on Vault and every
	// folder in the default workspace.
	Workspaces *Workspaces
}

// New returns a store over the rows and the instance vault.
func New(secrets *storage.SecretStorage, vault *Vault) *Store {
	return &Store{Secrets: secrets, Vault: vault}
}

// vault returns the vault of a scope's secrets.
func (s *Store) vault(ctx context.Context, scope string) (*Vault, error) {
	if s.Workspaces == nil || Service(scope) {
		return s.Vault, nil
	}
	row, err := s.Workspaces.Of(ctx, scope)
	if err != nil {
		return nil, err
	}
	return s.Workspaces.Vault(row), nil
}

// Workspace returns the folder of the workspace a scope belongs to.
func (s *Store) Workspace(ctx context.Context, scope string) (string, error) {
	return s.Workspaces.Name(ctx, scope)
}

// Set creates a secret or replaces its content, encrypted. The scope
// all is the default workspace's shared scope.
func (s *Store) Set(ctx context.Context, secret *model.Secret) error {
	if secret.Scope == storage.ScopeAll {
		secret.Scope = storage.SharedScope(model.Default)
	}
	vault, err := s.vault(ctx, secret.Scope)
	if err != nil {
		return err
	}
	encrypted, err := vault.Encrypt(secret.Content)
	if err != nil {
		return err
	}
	row := *secret
	row.Content = encrypted
	if err := s.Secrets.Set(ctx, &row); err != nil {
		return err
	}
	secret.ID, secret.CreatedAt, secret.UpdatedAt = row.ID, row.CreatedAt, row.UpdatedAt
	return nil
}

// open decrypts a row's content in place, with its scope's vault.
func (s *Store) open(ctx context.Context, secret *model.Secret) error {
	vault, err := s.vault(ctx, secret.Scope)
	if err == nil {
		secret.Content, err = vault.Decrypt(secret.Content)
	}
	if err != nil {
		return fmt.Errorf("secret %s for %s: %w", secret.Name, secret.Scope, err)
	}
	return nil
}

// Get returns a secret by scope and name, decrypted.
func (s *Store) Get(ctx context.Context, ref model.Secret) (*model.Secret, error) {
	secret, err := s.Secrets.Get(ctx, ref)
	if err != nil {
		return nil, err
	}
	return secret, s.open(ctx, secret)
}

// List returns every secret, decrypted.
func (s *Store) List(ctx context.Context) ([]model.Secret, error) {
	secrets, err := s.Secrets.List(ctx)
	if err != nil {
		return nil, err
	}
	return secrets, s.openAll(ctx, secrets)
}

// Scope returns the secrets of one scope, decrypted.
func (s *Store) Scope(ctx context.Context, scope string) ([]model.Secret, error) {
	secrets, err := s.Secrets.Scoped(ctx, scope)
	if err != nil {
		return nil, err
	}
	return secrets, s.openAll(ctx, secrets)
}

func (s *Store) openAll(ctx context.Context, secrets []model.Secret) error {
	for i := range secrets {
		if err := s.open(ctx, &secrets[i]); err != nil {
			return err
		}
	}
	return nil
}

// For returns the secrets that reach a folder by name, decrypted: those
// its workspace shares, with the folder's own winning. A folder in a
// named workspace does not get the default workspace's.
func (s *Store) For(ctx context.Context, folder string) (map[string]string, error) {
	workspace, err := s.Workspace(ctx, folder)
	if err != nil {
		return nil, err
	}
	secrets, err := s.Secrets.Scoped(ctx, storage.SharedScope(workspace), folder)
	if err != nil {
		return nil, err
	}
	if err := s.openAll(ctx, secrets); err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, secret := range secrets {
		result[secret.Name] = secret.Content
	}
	return result, nil
}

// Delete removes a secret by ID.
func (s *Store) Delete(ctx context.Context, id string) error {
	return s.Secrets.Delete(ctx, id)
}

// Count returns how many secrets there are.
func (s *Store) Count(ctx context.Context) (int, error) {
	secrets, err := s.Secrets.List(ctx)
	return len(secrets), err
}

// OnInstance returns how many secrets the instance key encrypts.
func (s *Store) OnInstance(ctx context.Context) (int, error) {
	secrets, err := s.instance(ctx)
	return len(secrets), err
}

// instance returns the rows the instance key encrypts.
func (s *Store) instance(ctx context.Context) ([]model.Secret, error) {
	return s.rows(ctx, func(scope string) (bool, error) {
		return s.Workspaces == nil || Service(scope), nil
	})
}

// rows returns the rows whose scope the filter keeps, still encrypted.
func (s *Store) rows(ctx context.Context, keep func(scope string) (bool, error)) ([]model.Secret, error) {
	all, err := s.Secrets.List(ctx)
	if err != nil {
		return nil, err
	}
	var result []model.Secret
	for _, row := range all {
		ok, err := keep(row.Scope)
		if err != nil {
			return nil, err
		}
		if ok {
			result = append(result, row)
		}
	}
	return result, nil
}

// inWorkspace returns the rows of a workspace's secrets, still
// encrypted: its shared ones and its folders'.
func (s *Store) inWorkspace(ctx context.Context, workspace string) ([]model.Secret, error) {
	return s.rows(ctx, func(scope string) (bool, error) {
		if Service(scope) {
			return false, nil
		}
		name, err := s.Workspace(ctx, scope)
		return name == workspace, err
	})
}

// Encrypted returns the rows of a workspace's secrets as storage keeps
// them, encrypted to its key: its shared ones and its folders'. An
// export carries them unchanged.
func (s *Store) Encrypted(ctx context.Context, workspace string) ([]model.Secret, error) {
	return s.inWorkspace(ctx, workspace)
}

// Rekey writes every secret again for its scope's current key, the
// excluded names of the instance key left out: after a rotation, or
// before a key goes. It returns the number it wrote.
func (s *Store) Rekey(ctx context.Context, exclude ...string) (int, error) {
	secrets, err := s.Secrets.List(ctx)
	if err != nil {
		return 0, err
	}
	return s.rekey(ctx, secrets, exclude)
}

// RekeyInstance writes the secrets the instance key encrypts again for
// its current key, the excluded names left out.
func (s *Store) RekeyInstance(ctx context.Context, exclude ...string) (int, error) {
	secrets, err := s.instance(ctx)
	if err != nil {
		return 0, err
	}
	return s.rekey(ctx, secrets, exclude)
}

// RekeyWorkspace writes a workspace's secrets again for its current
// key, after a rotation; the backup of the previous key opens them.
func (s *Store) RekeyWorkspace(ctx context.Context, workspace string) (int, error) {
	secrets, err := s.inWorkspace(ctx, workspace)
	if err != nil {
		return 0, err
	}
	return s.rekey(ctx, secrets, nil)
}

// rekey writes rows again, each with its scope's vault; the excluded
// names apply to the instance vault.
func (s *Store) rekey(ctx context.Context, secrets []model.Secret, exclude []string) (int, error) {
	for i := range secrets {
		vault, err := s.vault(ctx, secrets[i].Scope)
		if err != nil {
			return i, err
		}
		if err := s.open(ctx, &secrets[i]); err != nil {
			return i, err
		}
		var skip []string
		if vault == s.Vault {
			skip = exclude
		}
		encrypted, err := vault.Encrypt(secrets[i].Content, skip...)
		if err != nil {
			return i, err
		}
		secrets[i].Content = encrypted
		if err := s.Secrets.Update(ctx, &secrets[i]); err != nil {
			return i, err
		}
	}
	return len(secrets), nil
}

// Forget removes a workspace's secrets, its shared ones and its
// folders', when the workspace and its key go: nothing opens them after.
// It runs while the workspace's row is there.
func (s *Store) Forget(ctx context.Context, workspace string) (int, error) {
	secrets, err := s.inWorkspace(ctx, workspace)
	if err != nil {
		return 0, err
	}
	for i, secret := range secrets {
		if err := s.Secrets.Delete(ctx, secret.ID); err != nil {
			return i, err
		}
	}
	return len(secrets), nil
}
