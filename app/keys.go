package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/sshagent"
	"github.com/titpetric/task-ui/storage"
)

// keys returns the key folder, TASK_UI_SSH_DIR, over the database.
func (a *App) keys() *keystore.Local {
	return &keystore.Local{Dir: a.Config.SSHDir, Rows: storage.NewKeyInfoStorage(a.DB), Settings: storage.NewSettingStorage(a.DB)}
}

// ensureKeys makes the key folder, private to the server, removes the
// agent directories a process before this one left, takes the keys of
// the older layout into the folder once, and makes the instance's
// encryption key when there is none: the one key that encrypts every
// secret.
func (a *App) ensureKeys(ctx context.Context) error {
	if a.Config.SSHDir == "" {
		a.Config.SSHDir = filepath.Join(a.Config.DataDir, "keys")
	}
	if err := os.MkdirAll(a.Config.SSHDir, 0o700); err != nil {
		return fmt.Errorf("the key folder: %w", err)
	}
	if err := os.Chmod(a.Config.SSHDir, 0o700); err != nil {
		return fmt.Errorf("the key folder: %w", err)
	}
	if n := sshagent.Clean(""); n > 0 {
		slog.Info("agent directories a previous process left removed", "count", n)
	}
	keys := a.keys()
	workspaces, err := storage.NewWorkspaceStorage(a.DB).List(ctx)
	if err != nil {
		return err
	}
	if _, err := keys.Migrate(ctx, keystore.Legacy{
		Data: a.Config.DataDir, Workspaces: workspaces,
		Policies: storage.NewSSHKeyPolicyStorage(a.DB), Grants: storage.NewWorkspaceKeyGrantStorage(a.DB),
	}); err != nil {
		return fmt.Errorf("moving the keys into the key folder: %w", err)
	}
	if keys.EncryptKey(ctx, storage.ScopeInstance) == nil {
		key, err := keys.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt, Comment: "task-ui instance"})
		if err != nil {
			return fmt.Errorf("the instance's encryption key: %w", err)
		}
		slog.Info("instance encryption key generated", "name", key.Name, "fingerprint", key.Fingerprint)
	}
	return nil
}
