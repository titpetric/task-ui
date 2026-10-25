package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// A repository clones over ssh, or https when asked, and names its owner.
func TestRepo(t *testing.T) {
	t.Parallel()
	r := repo("Malta", "apps", "main", false)
	assert.Equal(t, "Malta", r.Owner())
	assert.Equal(t, "ssh://git@git.example:2222/Malta/apps.git", r.URL(false))
	assert.Equal(t, "https://git.example/Malta/apps.git", r.URL(true))
}

// Clone URLs of every scheme compare by owner/name, in lower case.
func TestKey(t *testing.T) {
	t.Parallel()
	for _, url := range []string{
		"ssh://git@git.cdn.si:2222/Malta/apps.git",
		"https://git.cdn.si/malta/apps.git",
		"https://git.cdn.si/malta/apps",
		"git@git.cdn.si:malta/apps.git",
		"malta/apps",
		"  malta/apps/ ",
	} {
		assert.Equal(t, "malta/apps", forgejo.Key(url), url)
	}
	assert.Equal(t, "", forgejo.Key("apps"))
	assert.Equal(t, "", forgejo.Key("ssh://git@git.cdn.si/apps"))
	assert.Equal(t, "", forgejo.Key(""))
}
