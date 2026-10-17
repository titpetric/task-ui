package storage

import (
	"context"
	"errors"
	"regexp"

	"github.com/titpetric/task-ui/model"
)

// ScopeAll is the secret scope that applies to every folder.
const ScopeAll = "all"

var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ErrInvalidName is returned for a variable name a shell would not accept.
var ErrInvalidName = errors.New("a variable name is letters, digits and underscores, and does not start with a digit")

// SecretStorage keeps the variables added to runs.
type SecretStorage struct {
	db DB
}

// NewSecretStorage returns the storage over a pool.
func NewSecretStorage(db DB) *SecretStorage {
	return &SecretStorage{db: db}
}

// ValidName reports whether name is a variable name.
func ValidName(name string) bool {
	return validName.MatchString(name)
}

// Create inserts a secret.
func (s *SecretStorage) Create(ctx context.Context, secret *model.Secret) error {
	if secret.Scope == "" {
		return errors.New("a secret needs a scope")
	}
	if !ValidName(secret.Name) {
		return ErrInvalidName
	}
	if secret.ID == "" {
		secret.ID = model.NewID()
	}
	secret.CreatedAt = stamp()
	secret.UpdatedAt = secret.CreatedAt
	return client(s.db).Insert(ctx, model.SecretTable, secret)
}

// Get returns a secret by scope and name, as the ref names them.
func (s *SecretStorage) Get(ctx context.Context, ref model.Secret) (*model.Secret, error) {
	secret, err := client(s.db).Get[model.Secret](ctx, "SELECT * FROM secret WHERE scope = ? AND name = ?", ref.Scope, ref.Name)
	return secret, notFound(err)
}

// List returns every secret, shared ones first, then by folder and name.
func (s *SecretStorage) List(ctx context.Context) ([]model.Secret, error) {
	return client(s.db).Select[model.Secret](ctx, "SELECT * FROM secret ORDER BY CASE WHEN scope = ? THEN 0 ELSE 1 END, scope, name", ScopeAll)
}

// Scoped returns the secrets that reach a folder: the shared ones and
// the folder's own.
func (s *SecretStorage) Scoped(ctx context.Context, folder string) ([]model.Secret, error) {
	return client(s.db).Select[model.Secret](ctx, "SELECT * FROM secret WHERE scope IN (?, ?) ORDER BY CASE WHEN scope = ? THEN 0 ELSE 1 END, name", ScopeAll, folder, ScopeAll)
}

// For returns the variables a folder's runs receive: the shared ones,
// then the folder's own, which win.
func (s *SecretStorage) For(ctx context.Context, folder string) (map[string]string, error) {
	secrets, err := s.Scoped(ctx, folder)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, secret := range secrets {
		result[secret.Name] = secret.Content
	}
	return result, nil
}

// Set creates a secret, or replaces the content of the one under the
// same scope and name.
func (s *SecretStorage) Set(ctx context.Context, secret *model.Secret) error {
	existing, err := s.Get(ctx, *secret)
	if errors.Is(err, ErrNotFound) {
		return s.Create(ctx, secret)
	}
	if err != nil {
		return err
	}
	secret.ID = existing.ID
	secret.CreatedAt = existing.CreatedAt
	return s.Update(ctx, secret)
}

// Update writes a secret back, by ID.
func (s *SecretStorage) Update(ctx context.Context, secret *model.Secret) error {
	if !ValidName(secret.Name) {
		return ErrInvalidName
	}
	secret.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.SecretTable, secret, "id")
}

// Delete removes a secret by ID.
func (s *SecretStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM secret WHERE id = ?", id)
}
