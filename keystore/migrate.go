package keystore

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// migratedSetting records that the keys of the older layout went into
// the key folder.
const migratedSetting = "keys.migrated"

// Migrate takes the keys of the older layout into the folder once, as
// named keys, and records it: the instance's encryption key as
// task_encryption_<algo>, a workspace's as
// workspace_<folder>_encryption_<algo> and its SSH keys as
// workspace_<folder>_<label>_<algo>, the folder's id_<type> as
// task_ssh_<algo>, which git pulls with, and id_<type>_<label> as
// <label>_<algo>, each with the roles and users the policies gave it.
// Backups come along, to decrypt what they encrypted. The old files
// stay where they are; the log names them. A file it cannot take stays,
// with a warning.
func (l *Local) Migrate(ctx context.Context, legacy Legacy) ([]Moved, error) {
	values, err := l.Settings.Values(ctx)
	if err != nil || values[migratedSetting] != "" {
		return nil, err
	}
	m := &migration{l: l, legacy: legacy}
	m.encrypt(ctx, filepath.Join(legacy.Data, "instance", "encrypt"), "task_encryption", storage.ScopeInstance)
	for _, ws := range legacy.Workspaces {
		dir := filepath.Join(legacy.Data, "workspaces", ws.ID)
		m.encrypt(ctx, filepath.Join(dir, "encrypt"), WorkspaceKeyName(ws.Folder, "encryption"), ScopeWorkspace(ws.ID))
		m.workspaceSSH(ctx, filepath.Join(dir, "ssh"), ws.Folder, ws.ID)
	}
	m.instanceSSH(ctx)
	if err := l.Settings.Set(ctx, map[string]string{migratedSetting: "done"}); err != nil {
		return m.moved, err
	}
	for _, moved := range m.moved {
		slog.Info("key moved into the key folder", "from", moved.From, "name", moved.Name, "fingerprint", moved.Fingerprint)
	}
	if len(m.moved) > 0 {
		slog.Info("the old key files stay where they were and nothing reads them; remove them once the keys work", "count", len(m.moved))
	}
	return m.moved, nil
}

// migration is one run of Migrate.
type migration struct {
	l      *Local
	legacy Legacy
	moved  []Moved
}

// take imports one old key file under a base name, with its backups.
func (m *migration) take(ctx context.Context, path, name, purpose, scope string) *Key {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	comment := ""
	if pub, err := os.ReadFile(path + ".pub"); err == nil {
		if fields := strings.Fields(string(pub)); len(fields) > 2 {
			comment = strings.Join(fields[2:], " ")
		}
	}
	key, err := m.l.Import(ctx, Spec{Name: name, Purpose: purpose, Scope: scope, Comment: comment}, data)
	if err != nil {
		slog.Warn("key left where it is", "file", path, "err", err)
		return nil
	}
	backups, _ := filepath.Glob(path + ".bak-*")
	for _, backup := range backups {
		if strings.HasSuffix(backup, ".pub") {
			continue
		}
		if data, err := os.ReadFile(backup); err == nil {
			stamp := strings.TrimPrefix(backup, path)
			_ = os.WriteFile(m.l.path(key.Name, ".key"+stamp), data, 0o600)
		}
	}
	m.moved = append(m.moved, Moved{From: path, Name: key.Name, Fingerprint: key.Fingerprint})
	return key
}

// keyFiles returns the current id_* private key files of a folder by
// name, backups and public halves left out.
func keyFiles(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type().IsRegular() && sshkey.KindOf(name) != "" && !strings.HasSuffix(name, ".pub") {
			names = append(names, name)
		}
	}
	return names
}

// encrypt takes the encryption key of a folder.
func (m *migration) encrypt(ctx context.Context, dir, name, scope string) {
	for _, file := range keyFiles(dir) {
		if m.take(ctx, filepath.Join(dir, file), name, PurposeEncrypt, scope) != nil {
			return
		}
	}
}

// label returns the name an old key file had after its type, or the
// fallback for a type's own key.
func label(file, fallback string) string {
	kind := sshkey.KindOf(file)
	if rest := strings.TrimPrefix(strings.TrimPrefix(file, sshkey.Types[kind]), "_"); rest != "" {
		return rest
	}
	return fallback
}

// workspaceSSH takes a workspace's SSH keys, with their grants.
func (m *migration) workspaceSSH(ctx context.Context, dir, folder, id string) {
	var roles, users map[string][]string
	if m.legacy.Grants != nil {
		roles, users, _ = m.legacy.Grants.All(ctx, id)
	}
	for _, file := range keyFiles(dir) {
		key := m.take(ctx, filepath.Join(dir, file), WorkspaceKeyName(folder, label(file, "ssh")), PurposeSSH, ScopeWorkspace(id))
		if key != nil {
			m.link(ctx, key.Name, roles[file], users[file])
		}
	}
}

// instanceSSH takes the id_* keys of the key folder, with their
// policies; the type's own key, task_ssh_<algo>, pulls with git.
func (m *migration) instanceSSH(ctx context.Context) {
	var roles, users map[string][]string
	if m.legacy.Policies != nil {
		roles, _ = m.legacy.Policies.All(ctx)
		users, _ = m.legacy.Policies.AllUsers(ctx)
	}
	git := ""
	for _, file := range keyFiles(m.l.Dir) {
		key := m.take(ctx, filepath.Join(m.l.Dir, file), label(file, "task_ssh"), PurposeSSH, storage.ScopeInstance)
		if key == nil {
			continue
		}
		m.link(ctx, key.Name, roles[file], users[file])
		if label(file, "") == "" && (git == "" || key.Algo == "ed25519") {
			git = key.Name
		}
	}
	if git != "" && m.l.gitName(ctx) == "" {
		if err := m.l.SetGit(ctx, git); err != nil {
			slog.Warn("the git key is not set", "key", git, "err", err)
		}
	}
}

// link sets on a moved key the roles and users it had.
func (m *migration) link(ctx context.Context, name string, roles, users []string) {
	if len(roles)+len(users) == 0 {
		return
	}
	if err := m.l.Link(ctx, name, roles, users); err != nil {
		slog.Warn("the key's policy did not move", "key", name, "err", err)
	}
}

// WorkspaceKeyName returns the name, before the algorithm, of a key of
// a workspace by its folder: workspace_<folder>_<what>, the default
// workspace, _, as default.
func WorkspaceKeyName(folder, what string) string {
	if folder == model.Default {
		folder = "default"
	}
	return "workspace_" + strings.ReplaceAll(folder, ".", "-") + "_" + what
}
