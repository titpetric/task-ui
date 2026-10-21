package agent

import (
	"os"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Namespace is the ID of the folder the host commands make up.
const Namespace = model.Internal

// Folder returns the internal folder: the host commands as jobs that run
// task-ui agent in the root. The notes agent's job allows the agent's
// user, the other two admins; only admins see the folder at all. The
// jobs are host commands: a run of one in a directory under the root
// runs there, as claude-prompt does for a prompt in a folder.
func Folder(root, agentUser string) *folder.Folder {
	exe, err := os.Executable()
	if err != nil {
		exe = "task-ui"
	}
	noteAllow := []string{model.RoleAdmin}
	if agentUser != "" {
		noteAllow = []string{agentUser}
	}
	source := &detector.Source{
		Kind:        "script",
		File:        Namespace,
		Name:        Namespace,
		Description: "Host commands: Claude Code on the Notes board, on a workspace, and in a terminal.",
		Command:     []string{exe, "agent"},
		Targets: []*detector.Target{
			{Name: "claude-note", Description: "Claude Code on an instruction from the Notes board, until it is done", Args: []string{"note"}},
			{Name: "claude-prompt", Description: "Claude Code on a prompt against a workspace, in its checkout", Args: []string{"prompt"}},
			{Name: "claude", Description: "Claude Code in a terminal, in the folders task-ui runs", Args: []string{"shell"}, Interactive: true},
		},
	}
	allow := map[string][]string{"claude-note": noteAllow, "claude-prompt": {model.RoleAdmin}, "claude": {model.RoleAdmin}}
	f := &folder.Folder{ID: Namespace, Title: Namespace, Description: source.Description, Dir: root, Sources: []*detector.Source{source}, HostJobs: true}
	for _, target := range source.Targets {
		f.Jobs = append(f.Jobs, &folder.Job{
			Target:  *target,
			Folder:  Namespace,
			ID:      target.Name,
			Kind:    source.Kind,
			Source:  source.File,
			Command: append(append([]string{}, source.Command...), target.Argv()...),
			Allow:   allow[target.Name],
		})
	}
	return f
}
