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
// instance key, the workspaces' on each workspace's key.
func (a *App) secrets(keys *sshkey.Store) *secret.Store {
	store := secret.New(storage.NewSecretStorage(a.DB), secret.NewVault(keys))
	store.Workspaces = &secret.Workspaces{Rows: storage.NewWorkspaceStorage(a.DB), Keys: a.Config.Keyring()}
	return store
}

// migrateSecrets puts every workspace secret on its workspace's key,
// after copying the database to <db>.pre-workspace-keys. With nothing
// to migrate it copies nothing; a copy already there stays.
func (a *App) migrateSecrets(ctx context.Context) error {
	store := a.secrets(&sshkey.Store{Dir: a.Config.SSHDir})
	copyDB := func(ctx context.Context) error {
		file, _, _ := strings.Cut(a.Config.DB, "?")
		if file == "" || file == ":memory:" {
			return nil
		}
		path := file + ".pre-workspace-keys"
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
	result, err := store.Migrate(ctx, copyDB)
	if err != nil {
		return err
	}
	log := slog.Info
	if result.Failed > 0 {
		log = slog.Warn
	}
	log("secrets on their workspace's key", "migrated", result.Migrated, "current", result.Current, "failed", result.Failed)
	return nil
}
