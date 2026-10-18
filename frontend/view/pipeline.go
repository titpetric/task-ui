package view

// Pipeline is the workflow's steps as the page draws them top to
// bottom, with the states of its latest run when that run had the same
// steps. A run of an older chain lends no states.
func Pipeline(steps, last []Step) []Link {
	if len(last) == len(steps) && len(steps) > 0 {
		same := true
		for i := range steps {
			if steps[i].Folder != last[i].Folder || steps[i].Job != last[i].Job || steps[i].WorkflowID != last[i].WorkflowID {
				same = false
				break
			}
		}
		if same {
			return Chain(last)
		}
	}
	return Chain(steps)
}
