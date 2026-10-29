package keystore

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

var _ Rows = (*storage.KeyInfoStorage)(nil)

// The purposes of a key.
const (
	PurposeEncrypt = storage.KeyEncrypt
	PurposeSSH     = storage.KeySSH
)

// scopeWorkspace starts the scope of a workspace's key.
const scopeWorkspace = "workspace:"

// ScopeWorkspace returns the scope of a workspace's keys.
func ScopeWorkspace(id string) string {
	return scopeWorkspace + id
}

// gitSetting names the setting that holds the key git pulls with.
const gitSetting = "keys.git"

// Local is the key folder: the files of every key, flat, the rows of
// what each is, and the setting of the key git uses. The folder is
// 0700, every file in it 0600.
type Local struct {
	Dir      string
	Rows     Rows
	Settings Settings
	// Tmp is where agents put their sockets, "" for the system's.
	Tmp string

	mu sync.Mutex
}

var _ Store = (*Local)(nil)

// Errors of the store.
var (
	ErrName    = errors.New("a key name is lower case letters, digits, dashes and underscores, up to 64, starting with a letter or a digit")
	ErrPurpose = errors.New("a key is an encryption key, encrypt, or an SSH key, ssh")
	ErrEncrypt = errors.New("an encryption key is ed25519 or rsa")
	ErrNotSSH  = errors.New("an encryption key never authenticates: it serves no run, no git and no login")
	ErrExists  = storage.ErrKeyExists
	// ErrBackup refuses a rotation of a key that holds a backup still.
	ErrBackup = errors.New("the key holds a backup from its last rotation: delete the backup first")
	// ErrNoBackup reports a key without a backup to delete.
	ErrNoBackup = errors.New("the key holds no backup")
)

// backupDir is the folder of the key folder that holds the backups.
const backupDir = "backup"

// BackupName names the backup of a key: backup/<name>.
func BackupName(name string) string {
	return backupDir + "/" + name
}

// backupPath returns the path of a file of a key's backup.
func (l *Local) backupPath(name, ext string) string {
	return filepath.Join(l.Dir, backupDir, name+ext)
}

// base is what a key's name holds before its algorithm.
var base = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// tag is the algorithm and size a name ends with: ed25519, ecdsa256,
// rsa4096.
func tag(algo string, bits int) string {
	if algo == "ed25519" || bits == 0 {
		return algo
	}
	return algo + strconv.Itoa(bits)
}

// Stem returns a key name of a base, an underscore and the tag of an algorithm and a size.
// A base that ends with the tag already keeps it once.
func Stem(name, algo string, bits int) (string, error) {
	name = strings.TrimSpace(name)
	t := tag(algo, bits)
	name = strings.TrimSuffix(name, "_"+t)
	if !base.MatchString(name) {
		return "", ErrName
	}
	return name + "_" + t, nil
}

// baseOf returns a key name without its tag.
func baseOf(row *model.KeyInfo) string {
	return strings.TrimSuffix(row.Name, "_"+tag(row.Algo, int(row.Bits)))
}

// path returns the path of a file of a key.
func (l *Local) path(name, ext string) string {
	return filepath.Join(l.Dir, name+ext)
}

// PublicPath returns a key's public file, for ssh -i while the run's agent holds the private half.
func (l *Local) PublicPath(name string) string {
	return l.path(name, ".pub")
}

// KnownHosts returns the folder's known_hosts file.
func (l *Local) KnownHosts() string {
	return filepath.Join(l.Dir, "known_hosts")
}

// List returns every key, by name, with its links and its public key.
func (l *Local) List(ctx context.Context) ([]Key, error) {
	rows, err := l.Rows.List(ctx)
	if err != nil {
		return nil, err
	}
	roles, users, err := l.Rows.Links(ctx)
	if err != nil {
		return nil, err
	}
	git := l.gitName(ctx)
	keys := make([]Key, 0, len(rows))
	for i := range rows {
		keys = append(keys, l.key(&rows[i], roles[rows[i].Name], users[rows[i].Name], git))
	}
	return keys, nil
}

// key returns the view of a row.
func (l *Local) key(row *model.KeyInfo, roles, users []string, git string) Key {
	k := Key{
		Name: row.Name, Purpose: row.Purpose, Scope: row.Scope, Algo: row.Algo, Bits: int(row.Bits),
		Fingerprint: row.Fingerprint, Comment: row.Comment, Roles: roles, Users: users, Git: row.Name == git,
		Backup: row.Backup, BackupFingerprint: row.BackupFingerprint, RotatedAt: row.RotatedAt,
	}
	if k.Roles == nil {
		k.Roles = []string{}
	}
	if k.Users == nil {
		k.Users = []string{}
	}
	if row.CreatedAt != nil {
		k.CreatedAt = *row.CreatedAt
	}
	if data, err := os.ReadFile(l.path(row.Name, ".pub")); err == nil {
		k.PublicKey = strings.TrimSpace(string(data))
	}
	return k
}

