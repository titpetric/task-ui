// Package integration keeps the settings of the services task-ui talks
// to, such as Claude Code and Beszel: API keys, tokens, addresses. They are
// secrets in a scope of their own, integration:<kind>, encrypted with the
// instance key, which the workspaces' secrets do not use, and rekeyed
// with it. That scope never
// matches a folder, so no run sees them as secrets; an integration passes
// its settings to the runs it serves.
package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
)

// ScopePrefix starts the scope of an integration's secrets.
const ScopePrefix = "integration:"

// ErrUnknown reports an integration or a setting task-ui does not know.
var ErrUnknown = errors.New("unknown integration")

// Store reads and writes the integrations' settings.
type Store struct {
	Secrets *secret.Store
}

// Values returns an integration's settings, decrypted, by field name.
func (s *Store) Values(ctx context.Context, id string) (map[string]string, error) {
	if KindOf(id) == nil {
		return nil, ErrUnknown
	}
	rows, err := s.Secrets.Scope(ctx, ScopePrefix+id)
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Name] = row.Content
	}
	return values, nil
}

// Save writes an integration's settings. An empty secret field keeps the
// value it has, so a form never has to show one; an empty plain field
// clears it.
func (s *Store) Save(ctx context.Context, id string, values map[string]string) error {
	kind := KindOf(id)
	if kind == nil {
		return ErrUnknown
	}
	for name := range values {
		if kind.Field(name) == nil {
			return fmt.Errorf("%w setting %s for %s", ErrUnknown, name, id)
		}
	}
	current, err := s.rows(ctx, id)
	if err != nil {
		return err
	}
	for _, field := range kind.Fields {
		value := strings.TrimSpace(values[field.Name])
		row, exists := current[field.Name]
		switch {
		case value == "" && field.Secret:
			continue
		case value == "" && exists:
			if err := s.Secrets.Delete(ctx, row.ID); err != nil {
				return err
			}
		case value != "":
			if err := s.Secrets.Set(ctx, &model.Secret{Scope: ScopePrefix + id, Name: field.Name, Content: value}); err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete removes every setting of an integration.
func (s *Store) Delete(ctx context.Context, id string) error {
	if KindOf(id) == nil {
		return ErrUnknown
	}
	rows, err := s.rows(ctx, id)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := s.Secrets.Delete(ctx, row.ID); err != nil {
			return err
		}
	}
	return nil
}

// Env returns what the integrations add to a run by a user of a role: the
// Claude token, when the Claude integration names the user or the role, or
// names nobody.
func (s *Store) Env(ctx context.Context, user, role string) (map[string]string, error) {
	values, err := s.Values(ctx, Claude)
	if err != nil || values["token"] == "" {
		return nil, err
	}
	if strings.TrimSpace(values["who"]) != "" && !listed(values["who"], user) && !listed(values["who"], role) {
		return nil, nil
	}
	return map[string]string{"CLAUDE_CODE_OAUTH_TOKEN": values["token"]}, nil
}

// listed reports whether a comma separated list names the item, or is *.
func listed(list, item string) bool {
	for _, name := range strings.Split(list, ",") {
		name = strings.TrimSpace(name)
		if name == "*" || (name != "" && name == item) {
			return true
		}
	}
	return false
}

// rows returns an integration's secret rows, still encrypted, by name.
func (s *Store) rows(ctx context.Context, id string) (map[string]model.Secret, error) {
	all, err := s.Secrets.Secrets.List(ctx)
	if err != nil {
		return nil, err
	}
	rows := map[string]model.Secret{}
	for _, row := range all {
		if row.Scope == ScopePrefix+id {
			rows[row.Name] = row
		}
	}
	return rows, nil
}

// Internal reports whether a secret's scope belongs to an integration, hidden from the secrets page.
func Internal(scope string) bool {
	return strings.HasPrefix(scope, ScopePrefix)
}

// Configured reports whether every setting of an integration that is not
// optional has a value.
func Configured(kind Kind, values map[string]string) bool {
	for _, field := range kind.Fields {
		if !field.Optional && values[field.Name] == "" {
			return false
		}
	}
	return true
}
