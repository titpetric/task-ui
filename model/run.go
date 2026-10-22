package model

import (
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
)

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

// IsClaude reports whether the run is one of Claude's own: a prompt's,
// the Notes agent's, Claude Code in a terminal, or any run of the
// internal folder the older ones carry. No prompt asks about one.
func (r *Run) IsClaude() bool {
	if r == nil {
		return false
	}
	return r.IsPrompt() || r.IsNote() || r.Job == ClaudeJob || IsInternal(r.Folder)
}

// Key is the identity the run lock uses: one run per job per folder
// and directory inside it.
func (r *Run) Key() string {
	return RunKey(r.Folder, r.Dir, r.Job)
}

// RunKey is the key of a job: folder/job, or folder/dir/job for a job
// of a directory inside the folder.
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
