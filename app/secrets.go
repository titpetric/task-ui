package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// secrets returns the secret store: every secret on the instance's
// encryption key, the workspaces' and the services'. No SSH key
// encrypts or decrypts with it.
func (a *App) secrets() *secret.Store {
	keys := a.keys()
	store := secret.New(storage.NewSecretStorage(a.DB), secret.NewVault(keys.Vault(storage.ScopeInstance)))
	store.Workspaces = &secret.Workspaces{Rows: storage.NewWorkspaceStorage(a.DB)}
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

// migrateSecrets puts every secret on the instance's encryption key:
// the workspaces' secrets, after copying the database to
// <db>.pre-instance-key, then the integrations, after copying it to
// <db>.pre-keysplit. They read with the legacy keys, the SSH keys in
// TASK_UI_SSH_DIR that encrypted the secrets before and the
// workspaces' encryption keys. A workspace's encryption key whose
// secrets the instance's key opens then goes, its row and its files;
// one with a secret it does not open stays, with a warning. With
// nothing to migrate they copy nothing.
func (a *App) migrateSecrets(ctx context.Context) error {
	keys := a.keys()
	old, err := keys.WorkspaceEncryptKeys(ctx)
	if err != nil {
		return err
	}
	legacy := secret.Chain{secret.DirKeys{Store: &sshkey.Store{Dir: a.Config.SSHDir}}}
	for _, key := range old {
		legacy = append(legacy, keys.Vault(key.Scope))
	}
	store := a.secrets()
	store.Legacy = secret.NewVault(legacy)
	result, err := store.Migrate(ctx, a.copyDB("pre-instance-key"))
	if err != nil {
		return err
	}
	logMigration("secrets on the instance's encryption key", result)
	result, err = store.MigrateInstance(ctx, a.copyDB("pre-keysplit"))
	if err != nil {
		return err
	}
	logMigration("integrations on the instance's encryption key", result)
	return a.dropWorkspaceKeys(ctx, keys, store, old)
}

// dropWorkspaceKeys removes the workspaces' encryption keys whose
// secrets the instance's key opens: one encryption key per instance.
func (a *App) dropWorkspaceKeys(ctx context.Context, keys *keystore.Local, store *secret.Store, old []keystore.Key) error {
	workspaces := storage.NewWorkspaceStorage(a.DB)
	for _, key := range old {
		row, err := workspaces.Get(ctx, key.Workspace())
		if err != nil {
			slog.Warn("workspace encryption key kept: its workspace is gone", "name", key.Name, "error", err)
			continue
		}
		held, err := store.Encrypted(ctx, row.Folder)
		if err != nil {
			return err
		}
		unopened, err := store.Unopened(ctx, row.Folder)
		if err != nil {
			return err
		}
		if unopened > 0 {
			slog.Warn("workspace encryption key kept: the instance's key does not open some of its secrets", "name", key.Name, "workspace", row.Folder, "secrets", len(held), "unopened", unopened)
			continue
		}
		if err := keys.Purge(ctx, key.Name); err != nil {
			return fmt.Errorf("removing the workspace encryption key %s: %w", key.Name, err)
		}
		slog.Info("workspace encryption key removed: the instance's key encrypts its secrets", "name", key.Name, "workspace", row.Folder, "secrets", len(held))
	}
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
