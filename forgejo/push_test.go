package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// A push names its repository by clone URL, which keys the checkout to pull.
func TestParsePush(t *testing.T) {
	t.Parallel()
	push, err := forgejo.ParsePush([]byte(`{"ref":"refs/heads/main","repository":{"full_name":"Malta/apps","ssh_url":"ssh://git@git.cdn.si:2222/Malta/apps.git","clone_url":"https://git.cdn.si/Malta/apps.git"}}`))
	assert.NoError(t, err)
	assert.Equal(t, "refs/heads/main", push.Ref)
	assert.Equal(t, "malta/apps", push.Key())

	push, err = forgejo.ParsePush([]byte(`{"repository":{"full_name":"malta/deploy"}}`))
	assert.NoError(t, err)
	assert.Equal(t, "malta/deploy", push.Key())

	_, err = forgejo.ParsePush([]byte(`{"zen":"ping"}`))
	assert.ErrorIs(t, err, forgejo.ErrNoRepository)
	_, err = forgejo.ParsePush([]byte(`not json`))
	assert.Error(t, err)
}
