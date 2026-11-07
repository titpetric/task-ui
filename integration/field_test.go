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
