package secret

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// pending is a row Migrate writes again: under its target scope, for
// the target's vault.
type pending struct {
	row    model.Secret
	target string
	vault  *Vault
}

// Migrate puts every secret but the services' on its workspace's key:
// the shared ones, scope all, become the default workspace's,
// workspace:_, and Migrate decrypts a row its workspace's key does not
// open with the legacy SSH keys and encrypts it for the workspace's. Before
// it writes anything it calls before, to copy the database, say; with
// nothing to do it does not. Migrate removes a row only after it wrote
// the replacement and read it back; a row nothing opens stays untouched,
// logged. Run again, it finds every row current and does nothing.
func (s *Store) Migrate(ctx context.Context, before func(context.Context) error) (Migration, error) {
	var result Migration
	if s.Workspaces == nil {
		return result, nil
	}
	work, err := s.pending(ctx, &result)
	if err != nil || len(work) == 0 {
		return result, err
	}
	if before != nil {
		if err := before(ctx); err != nil {
			return result, fmt.Errorf("before migrating the secrets: %w", err)
		}
	}
	for _, item := range work {
		if err := s.migrate(ctx, item); err != nil {
			result.Failed++
			slog.Warn("secret left as it is", "scope", item.row.Scope, "name", item.row.Name, "error", err)
			continue
		}
		result.Migrated++
		slog.Info("secret written for its workspace's key", "scope", item.row.Scope, "to", item.target, "name", item.row.Name)
	}
	return result, nil
}

// pending returns the rows Migrate writes again, counting those
// current and those it cannot place.
func (s *Store) pending(ctx context.Context, result *Migration) ([]pending, error) {
	rows, err := s.Secrets.List(ctx)
	if err != nil {
		return nil, err
	}
	var work []pending
	for _, row := range rows {
		if Service(row.Scope) {
			continue
		}
		target := row.Scope
		if target == storage.ScopeAll {
			target = storage.SharedScope(model.Default)
		}
		vault, err := s.vault(ctx, target)
		if err != nil {
			result.Failed++
			slog.Warn("secret left as it is", "scope", row.Scope, "name", row.Name, "error", err)
			continue
		}
		if target == row.Scope && Encrypted(row.Content) && opens(vault, row.Content) {
			result.Current++
			continue
		}
		work = append(work, pending{row: row, target: target, vault: vault})
	}
	return work, nil
}

// opens reports whether a vault decrypts a value.
func opens(vault *Vault, text string) bool {
	_, err := vault.Decrypt(text)
	return err == nil
}

// migrate writes one row for its target vault and checks it reads back.
func (s *Store) migrate(ctx context.Context, item pending) error {
	plain, err := item.vault.Decrypt(item.row.Content)
	if err != nil {
		if plain, err = s.legacy(item.row.Content); err != nil {
			return fmt.Errorf("neither the workspace's key nor the legacy SSH keys open it: %w", err)
		}
	}
	encrypted, err := item.vault.Encrypt(plain)
	if err != nil {
		return err
	}
	if back, err := item.vault.Decrypt(encrypted); err != nil || back != plain {
		return errors.Join(errors.New("the workspace's key does not read back what it wrote"), err)
	}
	if item.target == item.row.Scope {
		row := item.row
		row.Content = encrypted
		if err := s.Secrets.Update(ctx, &row); err != nil {
			return err
		}
		return s.check(ctx, item, plain)
	}
	if _, err := s.Secrets.Get(ctx, model.Secret{Scope: item.target, Name: item.row.Name}); err == nil {
		return fmt.Errorf("%s already has a secret %s", item.target, item.row.Name)
	} else if !errors.Is(err, storage.ErrNotFound) {
		return err
	}
	if err := s.Secrets.Create(ctx, &model.Secret{Scope: item.target, Name: item.row.Name, Content: encrypted}); err != nil {
		return err
	}
	if err := s.check(ctx, item, plain); err != nil {
		return err
	}
	return s.Secrets.Delete(ctx, item.row.ID)
}

// check reads a written row back and decrypts it.
func (s *Store) check(ctx context.Context, item pending, plain string) error {
	row, err := s.Secrets.Get(ctx, model.Secret{Scope: item.target, Name: item.row.Name})
	if err != nil {
		return err
	}
	back, err := item.vault.Decrypt(row.Content)
	if err != nil {
		return fmt.Errorf("reading %s %s back: %w", item.target, item.row.Name, err)
	}
	if back != plain {
		return fmt.Errorf("%s %s reads back different", item.target, item.row.Name)
	}
	return nil
}

// legacy decrypts a value with the SSH keys that encrypted the secrets
// before encryption had keys of its own.
func (s *Store) legacy(text string) (string, error) {
	if s.Legacy == nil {
		return "", errors.New("no legacy SSH keys")
	}
	return s.Legacy.Decrypt(text)
}
