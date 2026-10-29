package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// Projects lists the projects of a workspace, the folders a workflow
// that runs over it walks, in order. A run walks every step once per
// project: a step whose folder is in a project of the workspace runs in
// each project in turn, at the same path inside it, and any other step
// runs as it stands. app.Catalogue is one.
type Projects interface {
	WorkspaceProjects(ctx context.Context, workspace string) []*folder.Folder
}

// Expand returns the steps a run over the workspace's projects walks, with the project index of each.
func Expand(steps []model.WorkflowStep, workspace string, projects []string) ([]model.WorkflowStep, []int) {
	listed := map[string]bool{}
	for _, project := range projects {
		listed[project] = true
	}
	expanded := make([]model.WorkflowStep, 0, len(steps)*len(projects))
	groups := make([]int, 0, cap(expanded))
	for i, project := range projects {
		for _, step := range steps {
			if name, rest, ok := projectOf(step.Folder, workspace); ok && listed[name] && !step.IsWorkflow() {
				step.Folder = layout.Join(workspace, project+rest)
			}
			expanded = append(expanded, step)
			groups = append(groups, i)
		}
	}
	return expanded, groups
}

// projectOf splits a folder ID of the workspace into its project's name
// and the path below it, with its leading slash; false for a folder of
// another workspace.
func projectOf(id, workspace string) (name, rest string, ok bool) {
	rel := id
	if workspace != "" && workspace != model.Default {
		if rel, ok = strings.CutPrefix(id, workspace+"/"); !ok {
			return "", "", false
		}
	}
	if rel == "" || rel == model.Default {
		return "", "", false
	}
	name, rest, _ = strings.Cut(rel, "/")
	if rest != "" {
		rest = "/" + rest
	}
	return name, rest, true
}

// over returns the steps a run of the workflow walks, each with its
// project's index, and the projects' folders by ID: the steps as they
// stand for a workflow that runs over no workspace.
func (r *Runner) over(ctx context.Context, wf *model.Workflow, steps []model.WorkflowStep) ([]model.WorkflowStep, []int, map[string]*folder.Folder, error) {
	if wf == nil || wf.RunsOver == "" {
		return steps, make([]int, len(steps)), nil, nil
	}
	var folders []*folder.Folder
	if r.Projects != nil {
		folders = r.Projects.WorkspaceProjects(ctx, wf.RunsOver)
	}
	if len(folders) == 0 {
		return nil, nil, nil, fmt.Errorf("the workspace %s has no projects to run over", wf.RunsOver)
	}
	byID := map[string]*folder.Folder{}
	names := make([]string, 0, len(folders))
	for _, f := range folders {
		byID[f.ID] = f
		name, _, _ := projectOf(f.ID, wf.RunsOver)
		names = append(names, name)
	}
	expanded, groups := Expand(steps, wf.RunsOver, names)
	return expanded, groups, byID, nil
}
