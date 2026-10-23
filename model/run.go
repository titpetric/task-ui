package model

import (
	"strings"
	"time"
)

// The job IDs of the Claude built-ins, which every folder has beside
// @shell.
const (
	// PromptJob is the job of a prompt's run, in the folder and the
	// directory the prompt asks in.
	PromptJob = "@claude-prompt"
	// NoteJob is the job of the Notes agent's run, in the root, _.
	NoteJob = "@claude-note"
	// ClaudeJob is Claude Code in a terminal, in the folder or a
	// directory inside it.
	ClaudeJob = "@claude"
	// HostPrefix starts the ID of a command on a host of the inventory
	// over ssh, a job of the root, _: @host:dev1 runs on dev1, its run's
	// target, and @host:* on every host with an ssh login, a batch of
	// runs of @host:<host>. The colon keeps it apart from a mention.
	HostPrefix = "@host:"
	// HostAll is the job of a command on every host with an ssh login.
	HostAll = HostPrefix + "*"
	// HostBare is no job: a command on a host names its host.
	HostBare = "@host"
)

// HostJob returns the ID of the job of a command on a host, or on every host for *.
func HostJob(host string) string {
	return HostPrefix + host
}

// IsHost reports whether the run is a command on a host, @host:<host>.
func (r *Run) IsHost() bool {
	return r != nil && strings.HasPrefix(r.Job, HostPrefix)
}

// IsPrompt reports whether the run is a prompt's: @claude-prompt, or
// internal/claude-prompt, which ran the prompts of older threads.
func (r *Run) IsPrompt() bool {
	if r == nil {
		return false
	}
	return r.Job == PromptJob || (IsInternal(r.Folder) && r.Job == "claude-prompt")
}

// IsNote reports whether the run is the Notes agent's: @claude-note, or
// internal/claude-note, which ran the instructions of older notes.
func (r *Run) IsNote() bool {
	if r == nil {
		return false
	}
	return r.Job == NoteJob || (IsInternal(r.Folder) && r.Job == "claude-note")
}

// IsClaude reports whether the run is Claude's own: a prompt's, the Notes agent's, Claude Code in a terminal or an internal one.
// No prompt asks about one.
func (r *Run) IsClaude() bool {
	if r == nil {
		return false
	}
	return r.IsPrompt() || r.IsNote() || r.Job == ClaudeJob || IsInternal(r.Folder)
}

// Key is the identity of a job's runs that the run lock takes, as RunKey builds it.
func (r *Run) Key() string {
	return RunKey(r.Folder, r.Dir, r.Job)
}

// Parallel reports whether a job's runs go side by side: a prompt's and the Notes agent's, each a Claude session.
func (r *Run) Parallel() bool {
	return r.Job == PromptJob || r.Job == NoteJob
}

// LockKey is the key a run holds while it runs: Key, so one run per job.
// A job whose runs go in parallel adds the run's ID, so each run holds its own.
func (r *Run) LockKey() string {
	if r.Parallel() {
		return r.Key() + "#" + r.ID
	}
	return r.Key()
}

// RunKey is folder/dir/job for a job of a directory inside the folder, else
// folder/job; a host job's ID names its host, so hosts run at once, one each.
func RunKey(folder, dir, job string) string {
	if dir != "" {
		return folder + "/" + dir + "/" + job
	}
	return folder + "/" + job
}

// Where is what a title names the run by before its job: the folder,
// or folder/dir for a job of a directory inside it.
func (r Run) Where() string {
	if r.Dir != "" {
		return r.Folder + "/" + r.Dir
	}
	return r.Folder
}

// Place is the folder a page names the run by: Where, with the root's
// _/ dropped from a directory of it; _ alone for a job of the root.
func (r Run) Place() string {
	if r.Folder == Default && r.Dir != "" {
		return r.Dir
	}
	return r.Where()
}

// Name returns the run's title on a page: Place / job, or the job alone for the root, such as @host:web1.
// The stored row keeps the folder _.
func (r Run) Name() string {
	if r.Folder == Default && r.Dir == "" {
		return r.Job
	}
	return r.Place() + " / " + r.Job
}

// Path names the run in one word, folder/job without the root's _/: @claude-note, deploy/up.
func (r Run) Path() string {
	return strings.TrimPrefix(RunKey(r.Folder, r.Dir, r.Job), Default+"/")
}

// ArgList returns the arguments the run passed the job, after --.
func (r *Run) ArgList() []string {
	return DecodeList(r.Args)
}

// Finished reports whether the run ended.
func (r *Run) Finished() bool {
	return r != nil && Finished(r.State)
}

// Elapsed is how long the run took, or has been going.
func (r *Run) Elapsed() time.Duration {
	if r == nil || r.StartedAt == nil {
		return 0
	}
	if r.FinishedAt != nil {
		return r.FinishedAt.Sub(*r.StartedAt)
	}
	return time.Since(*r.StartedAt)
}
