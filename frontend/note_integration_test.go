//go:build integration

package frontend_test

import (
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The board shows the notes newest first with their replies; a note to the
// agent carries its state, and its author or an admin may delete it.
func TestNotes(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	ops, admin := f.users[asOps], f.users[asAdmin]
	first, err := f.opts.Board.Post(ctx, ops, "", "@claude update `weather.php`")
	assert.NoError(t, err)
	reply, err := f.opts.Board.Post(ctx, admin, first.ID, "and push it")
	assert.NoError(t, err)
	_, err = f.opts.Board.Post(ctx, admin, "", "@ops a plain note")
	assert.NoError(t, err)

	body := f.page(t, asOps, "/notes")
	assert.Contains(t, body, `href="/notes" class="active"`)
	assert.Contains(t, body, `<span class="mention">@claude</span> update <code>weather.php</code>`)
	assert.Contains(t, body, `class="badge note-state open">open</span>`)
	assert.Contains(t, body, "and push it")
	assert.Contains(t, body, `<span class="avatar sm role-runner" title="ops">OP</span>`)
	assert.Contains(t, body, `<header><span class="author"><span class="user-ref"><span class="avatar sm role-admin" title="admin">AD</span> <span class="user-name">admin</span><details class="more-menu">`)
	assert.Contains(t, body, `data-delete="/api/notes/`+first.ID+`"`)
	// A reply deletes on its own, by its author or an admin: ops did not
	// write it, an admin may.
	assert.NotContains(t, body, `data-delete="/api/notes/`+reply.ID+`"`)
	assert.Contains(t, f.page(t, asAdmin, "/notes"), `data-delete="/api/notes/`+reply.ID+`" data-confirm="Delete this reply of admin?"`)
	assert.True(t, strings.Index(body, "a plain note") < strings.Index(body, "weather.php"), "newest first")

	viewer := f.page(t, asViewer, "/notes")
	assert.NotContains(t, viewer, `data-delete="/api/notes/`+first.ID+`"`)
}

// A note links the agent's run for the users who see it: the older runs
// of the internal folder are the admins', the Notes agent's take
// note/view.
func TestNotes_runLink(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	note, err := f.opts.Board.Post(ctx, f.users[asOps], "", "@claude tidy up")
	assert.NoError(t, err)
	internal := f.run(t, model.Internal, "claude-note", "claude", model.StateOK)
	assert.NoError(t, f.opts.Board.Notes.SetState(ctx, note.ID, model.NoteDone, internal.ID))
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, &model.Note{Channel: model.ChannelNotes, ParentID: note.ID, UserName: "claude", Body: "Tidied.", RunID: internal.ID}))
	other, err := f.opts.Board.Post(ctx, f.users[asOps], "", "@claude build it")
	assert.NoError(t, err)
	web := f.run(t, "web", "build", "claude", model.StateOK)
	assert.NoError(t, f.opts.Board.Notes.SetState(ctx, other.ID, model.NoteDone, web.ID))

	link := func(id string) string { return `href="/runs/` + id + `" title="The run` }
	admin := f.page(t, asAdmin, "/notes")
	assert.Equal(t, 2, strings.Count(admin, link(internal.ID)), "the note and the reply")
	assert.Contains(t, admin, link(web.ID))
	ops := f.page(t, asOps, "/notes")
	assert.NotContains(t, ops, link(internal.ID))
	assert.Contains(t, ops, "Tidied.")
	assert.Contains(t, ops, link(web.ID))
}

// A note's text carries the edit hook for its author and the admins, and
// an edited note says so; the agent's answer has none.
func TestNotes_edit(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	// claude runs instance Claude when mentioned; the note goes in as
	// one it answered, with no run to start here.
	assert.NoError(t, f.opts.Users.Create(ctx, &model.User{Name: "claude", Role: model.RoleRunner, Mcp: true, AgentReply: true}))
	ops := f.users[asOps]
	note := &model.Note{Channel: model.ChannelNotes, UserID: ops.ID, UserName: ops.Name, Body: "@claude tidy up", State: model.NoteDone}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, note))
	assert.NoError(t, f.opts.Board.Notes.Edit(ctx, note.ID, "@claude tidy up, twice"))
	answer := &model.Note{Channel: model.ChannelNotes, ParentID: note.ID, UserName: "claude", Body: "Tidied."}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, answer))

	hook := `data-edit="` + note.ID + `" data-body="@claude tidy up, twice"`
	page := f.page(t, asOps, "/notes")
	assert.Contains(t, page, hook)
	assert.Contains(t, page, `class="faint edited"`)
	assert.Contains(t, f.page(t, asAdmin, "/notes"), hook)
	assert.NotContains(t, f.page(t, asViewer, "/notes"), hook)
	assert.NotContains(t, f.page(t, asAdmin, "/notes"), `data-edit="`+answer.ID+`"`)
}

// The fields of a note list the users a mention may name: the composer, a
// reply, and the field the script opens to edit, over the roles the body
// carries; a mention of a user links their profile.
func TestNotes_mentions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	_, err := f.opts.Board.Post(context.Background(), f.users[asAdmin], "", "@ops a plain note")
	assert.NoError(t, err)

	body := f.page(t, asOps, "/notes")
	assert.Contains(t, body, `<body data-roles='{`)
	assert.Contains(t, body, `<textarea name="body" rows="3" required data-mentions placeholder=`)
	assert.Contains(t, body, `<textarea name="body" rows="2" required data-mentions aria-label="A reply">`)
	assert.Contains(t, body, `<a class="mention" href="/users/ops">@ops</a> a plain note`)
	assert.Contains(t, f.get(t, asOps, "/static/task-ui.js").Body, `<textarea name="body" rows="3" required data-mentions aria-label="The note">`)
}
