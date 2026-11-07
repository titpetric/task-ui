package claude

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/titpetric/task-ui/git"
)

// PullTriage is what the triage of a failed git pull asks Claude: which case the checkout is in, and nothing else.
var PullTriage = "A git pull of a checkout failed. The input is what the pull said and the checkout's git status --porcelain=v2 --branch. " +
	"Name the case the checkout is in with one of these words and nothing else, no punctuation and no explanation: " + strings.Join(git.Cases(), ", ") + ". " +
	"dirty is work in the tree a fast forward would overwrite; diverged is commits the upstream does not have; " +
	"upstream-gone is a branch the remote no longer has; detached is a HEAD on no branch; index-lock is a lock file left behind; " +
	"auth is a remote that refused or did not answer; broken is a repository git cannot read; unknown is none of these. " +
	"Run nothing, change nothing, name no file and no path: task-ui runs the repair itself."

// triageTimeout bounds one triage when the Triager names none.
const triageTimeout = 90 * time.Second

// Triager asks Claude which of git's recovery cases a failed pull is.
type Triager struct {
	// Command is the Claude Code binary, claude unless set.
	Command string
	// Timeout bounds one triage.
	Timeout time.Duration
	// Env returns the integration variables Claude signs in with.
	Env func(ctx context.Context) (map[string]string, error)
}

// Triage returns the case Claude names for a failed pull, "" when it names none of them.
func (t *Triager) Triage(ctx context.Context, dir, failure, status string) (string, error) {
	if t.Env == nil {
		return "", ErrNoToken
	}
	env, err := t.Env(ctx)
	if err != nil {
		return "", err
	}
	timeout := t.Timeout
	if timeout == 0 {
		timeout = triageTimeout
	}
	// The checkout's name and not its path: a path the model reads is a path it can name back.
	document := "checkout: " + filepath.Base(dir) + "\n\nthe pull said:\n" + failure + "\n\nthe status is:\n" + status + "\n"
	answer, err := (&Summarizer{Command: t.Command, Timeout: timeout}).Summarize(ctx, env, "", PullTriage, document)
	if err != nil {
		return "", err
	}
	return caseOf(answer), nil
}

// caseOf returns the case an answer names, "" for an answer that is no case and classified nothing.
func caseOf(answer string) string {
	word := strings.ToLower(strings.Trim(strings.TrimSpace(answer), ".`\"'"))
	if slices.Contains(git.Cases(), word) {
		return word
	}
	return ""
}
