package note

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/app/model"
)

// source returns TASK_UI_PROMPT_SOURCE: the thread's note, its question's reply, place and link.
func (b *Board) source(ctx context.Context, thread *model.Note) string {
	var text strings.Builder
	fmt.Fprintf(&text, "## The prompt's source\n\nThe prompt is the thread of the task-ui note %s", thread.ID)
	if asking := b.asking(ctx, thread); asking != "" {
		fmt.Fprintf(&text, ", its question the reply %s", asking)
	}
	fmt.Fprintf(&text, ", %s, %s.\n\n", b.place(ctx, thread), b.link(b.threadPath(ctx, thread)))
	text.WriteString("Your final message is the answer the asker reads, which task-ui posts under the question.")
	return text.String()
}

// asking returns the ID of the last reply someone other than Claude wrote, or "".
func (b *Board) asking(ctx context.Context, thread *model.Note) string {
	replies, err := b.Notes.Replies(ctx, []string{thread.ID})
	if err != nil {
		return ""
	}
	answerer := b.AnswererOf(ctx, thread)
	list := replies[thread.ID]
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].UserName != answerer && list[i].UserName != b.Answerer() {
			return list[i].ID
		}
	}
	return ""
}

// place says where a prompt asks: its area, the run it is about, or the folder path.
func (b *Board) place(ctx context.Context, thread *model.Note) string {
	switch {
	case thread.Area == model.AreaWorkflows:
		return "the prompt of the Workflows page, in " + model.PlaceLabel(thread.Dir)
	case thread.AboutRunID != "":
		return "about the run " + thread.AboutRunID + ", in " + model.PlaceLabel(thread.Dir)
	case thread.AboutWorkflowRunID != "":
		return "a prompt of the workflow run " + thread.AboutWorkflowRunID + ", in " + model.PlaceLabel(thread.Dir)
	}
	return "in " + model.PlaceLabel(thread.Dir)
}

// threadPath returns the page path of a prompt's thread, with its anchor.
func (b *Board) threadPath(ctx context.Context, thread *model.Note) string {
	anchor := "#note-" + thread.ID
	switch {
	case thread.Area == model.AreaWorkflows:
		return "/workflows" + anchor
	case thread.AboutRunID != "":
		return "/runs/" + thread.AboutRunID + anchor
	case thread.AboutWorkflowRunID != "" && b.WorkflowRuns != nil:
		if run, err := b.WorkflowRuns.Get(ctx, thread.AboutWorkflowRunID); err == nil {
			return "/workflows/" + run.WorkflowID + "/runs/" + run.ID + anchor
		}
	}
	if thread.Dir == "" || thread.Dir == model.Default {
		return "/f/" + model.Default + anchor
	}
	// A named workspace's folder has its page by its ID; any other is under the root's.
	if thread.WorkspaceID != "" && thread.WorkspaceID != model.Default {
		return "/f/" + thread.Dir + anchor
	}
	return "/f/" + model.Default + "/" + thread.Dir + anchor
}

// RunReply is AnswerThread under the name the MCP note_comment hook called.
//
// Deprecated: a prompt run's final message is its answer, and the hook is being phased out. Use AnswerThread.
func (b *Board) RunReply(ctx context.Context, user *model.User, id, body, state string) (*model.Note, error) {
	return b.AnswerThread(ctx, user, id, body, state)
}

// AnswerThread posts a prompt run's finding to its thread, in the reply a rerun rewrites.
// The thread's state follows its run, which sets no state there; another thread takes it as Comment does.
func (b *Board) AnswerThread(ctx context.Context, user *model.User, id, body, state string) (*model.Note, error) {
	thread, err := b.Notes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !thread.IsPrompt() || thread.ParentID != "" {
		return b.Comment(ctx, user, id, body, state)
	}
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: a reply needs some text", ErrInvalid)
	}
	var reply *model.Note
	if again := b.working(ctx, thread); again != nil {
		b.answer(ctx, thread, again, body)
		reply = again
	} else {
		reply = b.reply(ctx, thread, body)
		if reply == nil {
			return nil, fmt.Errorf("the reply to %s was not saved", thread.ID)
		}
	}
	b.repliedMu.Lock()
	if b.replied == nil {
		b.replied = map[string]bool{}
	}
	b.replied[thread.ID] = true
	b.repliedMu.Unlock()
	return reply, nil
}

// takeReplied reports whether the thread's run replied since the last call, and forgets it.
func (b *Board) takeReplied(id string) bool {
	b.repliedMu.Lock()
	defer b.repliedMu.Unlock()
	replied := b.replied[id]
	delete(b.replied, id)
	return replied
}
