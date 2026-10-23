package view

// Pipeline is the workflow's steps top to bottom, with the states of a latest run that has the same steps.
// The workspaces place the jobs' links, as in Chain.
func Pipeline(steps, last []Step, workspaces map[string]int) []Link {
	if len(last) == len(steps) && len(steps) > 0 {
		same := true
		for i := range steps {
			if steps[i].Folder != last[i].Folder || steps[i].Job != last[i].Job || steps[i].WorkflowID != last[i].WorkflowID {
				same = false
				break
			}
		}
		if same {
			// A run's step keeps no directory of its prompt; the
			// workflow's step has it.
			for i := range last {
				last[i].Dir = steps[i].Dir
			}
			return Chain(last, workspaces)
		}
	}
	return Chain(steps, workspaces)
}
