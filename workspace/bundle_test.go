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

const manifest = `{"format":3,"workspace":{"folder":"acme"}}`

// passphrase is long enough for an export.
const passphrase = "correct horse battery"

// entries lists the names of a bundle's archive, decrypted.
func entries(t *testing.T, data []byte, passphrase string) []string {
	t.Helper()
	identity, err := age.NewScryptIdentity(passphrase)
	assert.NoError(t, err)
	decrypted, err := age.Decrypt(bytes.NewReader(data), identity)
	assert.NoError(t, err)
	zipped, err := gzip.NewReader(decrypted)
	assert.NoError(t, err)
	archive := tar.NewReader(zipped)
	var names []string
	for {
		header, err := archive.Next()
		if err != nil {
			break
		}
		names = append(names, header.Name)
	}
	return names
}

// A bundle reads back as written with its passphrase, and with no
// other: the manifest with its secrets in plain text, and the SSH keys
// under .ssh/, nothing else.
func TestBundle(t *testing.T) {
	t.Parallel()
	bundle := &workspace.Bundle{
		Manifest: workspace.Manifest{
			Format: workspace.Format, Version: "test",
			Workspace:   model.Workspace{Folder: "acme", Title: "Acme"},
			Projects:    []model.Project{{Folder: "api", SourceURL: "file:///api.git"}},
			HomeProject: "api",
			Secrets:     []model.Secret{{Scope: "workspace:", Name: "TOKEN", Content: "plain value"}},
		},
		Keys:       map[string][]byte{"workspace_acme_deploy_ed25519.key": []byte("private"), "workspace_acme_deploy_ed25519.pub": []byte("public")},
		WorkFactor: fast,
	}
	var out bytes.Buffer
	assert.NoError(t, bundle.Write(&out, passphrase))
	assert.Equal(t, []string{"manifest.json", ".ssh/workspace_acme_deploy_ed25519.key", ".ssh/workspace_acme_deploy_ed25519.pub"}, entries(t, out.Bytes(), passphrase))
	assert.False(t, bytes.Contains(out.Bytes(), []byte("plain value")), "the secrets are plain inside the encryption alone")
	assert.ErrorIs(t, bundle.Write(&bytes.Buffer{}, ""), workspace.ErrWeakPassphrase)
	assert.ErrorIs(t, bundle.Write(&bytes.Buffer{}, "short"), workspace.ErrWeakPassphrase)

	got, err := workspace.ReadBundle(bytes.NewReader(out.Bytes()), passphrase)
	assert.NoError(t, err)
	assert.Equal(t, "Acme", got.Manifest.Workspace.Title)
	assert.Equal(t, "api", got.Manifest.Projects[0].Folder)
	assert.Equal(t, "api", got.Manifest.HomeProject, "the home project goes by its folder")
	assert.Equal(t, "plain value", got.Manifest.Secrets[0].Content)
	assert.Equal(t, "private", string(got.Keys["workspace_acme_deploy_ed25519.key"]))
	assert.Equal(t, "public", string(got.Keys["workspace_acme_deploy_ed25519.pub"]))

	_, err = workspace.ReadBundle(bytes.NewReader(out.Bytes()), "wrong passphrase")
	assert.ErrorIs(t, err, workspace.ErrPassphrase)
	_, err = workspace.ReadBundle(strings.NewReader("not age"), passphrase)
	assert.ErrorIs(t, err, workspace.ErrBundle)
}

// A passphrase takes MinPassphrase characters, counted as runes.
func TestCheckPassphrase(t *testing.T) {
	t.Parallel()
	assert.NoError(t, workspace.CheckPassphrase(strings.Repeat("x", workspace.MinPassphrase)))
	assert.NoError(t, workspace.CheckPassphrase("abcd-efgh-ijkl-mnop-qrst-uvwx"))
	assert.ErrorIs(t, workspace.CheckPassphrase(strings.Repeat("x", workspace.MinPassphrase-1)), workspace.ErrWeakPassphrase)
	assert.ErrorIs(t, workspace.CheckPassphrase(strings.Repeat("č", workspace.MinPassphrase-1)), workspace.ErrWeakPassphrase)
}

// A bundle with an entry outside manifest.json and .ssh/, a link, a
// key file over its limit, no manifest or another format fails.
func TestReadBundle_Refused(t *testing.T) {
	t.Parallel()
	for name, entries := range map[string]map[string]string{
		"traversal":        {"manifest.json": manifest, "../evil": "x"},
		"traversal inside": {"manifest.json": manifest, ".ssh/../../evil.key": "x"},
		"absolute":         {"manifest.json": manifest, "/etc/evil": "x"},
		"other dir":        {"manifest.json": manifest, "keys/ssh/id_ed25519.key": "x"},
		"nested":           {"manifest.json": manifest, ".ssh/a/id_ed25519.key": "x"},
		"dot file":         {"manifest.json": manifest, ".ssh/.hidden.key": "x"},
		"no extension":     {"manifest.json": manifest, ".ssh/id_ed25519": "x"},
		"bare extension":   {"manifest.json": manifest, ".ssh/.key": "x"},
		"symlink":          {"manifest.json": manifest, ".ssh/id_ed25519.key@": "/etc/shadow"},
		"key over limit":   {"manifest.json": manifest, ".ssh/id_ed25519.key": strings.Repeat("x", workspace.MaxKeyFile+1)},
		"no manifest":      {".ssh/id_ed25519.key": "x"},
		"format":           {"manifest.json": `{"format":2}`},
	} {
		_, err := workspace.ReadBundle(bytes.NewReader(raw(t, "pass", entries)), "pass")
		assert.True(t, errors.Is(err, workspace.ErrBundle), name, err)
	}
	ok, err := workspace.ReadBundle(bytes.NewReader(raw(t, "pass", map[string]string{"manifest.json": manifest, ".ssh/id_ed25519.key": "x"})), "pass")
	assert.NoError(t, err)
	assert.Equal(t, "x", string(ok.Keys["id_ed25519.key"]))
}

// ReadBundle refuses a file over MaxBundle before reading it through.
func TestReadBundle_TooLarge(t *testing.T) {
	t.Parallel()
	big := raw(t, "pass", map[string]string{"manifest.json": manifest, ".ssh/id_ed25519.key": strings.Repeat("x", 1000)})
	padded := append(big, bytes.Repeat([]byte{0}, workspace.MaxBundle)...)
	_, err := workspace.ReadBundle(bytes.NewReader(padded), "pass")
	assert.True(t, errors.Is(err, workspace.ErrBundle), err)
}
