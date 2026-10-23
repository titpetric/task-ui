package keystore

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// Rotate writes new material for a key of its algorithm and size, and
// moves the previous material to backup/<name>.key and .pub, which the
// key serves beside the new one: an SSH key loads both into an agent,
// an encryption key decrypts with both. A key keeps one backup: while
// it has one, Rotate answers ErrBackup.
func (l *Local) Rotate(ctx context.Context, name string) (*Key, error) {
	row, err := l.Rows.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if row.Backup != "" || exists(l.backupPath(name, ".key")) {
		return nil, fmt.Errorf("%w: %s", ErrBackup, BackupName(name))
	}
	private, _, err := generate(row.Algo, int(row.Bits))
	if err != nil {
		return nil, err
	}
	block, err := ssh.MarshalPrivateKey(private, row.Comment)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := os.MkdirAll(filepath.Join(l.Dir, backupDir), 0o700); err != nil {
		return nil, err
	}
	for _, ext := range []string{".key", ".pub"} {
		if err := os.Rename(l.path(name, ext), l.backupPath(name, ext)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	if err := l.write(name, pem.EncodeToMemory(block), signer.PublicKey(), row.Comment); err != nil {
		return nil, err
	}
	if err := l.Rows.SetBackup(ctx, name, ssh.FingerprintSHA256(signer.PublicKey()), BackupName(name), row.Fingerprint); err != nil {
		return nil, err
	}
	return l.Get(ctx, name)
}

// DeleteBackup removes the backup a rotation kept of a key, its files
// and its record, after which the key rotates again. A key without one
// answers ErrNoBackup.
func (l *Local) DeleteBackup(ctx context.Context, name string) error {
	row, err := l.Rows.Get(ctx, name)
	if err != nil {
		return err
	}
	if row.Backup == "" && !exists(l.backupPath(name, ".key")) {
		return fmt.Errorf("%w: %s", ErrNoBackup, name)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, ext := range []string{".key", ".pub"} {
		if err := os.Remove(l.backupPath(name, ext)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return l.Rows.ClearBackup(ctx, name)
}

// Purge removes a key for good: its row and links, its files, the
// backup of its last rotation and the older backups beside it. A
// workspace's encryption key goes so, once the instance's key encrypts
// what it did.
func (l *Local) Purge(ctx context.Context, name string) error {
	if _, err := l.Rows.Get(ctx, name); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	paths := []string{l.path(name, ".key"), l.path(name, ".pub"), l.backupPath(name, ".key"), l.backupPath(name, ".pub")}
	for _, ext := range []string{".key", ".pub"} {
		older, _ := filepath.Glob(l.path(name, ext+".bak-*"))
		paths = append(paths, older...)
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return l.Rows.Delete(ctx, name)
}

// WorkspaceEncryptKeys returns the encryption keys of workspaces, which
// the instance's one key replaces.
func (l *Local) WorkspaceEncryptKeys(ctx context.Context) ([]Key, error) {
	all, err := l.List(ctx)
	if err != nil {
		return nil, err
	}
	var keys []Key
	for _, key := range all {
		if key.Encrypts() && key.Workspace() != "" {
			keys = append(keys, key)
		}
	}
	return keys, nil
}
