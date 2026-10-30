package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/storage"
)

// Dispatcher starts the Claude runs a module asks for. It is what
// `Task::dispatch($comment, "action/refinement.md")` reaches.
//
// A dispatch is an ordinary prompt: the same board, the same run, the
// same artifact tree and the same MCP server a prompt from a page gets.
// Nothing about a module's run is special, which is why it appears on
// /runs and its answer arrives as a note reply.
type Dispatcher struct {
	Board *note.Board
	Users *storage.UserStorage
	Store *artifactstore.Store
}

// NewDispatcher returns the dispatcher of an instance.
func NewDispatcher(board *note.Board, users *storage.UserStorage, store *artifactstore.Store) *Dispatcher {
	return &Dispatcher{Board: board, Users: users, Store: store}
}

// Dispatch starts a run over a comment and the action file naming what to
// do with it, and returns the identifier of the note that holds the
// thread. The run is started in the background: the note exists when this
// returns and Claude's reply arrives under it.
//
// The user is the one whose request reached the module, so a module's run
// has the permissions of whoever clicked rather than the instance's: a
// viewer's comment starts nothing, and the run reads the secrets and keys
// its own user may.
func (d *Dispatcher) Dispatch(ctx context.Context, m *phpscript.Module, comment, action, folder string) (string, error) {
	if d.Board == nil {
		return "", errors.New("this instance starts no module runs")
	}
	user := access.UserFrom(ctx)
	if user == nil {
		return "", errors.New("a module run needs the user who asked for it")
	}
	body, err := d.prompt(m, comment, action, folder)
	if err != nil {
		return "", err
	}
	// model.Default is the root folder, where a prompt with no folder of
	// its own runs: a module is not a checkout and has no folder.
	asked, err := d.Board.Ask(ctx, user, "", model.Default, body)
	if err != nil {
		return "", err
	}
	return asked.ID, nil
}

// prompt is the action file, then where to work, then the comment, then
// the index of what is in the folder already.
//
// Nothing else is added. The run's own environment already names the
// artifact tree, the MCP server and what the user may do, and the agent's
// context repeats them, so a module's prompt says what this module wants
// and lets the instance say the rest.
func (d *Dispatcher) prompt(m *phpscript.Module, comment, action, folder string) (string, error) {
	rel, err := phpscript.Inside(m.Root, action)
	if err != nil {
		return "", err
	}
	instructions, err := os.ReadFile(filepath.Join(m.Root, rel))
	if err != nil {
		return "", fmt.Errorf("%s: %w", action, err)
	}
	parts := []string{strings.TrimSpace(string(instructions))}
	if where := where(folder); where != "" {
		parts = append(parts, where)
	}
	parts = append(parts, "## The request\n\n"+strings.TrimSpace(comment))
	if index := d.artifacts(folder); index != "" {
		parts = append(parts, index)
	}
	return strings.Join(parts, "\n\n"), nil
}

// where names the folder of the artifact tree a run writes in, which is
// the only part of the tree it is asked to touch.
//
// A module that gives each of its threads a folder gets a scratch space
// per thread out of one tree: a new thread is an empty directory, so a
// run starts on nothing the way the first run of the instance did, and
// what an earlier thread left stays where it was.
//
// No folder says nothing, and the run works at the top of the tree with
// every other.
func where(folder string) string {
	folder = strings.Trim(folder, "/")
	if folder == "" {
		return ""
	}
	return "## Where to work\n\nEverything you read and write is under `" + folder +
		"/` in the artifact tree, `$TASK_UI_ARTIFACTS/" + folder +
		"`. That folder is this thread's and nothing else writes there. Make it if it is not there yet." +
		"\n\nEvery path you write in the report is relative to that folder, so a file at `" + folder +
		"/schema/user.up.sql` is linked as `schema/user.up.sql`. Nothing you write goes outside it."
}

// artifacts is what the folder holds now, as a list, so a run knows what
// an earlier one left without reading the directory first.
//
// An empty folder adds nothing: a list of nothing is a line of prompt
// that says nothing, and a new thread's folder is empty by definition.
func (d *Dispatcher) artifacts(folder string) string {
	if d.Store == nil {
		return ""
	}
	items := d.Store.List(strings.Trim(folder, "/"))
	if len(items) == 0 {
		return ""
	}
	held := "the instance holds"
	if folder != "" {
		held = "this thread has"
	}
	lines := make([]string, 0, len(items)+1)
	lines = append(lines, "## The artifacts "+held+" now")
	lines = append(lines, "")
	for _, item := range items {
		name := item.Name
		if item.Dir {
			name += "/"
		}
		lines = append(lines, "- "+name)
	}
	return strings.Join(lines, "\n")
}
