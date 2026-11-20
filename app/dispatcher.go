package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/artifacts"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/phpfront"
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

// Dispatch starts the run a module's comment asks for and returns the ID of its note.
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
	asked, err := d.Board.AskAs(ctx, user, "", model.Default, body, note.Spoken{}, "", origin(root, action))
	if err != nil {
		return "", err
	}
	return asked.ID, nil
}

// prompt returns the action file, where to work, the comment and the folder's last commit.
func (d *Dispatcher) prompt(root, comment, action, folder string) (string, error) {
	rel, err := phpfront.Inside(root, action)
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
	if history := d.history(folder); history != "" {
		parts = append(parts, history)
	}
	return strings.Join(parts, "\n\n"), nil
}

// origin returns search for the header's question and module:<name> for any other action.
func origin(root, action string) string {
	if filepath.Base(action) == "search.md" {
		return "search"
	}
	return "module:" + filepath.Base(filepath.Clean(root))
}

// Revisions counts the folder's commits the prompt names: the newest one.
const Revisions = 1

// history returns the folder's last commit by message, so a request can name it.
func (d *Dispatcher) history(folder string) string {
	if d.Store == nil {
		return ""
	}
	revisions, err := d.Store.History(context.Background(), strings.Trim(folder, "/"), Revisions)
	if err != nil || len(revisions) == 0 {
		return ""
	}
	lines := []string{
		"## The last commit",
		"",
		"A revision of the artifact tree is a commit, and `read_artifact` takes one",
		"as its `version`: a request that names an earlier state is a read of that",
		"revision and a write of what it said. The files of the commit are not",
		"listed; list the folder with `find`.",
		"",
	}
	for _, revision := range revisions {
		lines = append(lines, "- `"+revision.Short+"` "+revision.When+", "+revision.Subject)
	}
	return strings.Join(lines, "\n")
}

// where returns the Where to work section, naming the folder and the files command.
func where(folder string) string {
	folder = strings.Trim(folder, "/")
	if folder == "" {
		return ""
	}
	return "## Where to work\n\nEverything you read and write is under `" + folder +
		"/` in the artifact tree, `$TASK_UI_ARTIFACTS/" + folder +
		"`. That folder is this thread's and nothing else writes there. Make it if it is not there yet." +
		"\n\nEvery path you write in the report is relative to that folder, so a file at `" + folder +
		"/schema/user.up.sql` is linked as `schema/user.up.sql`. Nothing you write goes outside it." +
		"\n\nThe files are not listed here: list them with `find \"$TASK_UI_ARTIFACTS/" + folder + "\" -type f`."
}
