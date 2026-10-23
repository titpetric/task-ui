// Package folder is the catalogue: a folder is a working directory and the
// targets found in it, and a job is one of those targets with an identity
// inside the folder.
package folder

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/skills"
)

// Folder is a working directory and the conventions detected in it. Each
// convention is a Source, and every target of every source is a Job.
type Folder struct {
	// ID is the URL safe name of the folder.
	ID string `json:"id"`
	// Title is the name the first source declares, or the folder name.
	Title string `json:"title"`
	// Description is what the first source says about itself.
	Description string `json:"description,omitempty"`

	// Dir is the working directory the jobs run in.
	Dir string `json:"-"`
	// Rel is the directory inside the folder ID names that Dir is, for
	// the folder of a directory Sub returns; empty for the folder itself.
	Rel string `json:"rel,omitempty"`

	// Sources are the conventions detected in the folder, in detection
	// order: pipeline files first, then build files, then scripts.
	Sources []*detector.Source `json:"sources"`

	// Includes are the files the sources read into their variables.
	Includes []string `json:"includes,omitempty"`
	// EnvFiles are the dotenv files the sources read, relative to Dir.
	// These are the files task-ui can manage.
	EnvFiles []string `json:"env_files,omitempty"`

	// Allow lists the users and roles that may run this folder's jobs,
	// from the configuration. Empty allows every runner.
	Allow []string `json:"allow,omitempty"`
	// Hours says when the folder's jobs may run, as the hours package
	// reads them. Empty is always.
	Hours string `json:"hours,omitempty"`

	// Jobs are every source's jobs, in source order.
	Jobs []*Job `json:"jobs"`

	// Probe asks atkins for the skills that apply in Dir, the jobs
	// SkillJob resolves. Nil lists none.
	Probe *skills.Prober `json:"-"`

	// Error holds the parse errors of the sources, so a folder with a
	// broken file still shows in the listing and says what is wrong.
	Error string `json:"error,omitempty"`
}

// Job returns the job by ID, name or alias, or nil. The ID is what URLs
// and history use; the name is the target as its file spells it. An ID
// that starts with ./ is the job of a file that runs: see FileJob; @shell
// is the shell job, see ShellJob, and @claude Claude's, see ClaudeJob;
// @claude-prompt and @claude-note are none, see PromptJob and NoteJob;
// @host:<host> is a host job of the root, see HostJob, and @cleanup its
// cleanup job, see CleanupJob;
// any other that starts with @ the job of an atkins skill: see SkillJob.
func (f *Folder) Job(id string) *Job {
	if f == nil {
		return nil
	}
	switch id {
	case ShellID:
		return f.ShellJob()
	case ClaudeID:
		return f.ClaudeJob()
	case CleanupID:
		return f.CleanupJob()
	case PromptID, NoteID:
		return nil
	}
	if rel, ok := strings.CutPrefix(id, FilePrefix); ok {
		return f.FileJob(rel)
	}
	if host, ok := strings.CutPrefix(id, HostPrefix); ok {
		return f.HostJob(host)
	}
	if name, ok := strings.CutPrefix(id, SkillPrefix); ok {
		return f.SkillJob(name)
	}
	for _, j := range f.Jobs {
		if j.ID == id {
			return j
		}
	}
	for _, j := range f.Jobs {
		if j.Name == id || slices.Contains(j.Aliases, id) {
			return j
		}
	}
	return nil
}

// SourceJobs returns the jobs of one source, in order.
func (f *Folder) SourceJobs(source *detector.Source) []*Job {
	var jobs []*Job
	for _, job := range f.Jobs {
		if job.Source == source.File && job.Kind == source.Kind {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// HoursOf returns the hours a job may run in: its own, or the folder's.
func (f *Folder) HoursOf(job *Job) string {
	if job != nil && job.Hours != "" {
		return job.Hours
	}
	if f == nil {
		return ""
	}
	return f.Hours
}

// Name is the last segment of the ID: what the sidebar shows at its depth.
func (f *Folder) Name() string {
	return f.name()
}

// Lines counts the lines of a file in the folder, for the Show button of
// a file its pipeline reads: a last line without a newline counts. A
// file it cannot read, or one outside the folder, is 0.
func (f *Folder) Lines(name string) int {
	if name == "" || filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(filepath.FromSlash(name)), "..") {
		return 0
	}
	data, err := os.ReadFile(filepath.Join(f.Dir, filepath.FromSlash(name)))
	if err != nil || len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func (f *Folder) name() string {
	if i := strings.LastIndex(f.ID, "/"); i >= 0 {
		return f.ID[i+1:]
	}
	return f.ID
}

// Depth is how far below the root the folder is: zero for a folder of the root.
func (f *Folder) Depth() int {
	return strings.Count(f.ID, "/")
}

// Parent is the ID of the folder above, or "" at the top.
func (f *Folder) Parent() string {
	if i := strings.LastIndex(f.ID, "/"); i >= 0 {
		return f.ID[:i]
	}
	return ""
}
