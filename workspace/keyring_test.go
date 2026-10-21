package workspace_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/workspace"
)

// mode returns a file's permission bits.
func mode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	assert.NoError(t, err)
	return info.Mode().Perm()
}

// Ensure makes the directories private and the encryption key once; a
// second call keeps it. Every key file is 0600.
func TestKeyring_Ensure(t *testing.T) {
	t.Parallel()
	k := &workspace.Keyring{Data: t.TempDir()}
	assert.Nil(t, k.EncryptKey("01ws"))

	made, err := k.Ensure("01ws", "task-ui malta")
	assert.NoError(t, err)
	assert.True(t, made)
	key := k.EncryptKey("01ws")
	assert.NotNil(t, key)
	assert.Equal(t, "ssh-ed25519", key.Type)
	assert.Equal(t, filepath.Join(k.Data, "workspaces", "01ws"), k.Dir("01ws"))

	made, err = k.Ensure("01ws", "again")
	assert.NoError(t, err)
	assert.False(t, made)
	assert.Equal(t, key.Fingerprint, k.EncryptKey("01ws").Fingerprint)

	for _, dir := range []string{k.Dir("01ws"), k.Encrypt("01ws").Dir, k.SSH("01ws").Dir} {
		assert.Equal(t, os.FileMode(0o700), mode(t, dir), dir)
	}
	ssh, err := k.SSH("01ws").GenerateAs("ed25519", "deploy", "deploy")
	assert.NoError(t, err)
	for _, path := range []string{
		filepath.Join(k.Encrypt("01ws").Dir, "id_ed25519"), filepath.Join(k.Encrypt("01ws").Dir, "id_ed25519.pub"),
		filepath.Join(k.SSH("01ws").Dir, ssh.Name), filepath.Join(k.SSH("01ws").Dir, ssh.Name+".pub"),
	} {
		assert.Equal(t, os.FileMode(0o600), mode(t, path), path)
	}

	_, err = k.Ensure("../x", "")
	assert.Error(t, err)
	assert.Error(t, k.Remove(""))
	assert.NoError(t, k.Remove("01ws"))
	assert.NoFileExists(t, k.Dir("01ws"))
}

// A rotation or an upload keeps the previous encryption key as a
// backup; an upload takes ed25519 only.
func TestKeyring_RotateEncrypt(t *testing.T) {
	t.Parallel()
	k := &workspace.Keyring{Data: t.TempDir()}
	_, err := k.Ensure("01ws", "first")
	assert.NoError(t, err)
	first := k.EncryptKey("01ws")

	rotated, err := k.RotateEncrypt("01ws", "second")
	assert.NoError(t, err)
	assert.NotEqual(t, first.Fingerprint, rotated.Fingerprint)
	assert.Equal(t, rotated.Fingerprint, k.EncryptKey("01ws").Fingerprint)
	assert.Len(t, k.Encrypt("01ws").List(), 2)

	other := &workspace.Keyring{Data: t.TempDir()}
	upload, err := other.RotateEncrypt("02ws", "elsewhere")
	assert.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(other.Encrypt("02ws").Dir, "id_ed25519"))
	assert.NoError(t, err)
	uploaded, err := k.UploadEncrypt("01ws", data, "uploaded")
	assert.NoError(t, err)
	assert.Equal(t, upload.Fingerprint, uploaded.Fingerprint)
	assert.Equal(t, upload.Fingerprint, k.EncryptKey("01ws").Fingerprint)

	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	block, err := ssh.MarshalPrivateKey(ec, "")
	assert.NoError(t, err)
	_, err = k.UploadEncrypt("01ws", pem.EncodeToMemory(block), "")
	assert.ErrorIs(t, err, workspace.ErrEncryptKind)
	assert.Equal(t, upload.Fingerprint, k.EncryptKey("01ws").Fingerprint)
}

// Files reads every key file, backups among them, and Restore writes
// them back 0600 for a workspace on another instance, byte for byte;
// it refuses a workspace with keys and a name that leaves its
// directory.
func TestKeyring_FilesRestore(t *testing.T) {
	t.Parallel()
	k := &workspace.Keyring{Data: t.TempDir()}
	_, err := k.Ensure("01from", "task-ui acme")
	assert.NoError(t, err)
	_, err = k.SSH("01from").Generate("ed25519", "deploy")
	assert.NoError(t, err)
	files, err := k.Files("01from")
	assert.NoError(t, err)
	assert.Len(t, files, 4)
	assert.NotNil(t, files["encrypt/id_ed25519"])
	assert.NotNil(t, files["ssh/id_ed25519.pub"])

	other := &workspace.Keyring{Data: t.TempDir()}
	assert.NoError(t, other.Restore("01to", files))
	again, err := other.Files("01to")
	assert.NoError(t, err)
	for name, data := range files {
		assert.Equal(t, string(data), string(again[name]), name)
		assert.Equal(t, os.FileMode(0o600), mode(t, filepath.Join(other.Dir("01to"), filepath.FromSlash(name))), name)
	}
	assert.Equal(t, os.FileMode(0o700), mode(t, other.Dir("01to")))
	assert.Equal(t, k.EncryptKey("01from").Fingerprint, other.EncryptKey("01to").Fingerprint)

	assert.Error(t, other.Restore("01to", files), "keys there already")
	assert.Error(t, k.Restore("01bad", map[string][]byte{"ssh/../../x": nil}))
	assert.Error(t, k.Restore("01bad2", map[string][]byte{"other/x": nil}))
	empty, err := k.Files("01none")
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}

// Prepare makes the private key directories and no key.
func TestKeyring_Prepare(t *testing.T) {
	t.Parallel()
	k := &workspace.Keyring{Data: t.TempDir()}
	assert.NoError(t, k.Prepare("01ws"))
	assert.Nil(t, k.EncryptKey("01ws"))
	for _, dir := range []string{k.Dir("01ws"), k.Encrypt("01ws").Dir, k.SSH("01ws").Dir} {
		assert.Equal(t, os.FileMode(0o700), mode(t, dir), dir)
	}
	assert.Error(t, k.Prepare("../x"))
}
