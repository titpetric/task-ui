package secret

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/titpetric/task-ui/model"
)

// MigrateInstance puts the services' secrets, integration:<kind>, on
// the instance's encryption key: it decrypts a row that key does not
// open with the legacy SSH keys and encrypts it for that key. Before it
// writes anything it calls before, to copy the database, say; with
// nothing to do it does not. Migrate checks the new value decrypts to
// the old before it writes it, and reads the row back after; a row that
// reads back wrong gets its old content again. A row nothing opens
// stays untouched, logged. Run again, it finds every row current and
// does nothing.
func (s *Store) MigrateInstance(ctx context.Context, before func(context.Context) error) (Migration, error) {
	var result Migration
	rows, err := s.instance(ctx)
	if err != nil {
		return result, err
	}
	var work []model.Secret
	for _, row := range rows {
		if Encrypted(row.Content) && opens(s.Vault, row.Content) {
			result.Current++
			continue
		}
		work = append(work, row)
	}
	if len(work) == 0 {
		return result, nil
	}
	if before != nil {
		if err := before(ctx); err != nil {
			return result, fmt.Errorf("before migrating the integrations: %w", err)
		}
	}
	for _, row := range work {
		if err := s.migrateInstance(ctx, row); err != nil {
			result.Failed++
			slog.Warn("secret left as it is", "scope", row.Scope, "name", row.Name, "error", err)
			continue
		}
		result.Migrated++
		slog.Info("secret written for the instance's encryption key", "scope", row.Scope, "name", row.Name)
	}
	return result, nil
}

// migrateInstance writes one row for the instance's encryption key and
// checks it reads back.
func (s *Store) migrateInstance(ctx context.Context, row model.Secret) error {
	plain := row.Content
	if Encrypted(row.Content) {
		var err error
		if plain, err = s.legacy(row.Content); err != nil {
			return fmt.Errorf("neither the instance's encryption key nor the legacy SSH keys open it: %w", err)
		}
	}
	encrypted, err := s.Vault.Encrypt(plain)
	if err != nil {
		return err
	}
	if back, err := s.Vault.Decrypt(encrypted); err != nil || back != plain {
		return errors.Join(errors.New("the instance's encryption key does not read back what it wrote"), err)
	}
	old := row.Content
	row.Content = encrypted
	if err := s.Secrets.Update(ctx, &row); err != nil {
		return err
	}
	stored, err := s.Secrets.Get(ctx, model.Secret{Scope: row.Scope, Name: row.Name})
	if err == nil {
		var back string
		if back, err = s.Vault.Decrypt(stored.Content); err == nil && back != plain {
			err = errors.New("reads back different")
		}
	}
	if err != nil {
		row.Content = old
		return errors.Join(fmt.Errorf("reading %s %s back, the old content restored: %w", row.Scope, row.Name, err), s.Secrets.Update(ctx, &row))
	}
	return nil
}
