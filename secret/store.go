package secret

import (
	"context"
	"fmt"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Store is the secrets as the rest of task-ui reads them: plain on the
// way in and out, encrypted in storage.
type Store struct {
	Secrets *storage.SecretStorage
	Vault   *Vault
}

// New returns a store over the rows and the vault.
func New(secrets *storage.SecretStorage, vault *Vault) *Store {
	return &Store{Secrets: secrets, Vault: vault}
}

// Set creates a secret or replaces its content, encrypted.
func (s *Store) Set(ctx context.Context, secret *model.Secret) error {
	encrypted, err := s.Vault.Encrypt(secret.Content)
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

// open decrypts a row's content in place.
func (s *Store) open(secret *model.Secret) error {
	plain, err := s.Vault.Decrypt(secret.Content)
	if err != nil {
		return fmt.Errorf("secret %s for %s: %w", secret.Name, secret.Scope, err)
	}
	secret.Content = plain
	return nil
}

// Get returns a secret by scope and name, decrypted.
func (s *Store) Get(ctx context.Context, ref model.Secret) (*model.Secret, error) {
	secret, err := s.Secrets.Get(ctx, ref)
	if err != nil {
		return nil, err
	}
	return secret, s.open(secret)
}

// List returns every secret, decrypted.
func (s *Store) List(ctx context.Context) ([]model.Secret, error) {
	secrets, err := s.Secrets.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range secrets {
		if err := s.open(&secrets[i]); err != nil {
			return nil, err
		}
	}
	return secrets, nil
}

// For returns the secrets that reach a folder by name, decrypted: the
// shared ones, with the folder's own winning.
func (s *Store) For(ctx context.Context, folder string) (map[string]string, error) {
	secrets, err := s.Secrets.Scoped(ctx, folder)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for i := range secrets {
		if err := s.open(&secrets[i]); err != nil {
			return nil, err
		}
	}
	for _, secret := range secrets {
		if secret.Scope == storage.ScopeAll {
			result[secret.Name] = secret.Content
		}
	}
	for _, secret := range secrets {
		if secret.Scope == folder {
			result[secret.Name] = secret.Content
		}
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

// Rekey writes every secret again for the current keys, the excluded
// names left out: after a rotation, or before a key goes. It returns
// how many were written.
func (s *Store) Rekey(ctx context.Context, exclude ...string) (int, error) {
	secrets, err := s.Secrets.List(ctx)
	if err != nil {
		return 0, err
	}
	for i := range secrets {
		if err := s.open(&secrets[i]); err != nil {
			return i, err
		}
		encrypted, err := s.Vault.Encrypt(secrets[i].Content, exclude...)
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
