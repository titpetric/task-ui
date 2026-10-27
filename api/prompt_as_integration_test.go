//go:build integration

package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A prompt runs as the user an admin names with user: the thread and
// the run are theirs. A runner names only themself (403), a name nobody
// holds is a 404, a user without prompt/edit or without sight of the
// place a 403 naming them, and a reply keeps the thread's user: it takes
// no user (400), an admin's reply runs as the thread's user, and anyone
// else's but the user's own answers 403. A prompt about a run and
// the Workflows page's prompt take the user the same way.
func TestPrompts_RunAs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	promptCommand(t, f)
	ctx := t.Context()

	res := f.call(t, asOps, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "hello", User: asAdmin})
	assert.Equal(t, http.StatusForbidden, res.Status, string(res.Raw))
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "hello", User: "nobody"})
	assert.Equal(t, http.StatusNotFound, res.Status, string(res.Raw))
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "hello", User: asOps})
	assert.Equal(t, http.StatusForbidden, res.Status, string(res.Raw))
	assert.Equal(t, "ops may not prompt in infra", res.str("error"), "without prompt/edit")

	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asOps, "prompt/edit", "allow"))
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, "user", asLimited, "prompt/edit", "allow"))
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "private", Prompt: "hello", User: asLimited})
	assert.Equal(t, http.StatusForbidden, res.Status, string(res.Raw))
	assert.Equal(t, "limited may not prompt in private", res.str("error"), "without sight of the folder")
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "hello", User: " admin "})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, asAdmin, res.str("user_name"), "the caller's own name")
	assert.Equal(t, model.NoteDone, f.answered(t, res.str("id")).State)

	asked := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Folder: "infra", Prompt: "what runs here", User: asOps})
	assert.Equal(t, http.StatusCreated, asked.Status, string(asked.Raw))
	assert.Equal(t, asOps, asked.str("user_name"), "the thread is the picked user's")
	assert.Equal(t, f.users[asOps].ID, asked.str("user_id"))
	thread := f.answered(t, asked.str("id"))
	assert.Equal(t, model.NoteDone, thread.State)
	assert.Equal(t, asOps, f.finished(t, thread.RunID).UserName, "the run is the picked user's")

	// A reply keeps the thread's user.
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Parent: thread.ID, Prompt: "and then", User: asAdmin})
	assert.Equal(t, http.StatusBadRequest, res.Status, string(res.Raw))
	assert.Equal(t, "a thread keeps its user", res.str("error"))
	res = f.call(t, asLimited, http.MethodPost, "/prompts", api.PromptRequest{Parent: thread.ID, Prompt: "and then"})
	assert.Equal(t, http.StatusForbidden, res.Status, string(res.Raw))
	assert.Equal(t, "a thread keeps its user: ops or an admin replies", res.str("error"))
	reply := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Parent: thread.ID, Prompt: "and then"})
	assert.Equal(t, http.StatusCreated, reply.Status, string(reply.Raw))
	assert.Equal(t, asOps, reply.str("user_name"), "an admin's reply goes as the thread's user")
	assert.Equal(t, model.NoteDone, f.answered(t, thread.ID).State)
	own := f.call(t, asOps, http.MethodPost, "/prompts", api.PromptRequest{Parent: thread.ID, Prompt: "once more"})
	assert.Equal(t, http.StatusCreated, own.Status, string(own.Raw))
	assert.Equal(t, asOps, own.str("user_name"))
	assert.Equal(t, model.NoteDone, f.answered(t, thread.ID).State)

	// About a run, and on the Workflows page.
	run := f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: "infra", Job: "plan"})
	assert.Equal(t, http.StatusCreated, run.Status, string(run.Raw))
	f.finished(t, run.str("id"))
	about := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: run.str("id"), Prompt: "why", User: asOps})
	assert.Equal(t, http.StatusCreated, about.Status, string(about.Raw))
	assert.Equal(t, asOps, about.str("user_name"))
	assert.Equal(t, asOps, f.finished(t, f.answered(t, about.str("id")).RunID).UserName)
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Run: run.str("id"), Prompt: "why", User: asLimited})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw), "limited sees infra")
	assert.Equal(t, asLimited, res.str("user_name"))
	f.answered(t, res.str("id"))
	area := f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Area: model.AreaWorkflows, Folder: "infra", Prompt: "build one", User: asOps})
	assert.Equal(t, http.StatusCreated, area.Status, string(area.Raw))
	assert.Equal(t, asOps, area.str("user_name"))
	f.answered(t, area.str("id"))
	res = f.call(t, asAdmin, http.MethodPost, "/prompts", api.PromptRequest{Area: model.AreaWorkflows, Prompt: "build one", User: asLimited})
	assert.Equal(t, http.StatusForbidden, res.Status, string(res.Raw))
	assert.Equal(t, "limited may not prompt in internal", res.str("error"), "limited does not see the root")
}
