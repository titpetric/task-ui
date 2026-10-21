package workspace_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/workspace"
)

// fast keeps scrypt cheap in tests.
const fast = 10

// raw encrypts a tar.gz of the entries, by name, to a passphrase, as a
// hand made bundle would be; a name ending in @ is a symlink.
func raw(t *testing.T, passphrase string, entries map[string]string) []byte {
	t.Helper()
	recipient, err := age.NewScryptRecipient(passphrase)
	assert.NoError(t, err)
	recipient.SetWorkFactor(fast)
	var out bytes.Buffer
	encrypted, err := age.Encrypt(&out, recipient)
	assert.NoError(t, err)
	zipped := gzip.NewWriter(encrypted)
	archive := tar.NewWriter(zipped)
	for name, content := range entries {
		header := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if link, ok := strings.CutSuffix(name, "@"); ok {
			header = &tar.Header{Name: link, Linkname: content, Typeflag: tar.TypeSymlink}
		}
		assert.NoError(t, archive.WriteHeader(header))
		if header.Typeflag == tar.TypeReg {
			_, err := archive.Write([]byte(content))
			assert.NoError(t, err)
		}
	}
	assert.NoError(t, archive.Close())
	assert.NoError(t, zipped.Close())
	assert.NoError(t, encrypted.Close())
	return out.Bytes()
}

const manifest = `{"format":1,"workspace":{"folder":"acme"}}`

// A bundle reads back as written with its passphrase, and with no other.
func TestBundle(t *testing.T) {
	t.Parallel()
	bundle := &workspace.Bundle{
		Manifest: workspace.Manifest{
			Format: workspace.Format, Version: "test",
			Workspace: model.Workspace{Folder: "acme", Title: "Acme"},
			Projects:  []model.Project{{Folder: "api", SourceURL: "file:///api.git"}},
			Secrets:   []model.Secret{{Scope: "workspace:", Name: "TOKEN", Content: "ciphertext"}},
		},
		Keys:       map[string][]byte{"encrypt/id_ed25519": []byte("private"), "ssh/id_ed25519_deploy.pub": []byte("public")},
		WorkFactor: fast,
	}
	var out bytes.Buffer
	assert.NoError(t, bundle.Write(&out, "correct horse"))
	assert.Error(t, bundle.Write(&bytes.Buffer{}, ""), "a passphrase is required")

	got, err := workspace.ReadBundle(bytes.NewReader(out.Bytes()), "correct horse")
	assert.NoError(t, err)
	assert.Equal(t, "Acme", got.Manifest.Workspace.Title)
	assert.Equal(t, "api", got.Manifest.Projects[0].Folder)
	assert.Equal(t, "ciphertext", got.Manifest.Secrets[0].Content)
	assert.Equal(t, "private", string(got.Keys["encrypt/id_ed25519"]))
	assert.Equal(t, "public", string(got.Keys["ssh/id_ed25519_deploy.pub"]))

	_, err = workspace.ReadBundle(bytes.NewReader(out.Bytes()), "wrong")
	assert.True(t, errors.Is(err, workspace.ErrPassphrase), err)
	_, err = workspace.ReadBundle(strings.NewReader("not age"), "correct horse")
	assert.True(t, errors.Is(err, workspace.ErrBundle), err)
}

// A bundle with an entry outside manifest.json and keys/, a link, a
// key file over its limit, no manifest or another format fails.
func TestReadBundle_Refused(t *testing.T) {
	t.Parallel()
	for name, entries := range map[string]map[string]string{
		"traversal":        {"manifest.json": manifest, "../evil": "x"},
		"traversal inside": {"manifest.json": manifest, "keys/ssh/../../evil": "x"},
		"absolute":         {"manifest.json": manifest, "/etc/evil": "x"},
		"other dir":        {"manifest.json": manifest, "keys/other/id_ed25519": "x"},
		"nested":           {"manifest.json": manifest, "keys/ssh/a/id_ed25519": "x"},
		"dot file":         {"manifest.json": manifest, "keys/ssh/.hidden": "x"},
		"symlink":          {"manifest.json": manifest, "keys/ssh/id_ed25519@": "/etc/shadow"},
		"key over limit":   {"manifest.json": manifest, "keys/ssh/id_ed25519": strings.Repeat("x", workspace.MaxKeyFile+1)},
		"no manifest":      {"keys/ssh/id_ed25519": "x"},
		"format":           {"manifest.json": `{"format":2}`},
	} {
		_, err := workspace.ReadBundle(bytes.NewReader(raw(t, "pass", entries)), "pass")
		assert.True(t, errors.Is(err, workspace.ErrBundle), name, err)
	}
	ok, err := workspace.ReadBundle(bytes.NewReader(raw(t, "pass", map[string]string{"manifest.json": manifest, "keys/ssh/id_ed25519": "x"})), "pass")
	assert.NoError(t, err)
	assert.Equal(t, "x", string(ok.Keys["ssh/id_ed25519"]))
}

// ReadBundle refuses a file over MaxBundle before reading it through.
func TestReadBundle_TooLarge(t *testing.T) {
	t.Parallel()
	big := raw(t, "pass", map[string]string{"manifest.json": manifest, "keys/ssh/id_ed25519": strings.Repeat("x", 1000)})
	padded := append(big, bytes.Repeat([]byte{0}, workspace.MaxBundle)...)
	_, err := workspace.ReadBundle(bytes.NewReader(padded), "pass")
	assert.True(t, errors.Is(err, workspace.ErrBundle), err)
}
