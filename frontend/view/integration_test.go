package view_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A form shows the plain settings, and of a secret only whether it has a value.
func TestIntegration(t *testing.T) {
	t.Parallel()
	kind := *integration.KindOf(integration.Beszel)
	var got view.Integration = view.NewIntegration(kind, map[string]string{"url": "http://beszel:8090", "token": "t1"})
	assert.True(t, got.Configured)
	assert.Equal(t, "http://beszel:8090", got.Values["url"])
	assert.True(t, got.Set["token"])
	_, leaked := got.Values["token"]
	assert.False(t, leaked)
	assert.False(t, view.NewIntegration(kind, nil).Configured)

	// OpenRouter's extra variables have no switch: the key sets each.
	router := view.NewIntegration(*integration.KindOf(integration.OpenRouter), map[string]string{"token": "sk-or-1"})
	assert.True(t, router.Configured)
	assert.True(t, router.Set["token"])
	assert.Empty(t, router.Enabled)
}

// The page is a card per integration on a rail: the form and its
// variables, the steps to its key folded under it, in order; the Claude
// card's Who lists the users
// with the role they act with, the roles with their counts, and the
// summary line the script fills. Nothing of a secret shows.
func TestIntegration_Page(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := admin(t)
	p.Users = []model.User{{Name: "admin", Role: model.RoleAdmin}, {Name: "ops", Role: model.RoleRunner, Admin: true}, {Name: "dev", Role: model.RoleRunner}, {Name: "guest", Role: model.RoleViewer}}
	for _, kind := range integration.Kinds() {
		values := map[string]string{}
		if kind.ID == integration.Claude {
			values = map[string]string{"token": "sk-secret", "who": "admin"}
		}
		p.Integrations = append(p.Integrations, view.NewIntegration(kind, values))
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "integrations", p))
	body := buf.String()
	assert.NotContains(t, body, "sk-secret")
	assert.Contains(t, body, `<ol class="wf-rail">`)
	assert.Contains(t, body, `data-cards-fold aria-pressed="false">Expand all</button>`)
	for _, kind := range integration.Kinds() {
		assert.Contains(t, body, `<li class="wf-step integration-step kind-`+kind.ID+`" id="`+kind.ID+`">`)
		assert.Contains(t, body, `data-api="/api/integrations/`+kind.ID+`"`)
		assert.Contains(t, body, `data-card-toggle aria-expanded="false"`)
		assert.Contains(t, body, `type="submit" disabled>Save</button>`)
		last := 0
		for i, step := range kind.Steps {
			at := strings.Index(body, `<li>`+string(view.Message(step))+`</li>`)
			assert.True(t, at > last, kind.ID, i)
			last = at
		}
	}
	assert.Contains(t, body, `<summary>Get the OAuth token</summary>`)
	assert.Contains(t, body, `<summary>Get the API key</summary>`)
	assert.Contains(t, body, `<summary>Get the token</summary>`)
	assert.Contains(t, body, `class="table-wrap integration-env"`)
	assert.Contains(t, body, `<b class="mono">ANTHROPIC_DEFAULT_OPUS_MODEL</b>`)
	assert.Contains(t, body, `name="values.default_opus_model"`)
	assert.NotContains(t, body, `values.model.enabled`)
	assert.Contains(t, body, `value="deepseek/deepseek-v4.1-flash-1017"`)
	assert.Contains(t, body, `value="https://openrouter.ai/api"`)
	assert.Contains(t, body, "<li><p>On a machine where you are signed in to Claude Code, run <code>claude setup-token</code>.</p>\n</li>")
	assert.Contains(t, body, `<div class="whofield" data-who-empty="every run gets the token">`)
	assert.Contains(t, body, `value="admin" data-type="joined" checked> admin <small class="count">· 2</small></label>`)
	assert.Contains(t, body, `value="runner" data-type="joined"> runner <small class="count">· 1</small></label>`)
	assert.Contains(t, body, `<label class="check who-user" data-role="admin"><input type="checkbox" name="values.who" value="ops" data-type="joined"> ops <small class="role">admin</small> <small class="who-state"></small></label>`)
	assert.Contains(t, body, `<label class="check who-user" data-role="runner"><input type="checkbox" name="values.who" value="dev" data-type="joined"> dev <small class="role">runner</small>`)
	assert.Contains(t, body, `<p class="dim who-summary"></p>`)
	assert.Contains(t, body, `<span class="badge ok">set</span>`)
	assert.Contains(t, body, `<span class="badge ok">set up</span>`)
	assert.Contains(t, body, `data-confirm-ok="Clear"`)
	assert.Contains(t, body, "No sync ran yet")
	assert.Contains(t, body, "The instance has no public URL")
	assert.NotContains(t, body, `data-job="@forgejo-sync"`, "no sync while Forgejo is not set up")
	// Only the Claude kind has a Who field.
	assert.Equal(t, 1, strings.Count(body, `class="dim who-summary"`))
}
