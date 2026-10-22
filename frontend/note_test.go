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
	assert.Contains(t, body, `<header><span class="author"><span class="avatar sm role-admin" title="admin">AD</span> <b>admin</b></span>`)
	assert.Contains(t, body, `data-delete="/api/notes/`+first.ID+`"`)
	// A reply deletes on its own, by its author or an admin: ops did not
	// write it, an admin may.
	assert.NotContains(t, body, `data-delete="/api/notes/`+reply.ID+`"`)
	assert.Contains(t, f.page(t, asAdmin, "/notes"), `data-delete="/api/notes/`+reply.ID+`" data-confirm="Delete this reply of admin?"`)
	assert.True(t, strings.Index(body, "a plain note") < strings.Index(body, "weather.php"), "newest first")

	viewer := f.page(t, asViewer, "/notes")
	assert.NotContains(t, viewer, `data-delete="/api/notes/`+first.ID+`"`)
}

// A note links the agent's run for the users who see it: the runs of the
// internal folder are the admins'.
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
	note, err := f.opts.Board.Post(ctx, f.users[asOps], "", "@claude tidy up")
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Board.Notes.Edit(ctx, note.ID, "@claude tidy up, twice"))
	answer := &model.Note{Channel: model.ChannelNotes, ParentID: note.ID, UserName: f.opts.Board.Answerer(), Body: "Tidied."}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, answer))

	hook := `data-edit="` + note.ID + `" data-body="@claude tidy up, twice"`
	ops := f.page(t, asOps, "/notes")
	assert.Contains(t, ops, hook)
	assert.Contains(t, ops, `class="faint edited"`)
	assert.Contains(t, f.page(t, asAdmin, "/notes"), hook)
	assert.NotContains(t, f.page(t, asViewer, "/notes"), hook)
	assert.NotContains(t, f.page(t, asAdmin, "/notes"), `data-edit="`+answer.ID+`"`)
}
