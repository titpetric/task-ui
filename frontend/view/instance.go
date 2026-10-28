package view

import (
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/toolenv"
)

// Instance is what the server runs with, as the Instance step of the
// settings shows it: the configuration from the environment and the
// command line, less anything secret. A path shows; a token never does.
type Instance struct {
	// URL is TASK_UI_URL, empty when the links derive from each request.
	URL      string
	Listen   string
	Root     string
	DataDir  string
	SSHDir   string
	Htpasswd string
	// HostData is the data directory on the host, empty when the same.
	HostData string
	Discover bool
	// AutoLogin is the user every request signs in as, if any.
	AutoLogin string
	// TimeZone is the zone of the server's clock, as the pages and the
	// hours read it.
	TimeZone string
	// SessionTTL, Keep and Grace are TASK_UI_SESSION_TTL, TASK_UI_KEEP
	// and TASK_UI_GRACE, as durations read.
	SessionTTL string
	Keep       string
	Grace      string
	// ArtifactHistory is TASK_UI_ARTIFACT_HISTORY, 0 for every version.
	ArtifactHistory int
	// ComposeProject is the compose project task-ui runs in, if any.
	ComposeProject string
	// Beszel says the Beszel integration has a hub and a token, which
	// the Inventory reads.
	Beszel bool
	// Backlog is how the slots and the line for them stand, nil without
	// a runner.
	Backlog *runner.BacklogState
	// ParentEnv is the parent environment mode in effect, empty without
	// a runner; Env the names of the server's variables, never their
	// values, for an admin alone.
	ParentEnv string
	Env       []runner.EnvName
	// Storage is what the instance keeps on disk, by item, for an admin
	// alone; nil leaves the Storage table out.
	Storage *diskusage.Table
	// Tooling is what the instance found in its PATH at start, a table
	// per category; empty says there is none.
	Tooling []ToolGroup
}

// Tool is a tool task-ui found in the PATH at start, as the Tooling
// section of the Instance page lists it.
type Tool struct {
	// ID is the tool's name in the check URL, as "docker-compose".
	ID          string
	Name        string
	Version     string
	Path        string
	Description string
}

// ToolGroup is one table of the Tooling section.
type ToolGroup struct {
	// Title is the category's title, such as Runtimes.
	Title string
	Tools []Tool
}

// ToolCategories are the categories of the Tooling section, with their
// titles, in the order the page lists them.
var ToolCategories = []struct{ Category, Title string }{
	{Category: toolenv.CategoryRuntime, Title: "Runtimes"},
	{Category: toolenv.CategoryTooling, Title: "Tooling"},
	{Category: toolenv.CategoryCustom, Title: "Custom tooling"},
}

// ToolGroups groups the stored tools into the tables of the Tooling
// section, the categories in ToolCategories' order, a category with no
// tool left out.
func ToolGroups(tools []model.Tool) []ToolGroup {
	groups := []ToolGroup{}
	for _, category := range ToolCategories {
		group := ToolGroup{Title: category.Title}
		for _, tool := range tools {
			if tool.Category != category.Category {
				continue
			}
			id := ""
			if spec, ok := toolenv.SpecByName(tool.Name); ok {
				id = spec.ID
			}
			group.Tools = append(group.Tools, Tool{ID: id, Name: tool.Name, Version: tool.Version, Path: tool.Path, Description: tool.Description})
		}
		if len(group.Tools) > 0 {
			groups = append(groups, group)
		}
	}
	return groups
}

// EnvBase returns runner.BaseEnv, the variables None passes.
func (i *Instance) EnvBase() []string {
	return runner.BaseEnv
}

// EnvSensitive returns runner.SensitiveWords.
func (i *Instance) EnvSensitive() []string {
	return runner.SensitiveWords
}
