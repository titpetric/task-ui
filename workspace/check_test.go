package workspace_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/workspace"
)

// roles is a keyring with an instance encryption key, an instance SSH
// folder with a key, and two workspaces with their keys.
func roles(t *testing.T) (*workspace.Keyring, *sshkey.Store) {
	t.Helper()
	access := &sshkey.Store{Dir: t.TempDir()}
	_, err := access.Generate("ed25519", "access")
	assert.NoError(t, err)
	k := &workspace.Keyring{Data: t.TempDir(), Access: []*sshkey.Store{access}}
	_, err = k.EnsureInstance("instance")
	assert.NoError(t, err)
	for _, id := range []string{"01a", "01b"} {
		_, err := k.Ensure(id, id)
		assert.NoError(t, err)
		_, err = k.SSH(id).GenerateAs("ed25519", "deploy", id)
		assert.NoError(t, err)
	}
	return k, access
}

// read returns a file's bytes.
func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	return data
}

// A key for one use fails the check for the other, in both directions, by
// its private or its public half; an encryption key has one holder.
func TestKeyring_Roles(t *testing.T) {
	t.Parallel()
	k, access := roles(t)
	instance := read(t, filepath.Join(k.Instance().Dir, "id_ed25519"))
	instancePub := read(t, filepath.Join(k.Instance().Dir, "id_ed25519.pub"))
	wsEncrypt := read(t, filepath.Join(k.Encrypt("01a").Dir, "id_ed25519"))
	accessKey := read(t, filepath.Join(access.Dir, "id_ed25519"))
	wsSSH := read(t, filepath.Join(k.SSH("01b").Dir, "id_ed25519_deploy"))

	for name, data := range map[string][]byte{"instance": instance, "instance pub": instancePub, "workspace": wsEncrypt} {
		assert.True(t, errors.Is(k.CheckAccess(data), workspace.ErrRole), name)
	}
	assert.NoError(t, k.CheckAccess(accessKey), "an SSH key in two SSH stores is one use")
	assert.NoError(t, k.CheckAccess([]byte("garbage")), "no key is the upload's to refuse")

	for name, data := range map[string][]byte{"instance ssh": accessKey, "workspace ssh": wsSSH, "instance's": instance, "another workspace's": wsEncrypt} {
		assert.True(t, errors.Is(k.CheckEncrypt("01b", data), workspace.ErrRole), name)
	}
	assert.True(t, errors.Is(k.CheckEncrypt("", wsEncrypt), workspace.ErrRole), "a workspace's key for the instance")
	assert.NoError(t, k.CheckEncrypt("01a", wsEncrypt), "a workspace's own key, again")
	_, err := k.UploadEncrypt("01b", accessKey, "")
	assert.ErrorIs(t, err, workspace.ErrRole)

	assert.True(t, k.Encrypts(filepath.Join(k.Instance().Dir, "id_ed25519")))
	assert.True(t, k.Encrypts(filepath.Join(k.Encrypt("01b").Dir, "id_ed25519")))
	assert.False(t, k.Encrypts(filepath.Join(access.Dir, "id_ed25519")))
	assert.False(t, k.Encrypts(filepath.Join(access.Dir, "missing")))
	assert.Len(t, k.Overlaps(), 0)
}

// A bundle restores when its keys keep their use; its encryption key
// held by a workspace here, or an SSH key that encrypts, fails.
func TestKeyring_CheckBundle(t *testing.T) {
	t.Parallel()
	k, access := roles(t)
	files, err := k.Files("01a")
	assert.NoError(t, err)
	assert.ErrorIs(t, k.CheckBundle(files), workspace.ErrRole)
	assert.ErrorIs(t, k.Restore("01c", files), workspace.ErrRole, "acme's encryption key, a second holder")
	assert.NoDirExists(t, k.Dir("01c"))

	fresh := &workspace.Keyring{Data: t.TempDir()}
	_, err = fresh.Ensure("01x", "x")
	assert.NoError(t, err)
	bundle, err := fresh.Files("01x")
	assert.NoError(t, err)
	bundle["ssh/id_ed25519"] = read(t, filepath.Join(k.Instance().Dir, "id_ed25519"))
	assert.ErrorIs(t, k.Restore("01d", bundle), workspace.ErrRole, "an SSH key that encrypts")
	delete(bundle, "ssh/id_ed25519")
	bundle["ssh/id_ed25519_mine"] = bundle["encrypt/id_ed25519"]
	assert.ErrorIs(t, k.Restore("01e", bundle), workspace.ErrRole, "its own encryption key as an SSH key")
	delete(bundle, "ssh/id_ed25519_mine")
	bundle["ssh/id_ed25519"] = read(t, filepath.Join(access.Dir, "id_ed25519"))
	assert.NoError(t, k.Restore("01f", bundle), "an SSH key the instance has too")
}

// Overlaps names a key both an encryption store and an SSH store hold,
// by the paths of both; every file stays.
func TestKeyring_Overlaps(t *testing.T) {
	t.Parallel()
	k, access := roles(t)
	for _, name := range []string{"id_ed25519", "id_ed25519.pub"} {
		data := read(t, filepath.Join(k.Encrypt("01a").Dir, name))
		assert.NoError(t, os.WriteFile(filepath.Join(access.Dir, "id_ed25519_copy"+name[len("id_ed25519"):]), data, 0o600))
	}
	overlaps := k.Overlaps()
	assert.Len(t, overlaps, 1)
	want := workspace.Overlap{
		Fingerprint: k.EncryptKey("01a").Fingerprint,
		Encrypt:     []string{filepath.Join(k.Encrypt("01a").Dir, "id_ed25519")},
		Access:      []string{filepath.Join(access.Dir, "id_ed25519_copy")},
	}
	assert.Equal(t, want, overlaps[0])
	assert.True(t, k.Encrypts(filepath.Join(access.Dir, "id_ed25519_copy")), "the copy counts as an encryption key")
	assert.FileExists(t, filepath.Join(access.Dir, "id_ed25519_copy"))
}
