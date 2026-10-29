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
func (d *Dispatcher) Dispatch(ctx context.Context, m *phpscript.Module, comment, action string) (string, error) {
	if d.Board == nil {
		return "", errors.New("this instance starts no module runs")
	}
	user := access.UserFrom(ctx)
	if user == nil {
		return "", errors.New("a module run needs the user who asked for it")
	}
	body, err := d.prompt(m, comment, action)
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

// prompt is the action file, then the comment, then the artifact index.
//
// Nothing else is added. The run's own environment already names the
// artifact tree, the MCP server and what the user may do, and the agent's
// context repeats them, so a module's prompt says what this module wants
// and lets the instance say the rest.
func (d *Dispatcher) prompt(m *phpscript.Module, comment, action string) (string, error) {
	rel, err := phpscript.Inside(m.Root, action)
	if err != nil {
		return "", err
	}
	instructions, err := os.ReadFile(filepath.Join(m.Root, rel))
	if err != nil {
		return "", fmt.Errorf("%s: %w", action, err)
	}
	parts := []string{
		strings.TrimSpace(string(instructions)),
		"## The request\n\n" + strings.TrimSpace(comment),
	}
	if index := d.artifacts(); index != "" {
		parts = append(parts, index)
	}
	return strings.Join(parts, "\n\n"), nil
}

// artifacts is the top of the artifact tree, as a list, so a run knows
// what an earlier one left without reading the directory first. The
// artifacts belong to the instance and not to the thread, which is what
// lets a later thread build on an earlier one.
//
// An empty tree adds nothing: a list of nothing is a line of prompt that
// says nothing.
func (d *Dispatcher) artifacts() string {
	if d.Store == nil {
		return ""
	}
	items := d.Store.List("")
	if len(items) == 0 {
		return ""
	}
	lines := make([]string, 0, len(items)+1)
	lines = append(lines, "## The artifacts the instance holds now")
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
