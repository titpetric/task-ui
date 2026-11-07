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

// Dispatcher starts the Claude runs of a module, behind
// `Task::dispatch($comment, "action/refinement.md")` and the front end's
// `Task\Host::dispatch`.
//
// A dispatch is an ordinary prompt: the same board, the same run, the same
// artifact tree and the same MCP server a prompt from a page gets. A module's
// run appears on /runs and its answer arrives as a note reply.
type Dispatcher struct {
	Board *note.Board
	Users *storage.UserStorage
	Store *artifactstore.Store
}

// NewDispatcher returns the dispatcher of an instance.
func NewDispatcher(board *note.Board, users *storage.UserStorage, store *artifactstore.Store) *Dispatcher {
	return &Dispatcher{Board: board, Users: users, Store: store}
}

// Dispatch starts a run over a comment and the action file naming what to do
// with it, and returns the ID of the thread's note. The run goes in the
// background: the note exists on return, and Claude's reply arrives as a reply to it.
//
// The root is the tree the action file sits in, a plugin's module root or the
// front end's own, and the action is a path below the root. phpscript.Inside
// resolves every path the prompt reads against the root and refuses the rest.
//
// A module's run has the permissions of the user whose request reached it: a
// viewer's comment starts nothing, and the run reads only the secrets and keys
// that user may.
func (d *Dispatcher) Dispatch(ctx context.Context, root, comment, action, folder string) (string, error) {
	if d.Board == nil {
		return "", errors.New("this instance starts no module runs")
	}
	user := access.UserFrom(ctx)
	if user == nil {
		return "", errors.New("a module run needs the user who asked for it")
	}
	body, err := d.prompt(root, comment, action, folder)
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

// prompt returns the action file, then where to work, then the comment, then
// the index of what the folder holds already.
//
// The run's environment already names the artifact tree, the MCP server and the
// user's permissions, and the agent's context repeats them; the prompt adds none
// of that.
func (d *Dispatcher) prompt(root, comment, action, folder string) (string, error) {
	rel, err := phpscript.Inside(root, action)
	if err != nil {
		return "", err
	}
	instructions, err := os.ReadFile(filepath.Join(root, rel))
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
	if history := d.history(folder); history != "" {
		parts = append(parts, history)
	}
	return strings.Join(parts, "\n\n"), nil
}

// Revisions is how many of a folder's commits the prompt lists: enough to roll
// back a few steps of refinement, short enough to stay a paragraph.
const Revisions = 12

// history returns the folder's last Revisions commits, newest first, so a
// request to restore an earlier state can name one. Every write to the
// artifact tree is a commit, and read_artifact takes a revision as its version.
//
// A folder with no commits returns "".
func (d *Dispatcher) history(folder string) string {
	if d.Store == nil {
		return ""
	}
	revisions, err := d.Store.History(context.Background(), strings.Trim(folder, "/"), Revisions)
	if err != nil || len(revisions) == 0 {
		return ""
	}
	lines := []string{
		"## The history",
		"",
		"The artifact tree is a git repository and every revision of it is a",
		"commit. These are the folder's, newest first. `read_artifact` takes one",
		"of them as its `version`, so a request to put something back the way it",
		"was is a read of the revision that held it and a write of what it said.",
		"",
	}
	for _, revision := range revisions {
		line := "- `" + revision.Short + "` " + revision.When + ", " + revision.Subject
		if len(revision.Files) > 0 {
			line += ": " + strings.Join(revision.Files, ", ")
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// where returns the Where to work section: the folder of the artifact tree a
// run writes in, and the only part of the tree it may touch. A module that
// names a folder per thread gets one scratch space per thread out of a single
// tree, each thread starting on an empty directory.
//
// An empty folder returns "", and the run works at the top of the tree.
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

// artifacts returns the folder's contents as a list, which saves a run walking
// the directory to find what an earlier one left. An empty folder returns "".
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
