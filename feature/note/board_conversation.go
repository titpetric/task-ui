package note

import (
	"context"
	"fmt"

	"github.com/titpetric/task-ui/app/model"
)

// Conversation returns a root note, its replies and the work of its runs.
func (b *Board) Conversation(ctx context.Context, id string) (*Conversation, error) {
	thread, err := b.Notes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if thread.ParentID != "" {
		return nil, fmt.Errorf("%w: a conversation is a root note", ErrInvalid)
	}
	replied, err := b.Notes.Replies(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	replies := replied[id]
	var events []model.NoteEvent
	if b.NoteEvents != nil {
		if events, err = b.NoteEvents.List(ctx, id); err != nil {
			return nil, err
		}
	}
	conversation := &Conversation{
		Thread:   thread,
		Agent:    thread.Agent,
		Messages: append([]model.Note{*thread}, replies...),
		Runs:     groupEvents(events),
	}
	if conversation.Agent == "" {
		conversation.Agent = b.AnswererOf(ctx, thread)
	}
	if b.Users != nil {
		if agents, err := b.Users.Agents(ctx); err == nil {
			for _, agent := range agents {
				conversation.Agents = append(conversation.Agents, agent.Name)
			}
		}
	}
	for i := range conversation.Runs {
		for _, reply := range replies {
			if reply.RunID != "" && reply.RunID == conversation.Runs[i].Run {
				conversation.Runs[i].Answer = reply.ID
				break
			}
		}
	}
	return conversation, nil
}

// groupEvents groups a thread's events by the question and run that made them.
func groupEvents(events []model.NoteEvent) []ConversationRun {
	var runs []ConversationRun
	index := map[string]int{}
	for _, event := range events {
		key := event.QuestionID + "\x00" + event.RunID
		at, ok := index[key]
		if !ok {
			at = len(runs)
			index[key] = at
			runs = append(runs, ConversationRun{Question: event.QuestionID, Run: event.RunID})
		}
		runs[at].Events = append(runs[at].Events, event)
	}
	return runs
}
