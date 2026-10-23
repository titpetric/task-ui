package note

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/titpetric/task-ui/model"
)

// EarlierAnswers is how many answers of a workflow's earlier runs its
// prompts get, and answerBytes the most of each.
const (
	EarlierAnswers = 3
	answerBytes    = 1536
)

// workflowContext is the Workflow context block every prompt of a
// workflow run gets before the steps: the workflow and its link, the
// run's number, start, starter and input, and the replies of the
// workflow's last EarlierAnswers finished runs that have one, newest
// first, each cut to answerBytes.
func (b *Board) workflowContext(ctx context.Context, run *model.WorkflowRun) string {
	var text strings.Builder
	text.WriteString("## Workflow context\n\n")
	fmt.Fprintf(&text, "The workflow %s, %s.\n", run.Name, b.link("/workflows/"+run.WorkflowID))
	n, _ := b.WorkflowRuns.Number(ctx, run)
	fmt.Fprintf(&text, "This is its run %d, %s", n, b.link("/workflows/"+run.WorkflowID+"/runs/"+run.ID))
	if run.StartedAt != nil {
		text.WriteString(", started " + run.StartedAt.UTC().Format(time.RFC3339))
	}
	text.WriteString(" by " + run.UserName)
	if input := strings.TrimSpace(run.Input); input != "" {
		text.WriteString(", with the input:\n\n" + input + "\n")
	} else {
		text.WriteString(", with no input.\n")
	}
	earlier, err := b.WorkflowRuns.Answered(ctx, run.WorkflowID, run.ID, EarlierAnswers)
	if err != nil || len(earlier) == 0 {
		text.WriteString("\nThere are no earlier runs with an answer to compare with.\n\n")
		return text.String()
	}
	fmt.Fprintf(&text, "\nThe answers of its last %d finished runs, newest first, to stay consistent with:\n", len(earlier))
	for _, prior := range earlier {
		when := ""
		if prior.StartedAt != nil {
			when = prior.StartedAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(&text, "\n### The run of %s, %s\n\n", when, prior.State)
		answer := strings.TrimSpace(prior.Answer)
		cut := len(answer) > answerBytes
		if cut {
			answer = strings.ToValidUTF8(answer[:answerBytes], "")
		}
		text.WriteString("```\n" + fence(answer) + "\n```\n")
		if cut {
			fmt.Fprintf(&text, "\nThe answer is cut to its first %.1f KiB.\n", float64(answerBytes)/1024)
		}
	}
	return text.String() + "\n"
}

// link is a path of the instance as a prompt gets it: under BaseURL,
// or the path alone without one.
func (b *Board) link(path string) string {
	return strings.TrimRight(b.BaseURL, "/") + path
}
