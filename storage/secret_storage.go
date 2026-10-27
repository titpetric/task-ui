package storage

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// ScopeAll is the scope of a variable that reaches every workspace.
const ScopeAll = "all"

// ScopeWorkspace starts the scope of the variables of one workspace:
// workspace:<folder>.
const ScopeWorkspace = "workspace:"

// SharedScope returns the scope of a workspace's variables.
func SharedScope(workspace string) string {
	return ScopeWorkspace + workspace
}

// ScopeWorkspaceName returns the workspace a workspace scope names, and
// whether the scope names one at all.
func ScopeWorkspaceName(scope string) (string, bool) {
	name, ok := strings.CutPrefix(scope, ScopeWorkspace)
	return name, ok
}

var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ErrInvalidName reports a variable name a shell would not accept.
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

// GetByID returns a secret by its row ID.
func (s *SecretStorage) GetByID(ctx context.Context, id string) (*model.Secret, error) {
	secret, err := client(s.db).Get[model.Secret](ctx, "SELECT * FROM secret WHERE id = ?", id)
	return secret, notFound(err)
}

// List returns every secret, the ones of every workspace first, then a
// workspace's, then a folder's, by scope and name.
func (s *SecretStorage) List(ctx context.Context) ([]model.Secret, error) {
	return client(s.db).Select[model.Secret](ctx, "SELECT * FROM secret ORDER BY CASE WHEN scope = ? THEN 0 WHEN scope LIKE ? THEN 1 ELSE 2 END, scope, name", ScopeAll, ScopeWorkspace+"%")
}

// Scoped returns the secrets of the scopes given, in their order, then
// by name.
func (s *SecretStorage) Scoped(ctx context.Context, scopes ...string) ([]model.Secret, error) {
	var result []model.Secret
	for _, scope := range scopes {
		secrets, err := client(s.db).Select[model.Secret](ctx, "SELECT * FROM secret WHERE scope = ? ORDER BY name", scope)
		if err != nil {
			return nil, err
		}
		result = append(result, secrets...)
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
