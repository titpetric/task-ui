package access_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestParseHtpasswd(t *testing.T) {
	t.Parallel()
	hash, err := access.HashPassword("demo")
	assert.NoError(t, err)
	apache := "$2y$" + strings.TrimPrefix(hash, "$2a$")

	creds, err := access.ParseHtpasswd(strings.NewReader("# who may sign in\n\nadmin:" + hash + "\n ops : " + apache + " \n"))
	assert.NoError(t, err)
	assert.Equal(t, []access.Credential{{Name: "admin", Hash: hash}, {Name: "ops", Hash: apache}}, creds)
	assert.True(t, access.CheckPassword(&model.User{Password: creds[1].Hash}, "demo"))

	for _, bad := range []string{"admin", "admin:", ":x", "admin:{SHA}abc", "admin:$apr1$abc", "admin:" + hash + "\nadmin:" + hash} {
		_, err := access.ParseHtpasswd(strings.NewReader(bad))
		assert.Error(t, err, bad)
	}
	_, err = access.ParseHtpasswd(strings.NewReader("ops:$apr1$abc"))
	assert.ErrorContains(t, err, "htpasswd -B")
}

func TestLoadHtpasswd(t *testing.T) {
	t.Parallel()
	hash, _ := access.HashPassword("demo")
	path := filepath.Join(t.TempDir(), ".htpasswd")
	assert.NoError(t, os.WriteFile(path, []byte("admin:"+hash+"\n"), 0o600))

	creds, err := access.LoadHtpasswd(path)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)

	_, err = access.LoadHtpasswd(filepath.Join(t.TempDir(), "missing"))
	assert.Error(t, err)
	assert.NoError(t, os.WriteFile(path, []byte("broken\n"), 0o600))
	_, err = access.LoadHtpasswd(path)
	assert.ErrorContains(t, err, ".htpasswd: line 1")
}

func TestSetHtpasswd(t *testing.T) {
	t.Parallel()
	first, _ := access.HashPassword("one")
	second, _ := access.HashPassword("two")
	path := filepath.Join(t.TempDir(), "etc", ".htpasswd")

	// The file is made, a user added, another added, the first replaced.
	assert.NoError(t, access.SetHtpasswd(path, "admin", first))
	assert.NoError(t, access.SetHtpasswd(path, " ops ", first))
	assert.NoError(t, access.SetHtpasswd(path, "admin", second))
	creds, err := access.LoadHtpasswd(path)
	assert.NoError(t, err)
	assert.Equal(t, []access.Credential{{Name: "admin", Hash: second}, {Name: "ops", Hash: first}}, creds)
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Comments and blank lines stay.
	assert.NoError(t, os.WriteFile(path, []byte("# people\n\nadmin:"+first+"\n"), 0o600))
	assert.NoError(t, access.SetHtpasswd(path, "admin", second))
	data, _ := os.ReadFile(path)
	assert.Equal(t, "# people\n\nadmin:"+second+"\n", string(data))

	assert.Error(t, access.SetHtpasswd(path, "", first))
	assert.Error(t, access.SetHtpasswd(path, "a:b", first))
	assert.Error(t, access.SetHtpasswd(path, "ops", "plain"))
}
