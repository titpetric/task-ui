package frontend_test

import (
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
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
	_, err = f.opts.Board.Post(ctx, admin, first.ID, "and push it")
	assert.NoError(t, err)
	_, err = f.opts.Board.Post(ctx, admin, "", "@ops a plain note")
	assert.NoError(t, err)

	body := f.page(t, asOps, "/notes")
	assert.Contains(t, body, `href="/notes" class="active"`)
	assert.Contains(t, body, `<span class="mention">@claude</span> update <code>weather.php</code>`)
	assert.Contains(t, body, `class="badge note-state open">open</span>`)
	assert.Contains(t, body, "and push it")
	assert.Contains(t, body, `data-delete="/api/notes/`+first.ID+`"`)
	assert.True(t, strings.Index(body, "a plain note") < strings.Index(body, "weather.php"), "newest first")

	viewer := f.page(t, asViewer, "/notes")
	assert.NotContains(t, viewer, `data-delete="/api/notes/`+first.ID+`"`)
}
