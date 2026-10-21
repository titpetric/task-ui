package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// secrets returns the secret store: the services' secrets on the
// instance's encryption key, the workspaces' on each workspace's key.
// No SSH key encrypts or decrypts with it.
func (a *App) secrets() *secret.Store {
	keyring := a.Config.Keyring()
	store := secret.New(storage.NewSecretStorage(a.DB), secret.NewVault(keyring.Instance()))
	store.Workspaces = &secret.Workspaces{Rows: storage.NewWorkspaceStorage(a.DB), Keys: keyring}
	return store
}

// copyDB returns what copies the database to <db>.<suffix> before a
// migration writes: a copy already there stays, an in-memory database
// has none.
func (a *App) copyDB(suffix string) func(context.Context) error {
	return func(ctx context.Context) error {
		file, _, _ := strings.Cut(a.Config.DB, "?")
		if file == "" || file == ":memory:" {
			return nil
		}
		path := file + "." + suffix
		err := storage.Copy(ctx, a.DB, path)
		if errors.Is(err, os.ErrExist) {
			slog.Info("database copy from before the secrets migration kept", "file", path)
			return nil
		}
		if err == nil {
			slog.Info("database copied before migrating the secrets", "file", path)
		}
		return err
	}
}

// migrateSecrets puts every workspace secret on its workspace's key,
// after copying the database to <db>.pre-workspace-keys, then every
// integration on the instance's encryption key, after copying it to
// <db>.pre-keysplit. Only these two read with the legacy SSH keys, the
// ones in TASK_UI_SSH_DIR that encrypted the secrets before. With
// nothing to migrate they copy nothing.
func (a *App) migrateSecrets(ctx context.Context) error {
	store := a.secrets()
	store.Legacy = secret.NewVault(&sshkey.Store{Dir: a.Config.SSHDir})
	result, err := store.Migrate(ctx, a.copyDB("pre-workspace-keys"))
	if err != nil {
		return err
	}
	logMigration("secrets on their workspace's key", result)
	result, err = store.MigrateInstance(ctx, a.copyDB("pre-keysplit"))
	if err != nil {
		return err
	}
	logMigration("integrations on the instance's encryption key", result)
	return nil
}

// logMigration logs what a migration did, a warning when it left a
// secret it could not open.
func logMigration(msg string, result secret.Migration) {
	log := slog.Info
	if result.Failed > 0 {
		log = slog.Warn
	}
	log(msg, "migrated", result.Migrated, "current", result.Current, "failed", result.Failed)
}
