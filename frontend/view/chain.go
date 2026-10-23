package view

// Chain turns steps into links. It names a folder once, for the steps
// that follow in it, a workflow step by its name, and a prompt step as
// a prompt in its place; a job links to its page, as FolderHref places
// its folder among the workspaces, a prompt to its folder's. A run's
// step links to the run it made.
func Chain(steps []Step, workspaces map[string]int) []Link {
	links := make([]Link, 0, len(steps))
	last := ""
	for _, s := range steps {
		l := Link{State: s.State, RunID: s.RunID, Disabled: s.Disabled}
		switch {
		case s.Prompt != "":
			place := s.Folder
			if s.Dir != "" {
				place += "/" + s.Dir
			}
			l.Prompt, l.Text, l.Href = s.Prompt, "prompt in "+place, FolderHref(s.Folder, workspaces)
			last = ""
		case s.WorkflowID != "":
			l.Workflow, l.Text, l.Href = true, s.WorkflowName, "/workflows/"+s.WorkflowID
			l.WorkflowID, l.WorkflowRunID = s.WorkflowID, s.WorkflowRunID
			if l.Text == "" {
				l.Text = "a workflow that is gone"
			}
			last = ""
		default:
			l.Text, l.Href = s.Job, FolderHref(s.Folder, workspaces)+"/"+s.Job
			if s.Folder != last {
				l.Text = s.Folder + " / " + s.Job
			}
			last = s.Folder
		}
		if s.RunID != "" {
			l.Href = "/runs/" + s.RunID
		}
		links = append(links, l)
	}
	return links
}
