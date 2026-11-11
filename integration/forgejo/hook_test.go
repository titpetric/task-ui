package forgejo_test

import (
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// The push hook posts JSON to the URL with the secret, and a hook is
// current when it does the same: a listing without the secret counts as
// current, one that shows another secret does not.
func TestHook(t *testing.T) {
	t.Parallel()
	url := "https://task.example/api/webhooks/forgejo"
	hook := forgejo.PushHook(url, "s3")
	assert.Equal(t, "forgejo", hook.Type)
	assert.Equal(t, "json", hook.Config["content_type"])
	assert.Equal(t, []string{"push"}, hook.Events)
	assert.True(t, hook.Active)
	assert.True(t, hook.Current(url, "s3"))
	assert.False(t, hook.Current(url, "other"))

	listed := forgejo.Hook{Config: map[string]string{"url": url, "content_type": "json"}, Events: []string{"push"}, Active: true}
	assert.True(t, listed.Current(url, "anything"))
	listed.Active = false
	assert.False(t, listed.Current(url, "anything"))
	listed.Active, listed.Events = true, []string{"create"}
	assert.False(t, listed.Current(url, "anything"))
	assert.False(t, forgejo.Hook{Config: map[string]string{"url": "https://elsewhere"}, Events: []string{"push"}, Active: true}.Current(url, "s3"))
}
