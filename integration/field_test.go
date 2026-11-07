package integration_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// A field marks a secret, and leaves out what it lacks.
func TestField(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(integration.Field{Name: "token", Label: "Token", Secret: true})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"token","label":"Token","secret":true}`, string(data))

	clone := integration.Field{Name: "clone", Options: []string{"ssh", "https"}}
	assert.True(t, clone.Valid("ssh"))
	assert.True(t, clone.Valid(""))
	assert.False(t, clone.Valid("ftp"))
	assert.True(t, integration.Field{Name: "url"}.Valid("anything"))

	// A variable's setting carries what a run gets and its switch.
	data, err = json.Marshal(integration.Field{Name: "model", Label: "Model", Env: "ANTHROPIC_MODEL", Default: "m", Toggle: true, DefaultOn: true})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"model","label":"Model","env":"ANTHROPIC_MODEL","default":"m","toggle":true,"default_on":true}`, string(data))
}

// A field a path names refuses a value that is no address: a hex
// secret pasted into a redirect URI is the mistake it catches.
func TestField_Check(t *testing.T) {
	t.Parallel()
	redirect := integration.Field{Name: "redirect_uri", Path: "/api/integrations/spotify/callback"}
	assert.NoError(t, redirect.Check(""))
	assert.NoError(t, redirect.Check("https://task.cdn.si/api/integrations/spotify/callback"))
	assert.NoError(t, redirect.Check("http://127.0.0.1:3000/api/integrations/spotify/callback"))
	for _, value := range []string{"043c12b09d494b93947f4e87d8c1edc9", "task.cdn.si/callback", "ftp://task.cdn.si/cb", "/api/integrations/spotify/callback"} {
		err := redirect.Check(value)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redirect_uri")
	}

	// A field with options still answers for its own list, and one with
	// neither takes anything.
	clone := integration.Field{Name: "clone", Options: []string{"ssh", "https"}}
	assert.NoError(t, clone.Check("ssh"))
	assert.Error(t, clone.Check("ftp"))
	assert.NoError(t, integration.Field{Name: "url"}.Check("043c12b09d494b93947f4e87d8c1edc9"))
}

// A field a path names shows the instance URL and that path, a stored
// value wins over it, and one trailing slash is not two.
func TestField_Value(t *testing.T) {
	t.Parallel()
	redirect := integration.Field{Name: "redirect_uri", Path: "/api/integrations/spotify/callback"}
	assert.Equal(t, "https://task.example.com/api/integrations/spotify/callback", redirect.Value("", "https://task.example.com"))
	assert.Equal(t, "https://task.example.com/api/integrations/spotify/callback", redirect.Value("", "https://task.example.com/"))
	assert.Equal(t, "https://other.example.com/cb", redirect.Value("https://other.example.com/cb", "https://task.example.com"))
	assert.Equal(t, "", redirect.Value("", ""))

	// Without a path the default stands, and a stored value still wins.
	model := integration.Field{Name: "model", Default: "m"}
	assert.Equal(t, "m", model.Value("", "https://task.example.com"))
	assert.Equal(t, "haiku", model.Value("haiku", ""))
}
