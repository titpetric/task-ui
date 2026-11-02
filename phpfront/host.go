package phpfront

import (
	"context"
)

// Host answers what is not a row of the instance's database: what is derived
// from the filesystem, held in memory by a running process, or owned by a Go
// package for a reason. Every row is ported into code/Task instead.
//
// A nil Host answers every call empty.
type Host interface {
	// Whoami is the name the Guard resolved, "" when nobody is signed in.
	Whoami(ctx context.Context) string

	// Folders is the catalogue: folder.Catalogue parses Taskfiles to find jobs.
	Folders(ctx context.Context) []Folder

	// Running is the runs in progress, which are in no table until they end.
	Running(ctx context.Context) []Running

	// Dirty counts changed and untracked files under a checkout directory.
	Dirty(ctx context.Context, dir string) int

	// Rules is config/permissions.yml, which is embedded in the binary.
	Rules(ctx context.Context) Rules

	// Publish sends an event on the stream every live label reads.
	Publish(ctx context.Context, topic string, payload map[string]any) error
}

// Folder is one folder of the catalogue.
type Folder struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Dir   string   `json:"dir"`
	Jobs  []string `json:"jobs"`
}

// Running is one run in progress, by the key its job is listed under.
type Running struct {
	Key   string `json:"key"`
	ID    string `json:"id"`
	State string `json:"state"`
}

// Rules is config/permissions.yml: the titles in file order, and the path
// lists the file decides. Paths and Checks come over already built, so the
// decision logic is the only part of permission/ that is ported.
type Rules struct {
	Rules       []Rule   `json:"rules"`
	Sections    []Rule   `json:"sections"`
	Permissions []Rule   `json:"permissions"`
	Paths       []string `json:"paths"`
	Checks      []string `json:"checks"`
	Extras      []string `json:"extras"`
}

// Rule is one titled key of the permission file.
type Rule struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}