// Get returns a key by name.
func (l *Local) Get(ctx context.Context, name string) (*Key, error) {
	row, err := l.Rows.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	roles, users, err := l.Rows.Links(ctx)
	if err != nil {
		return nil, err
	}
	key := l.key(row, roles[name], users[name], l.gitName(ctx))
	return &key, nil
}

// Of returns the keys of a scope with a purpose; empty purpose is both.
func (l *Local) Of(ctx context.Context, scope, purpose string) ([]Key, error) {
	all, err := l.List(ctx)
	if err != nil {
		return nil, err
	}
	var keys []Key
	for _, key := range all {
		if key.Scope == scope && (purpose == "" || key.Purpose == purpose) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// EncryptKey returns the encryption key of a scope, nil for none.
func (l *Local) EncryptKey(ctx context.Context, scope string) *Key {
	keys, _ := l.Of(ctx, scope, PurposeEncrypt)
	if len(keys) == 0 {
		return nil
	}
	return &keys[0]
}

// check refuses a spec: a name, a purpose, an algorithm encryption
// takes.
func check(spec *Spec) error {
	if spec.Purpose == "" {
		spec.Purpose = PurposeSSH
	}
	if spec.Purpose != PurposeEncrypt && spec.Purpose != PurposeSSH {
		return ErrPurpose
	}
	if spec.Scope == "" {
		spec.Scope = storage.ScopeInstance
	}
	if spec.Scope != storage.ScopeInstance && !strings.HasPrefix(spec.Scope, scopeWorkspace) {
		return fmt.Errorf("a key's scope is instance or workspace:<id>, not %q", spec.Scope)
	}
	if spec.Purpose == PurposeEncrypt && spec.Algo != "ed25519" && spec.Algo != "rsa" {
		return ErrEncrypt
	}
	return nil
}

// Generate makes a key: ed25519 by default, ecdsa of 256, 384 or 521
// bits, or rsa of 2048, 3072 or 4096. An encryption key is ed25519 or
// rsa, and a scope holds one.
func (l *Local) Generate(ctx context.Context, spec Spec) (*Key, error) {
	if spec.Algo == "" {
		spec.Algo = "ed25519"
	}
	private, bits, err := generate(spec.Algo, spec.Bits)
	if err != nil {
		return nil, err
	}
	spec.Bits = bits
	privatePEM, _, _, err := encode(private, spec.Comment)
	if err != nil {
		return nil, err
	}
	return l.add(ctx, spec, privatePEM, private)
}

// encode returns the PEM of a private key, its public half and that
// half's fingerprint.
func encode(private any, comment string) ([]byte, ssh.PublicKey, string, error) {
	block, err := ssh.MarshalPrivateKey(private, comment)
	if err != nil {
		return nil, nil, "", err
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return nil, nil, "", err
	}
	return pem.EncodeToMemory(block), signer.PublicKey(), ssh.FingerprintSHA256(signer.PublicKey()), nil
}

// Import takes a private key in PEM or OpenSSH form, its algorithm and
// size read from it. A key held already, under any name or purpose, is
// ErrExists: a key has one purpose.
func (l *Local) Import(ctx context.Context, spec Spec, data []byte) (*Key, error) {
	data = []byte(strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n")) + "\n")
	private, err := ssh.ParseRawPrivateKey(data)
	if err != nil {
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			return nil, sshkey.ErrPassphrase
		}
		return nil, fmt.Errorf("not a private key: %w", err)
	}
	if key, ok := private.(*ed25519.PrivateKey); ok {
		private = *key
	}
	if spec.Algo, spec.Bits, err = kindOf(private); err != nil {
		return nil, err
	}
	return l.add(ctx, spec, data, private)
}

// add writes a key's files and its row.
func (l *Local) add(ctx context.Context, spec Spec, privatePEM []byte, private any) (*Key, error) {
	if err := check(&spec); err != nil {
		return nil, err
	}
	name, err := Stem(spec.Name, spec.Algo, spec.Bits)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.free(ctx, strings.TrimSuffix(name, "_"+tag(spec.Algo, spec.Bits)), spec); err != nil {
		return nil, err
	}
	if _, err := l.Rows.ByFingerprint(ctx, ssh.FingerprintSHA256(signer.PublicKey())); err == nil {
		return nil, fmt.Errorf("%w (%s)", ErrExists, ssh.FingerprintSHA256(signer.PublicKey()))
	}
	if err := l.write(name, privatePEM, signer.PublicKey(), spec.Comment); err != nil {
		return nil, err
	}
	row := &model.KeyInfo{Name: name, Purpose: spec.Purpose, Algo: spec.Algo, Bits: int64(spec.Bits), Fingerprint: ssh.FingerprintSHA256(signer.PublicKey()), Scope: spec.Scope, Comment: spec.Comment}
	if err := l.Rows.Create(ctx, row); err != nil {
		_ = l.retire(name)
		return nil, err
	}
	return l.Get(ctx, name)
}

// free refuses a base another key holds, of any algorithm, and a
// second encryption key in a scope.
func (l *Local) free(ctx context.Context, name string, spec Spec) error {
	rows, err := l.Rows.List(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		if baseOf(&rows[i]) == name {
			return fmt.Errorf("%w: %s", ErrExists, rows[i].Name)
		}
		if spec.Purpose == PurposeEncrypt && rows[i].Purpose == PurposeEncrypt && rows[i].Scope == spec.Scope {
			return fmt.Errorf("%w: %s encrypts the %s secrets; rotate it instead", ErrExists, rows[i].Name, spec.Scope)
		}
	}
	return nil
}

// write writes a key's two files, 0600, into the folder, 0700.
func (l *Local) write(name string, privatePEM []byte, pub ssh.PublicKey, comment string) error {
	if err := os.MkdirAll(l.Dir, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(l.path(name, ".key"), privatePEM, 0o600); err != nil {
		return err
	}
	public := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))
	if comment != "" {
		public += " " + comment
	}
	return os.WriteFile(l.path(name, ".pub"), []byte(public+"\n"), 0o600)
}

// retire moves a key's files aside, as <name>.key.bak-<time>: a backup
// still decrypts what the key encrypted, and nothing lists it.
func (l *Local) retire(name string) error {
	stamp := ".bak-" + time.Now().UTC().Format("20060102T150405")
	suffix := stamp
	for n := 2; exists(l.path(name, ".key"+suffix)); n++ {
		suffix = fmt.Sprintf("%s-%d", stamp, n)
	}
	for _, ext := range []string{".key", ".pub"} {
		if err := os.Rename(l.path(name, ext), l.path(name, ext+suffix)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// exists reports a file.
func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// Remove retires a key's files and drops its row and links; the git
// key setting goes with it.
func (l *Local) Remove(ctx context.Context, name string) error {
	if _, err := l.Rows.Get(ctx, name); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.retire(name); err != nil {
		return err
	}
	if l.gitName(ctx) == name {
		if err := l.Settings.Set(ctx, map[string]string{gitSetting: ""}); err != nil {
			return err
		}
	}
	return l.Rows.Delete(ctx, name)
}

// Link sets the roles and users whose runs get an SSH key.
func (l *Local) Link(ctx context.Context, name string, roles, users []string) error {
	row, err := l.Rows.Get(ctx, name)
	if err != nil {
		return err
	}
	if row.Purpose != PurposeSSH && (len(roles) > 0 || len(users) > 0) {
		return ErrNotSSH
	}
	return l.Rows.Link(ctx, name, roles, users)
}

// Private returns a key's private half, for the store's own use: an
// agent, a vault, a recipient's encrypted copy. Nothing shows it.
func (l *Local) Private(ctx context.Context, name string) ([]byte, error) {
	if _, err := l.Rows.Get(ctx, name); err != nil {
		return nil, err
	}
	return os.ReadFile(l.path(name, ".key"))
}

// identities returns a key's private half and its backups', current
// first, then the backup of its last rotation, then older ones.
func (l *Local) identities(name string) [][]byte {
	var result [][]byte
	for _, path := range []string{l.path(name, ".key"), l.backupPath(name, ".key")} {
		if data, err := os.ReadFile(path); err == nil {
			result = append(result, data)
		}
	}
	backups, _ := filepath.Glob(l.path(name, ".key.bak-*"))
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))
	for _, path := range backups {
		if data, err := os.ReadFile(path); err == nil {
			result = append(result, data)
		}
	}
	return result
}

// RemoveScope retires every key of a scope, when its workspace goes.
func (l *Local) RemoveScope(ctx context.Context, scope string) error {
	keys, err := l.Of(ctx, scope, "")
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := l.Remove(ctx, key.Name); err != nil {
			return err
		}
	}
	return nil
}
