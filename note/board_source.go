package note

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// source is what a prompt's run reads of where the prompt comes from,
// for TASK_UI_PROMPT_SOURCE: the thread's note ID, the reply that asks
// when a reply asks, the place or area and the thread's link, and how
// to answer there, a reply with note_comment over the task MCP.
func (b *Board) source(ctx context.Context, thread *model.Note) string {
	var text strings.Builder
	fmt.Fprintf(&text, "## The prompt's source\n\nThe prompt is the thread of the task-ui note %s", thread.ID)
	if asking := b.asking(ctx, thread); asking != "" {
		fmt.Fprintf(&text, ", its question the reply %s", asking)
	}
	fmt.Fprintf(&text, ", %s, %s.\n\n", b.place(ctx, thread), b.link(b.threadPath(ctx, thread)))
	fmt.Fprintf(&text, "When you have the result, post it as a reply in that thread with the task MCP: note_comment with the id %s and the answer as the body, no state, then end. ", thread.ID)
	text.WriteString("That reply is the answer the asker reads. Your final message is a fallback, which task-ui posts only when you posted no reply.")
	return text.String()
}

// asking is the ID of the last reply in a thread that someone other than
// Claude wrote: the question of a prompt asked again. A thread with no
// such reply asks with its note.
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

// place says where a prompt asks: its area, the run or workflow run it
// is about, or its folder and directory.
func (b *Board) place(ctx context.Context, thread *model.Note) string {
	switch {
	case thread.Area == model.AreaWorkflows:
		return "the prompt of the Workflows page, in " + thread.Dir
	case thread.AboutRunID != "":
		return "about the run " + thread.AboutRunID + ", in " + thread.Dir
	case thread.AboutWorkflowRunID != "":
		return "a prompt of the workflow run " + thread.AboutWorkflowRunID + ", in " + thread.Dir
	}
	return "in " + thread.Dir
}

// threadPath is the path of the page that shows a prompt's thread, with
// the thread's anchor.
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
	// A folder of a named workspace has its page by its ID, any other
	// under the root's.
	if thread.WorkspaceID != "" && thread.WorkspaceID != model.Default {
		return "/f/" + thread.Dir + anchor
	}
	return "/f/" + model.Default + "/" + thread.Dir + anchor
}

// RunReply posts what a prompt's run found in its own thread, as the
// run's token asks with note_comment: a reply as Claude, the author of
// the thread's answers, with the thread's run, in the reply a rerun
// rewrites when there is one. The thread's state follows its run, so
// RunReply sets none, and the run's final message is no reply once one
// came this way. A thread that is no prompt's takes the reply as Comment
// does.
func (b *Board) RunReply(ctx context.Context, user *model.User, id, body, state string) (*model.Note, error) {
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

// takeReplied reports whether a run of the thread replied with RunReply
// since the last call, and forgets it.
func (b *Board) takeReplied(id string) bool {
	b.repliedMu.Lock()
	defer b.repliedMu.Unlock()
	replied := b.replied[id]
	delete(b.replied, id)
	return replied
}
