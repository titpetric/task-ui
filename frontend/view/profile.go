package view

import (
	"github.com/titpetric/task-ui/model"
)

// Profile is a user's page beyond the user: what they did lately, as
// far as the one who looks may see, on the profile, and their
// permissions on the edit page.
type Profile struct {
	// Own says the page is the signed in user's.
	Own bool
	// CanEdit shows Edit on the profile: the signed in user's own, or
	// user/edit.
	CanEdit bool
	// Runs are the user's latest runs in folders the viewer sees, and
	// FlowRuns their latest workflow runs.
	Runs     []Run
	FlowRuns []model.WorkflowRun
	// Notes are the notes the user started, for a viewer with
	// note/view; Prompts the prompts they asked, with prompt/view, in
	// places the viewer sees.
	Notes   []model.Note
	Prompts []model.Note
	// Matrix is the user's policy with what they may do, on the edit
	// page.
	Matrix *Matrix
}

// Activity reports whether the profile lists anything the user did.
func (p *Profile) Activity() bool {
	return len(p.Runs)+len(p.FlowRuns)+len(p.Notes)+len(p.Prompts) > 0
}
