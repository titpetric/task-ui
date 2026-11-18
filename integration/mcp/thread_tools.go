package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/storage"
)

// threadTools returns the tools that read conversations and the work of each run.
func threadTools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	return []Tool{
		{
			Name: "list_threads",
			Description: "List the conversations you can see, newest activity first: prompts to Claude and notes on the board, each with its agent, place and state. " +
				"With live, the open and working ones and the finished ones of the last hour.",
			InputSchema: object(map[string]Property{
				"workspace": str("Only threads of this workspace."),
				"folder":    str("Only threads in this folder or a directory under it, such as malta or malta/vhosts."),
				"user":      str("Only threads this user started."),
				"state":     {Type: "string", Description: "Only threads in this state.", Enum: []string{"open", "working", "done", "failed"}},
				"q":         str("Only threads whose text holds this."),
				"live":      {Type: "boolean", Description: "Only live threads: open and working, and those finished in the last hour."},
				"limit":     {Type: "integer", Description: "How many, 50 unless given."},
			}),
			call: listThreads,
		},
		{
			Name:        "read_thread",
			Description: "Read a conversation: the question, what Claude did in each run, and the replies.",
			InputSchema: object(map[string]Property{"id": str("The thread's ID, from list_threads.")}, "id"),
			call:        readThread,
		},
	}
}

// listArgs are the filters of list_threads.
type listArgs struct {
	Workspace string `json:"workspace"`
	Folder    string `json:"folder"`
	User      string `json:"user"`
	State     string `json:"state"`
	Query     string `json:"q"`
	Live      bool   `json:"live"`
	Limit     int    `json:"limit"`
}

// listThreads lists the thread feed as text.
func listThreads(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	in := listArgs{Limit: 50}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("threads", "1")
	q.Set("limit", fmt.Sprint(in.Limit))
	for key, value := range map[string]string{
		"workspace": in.Workspace,
		"folder":    in.Folder,
		"user":      in.User,
		"state":     in.State,
		"q":         in.Query,
	} {
		if value != "" {
			q.Set(key, value)
		}
	}
	if in.Live {
		q.Set("live", "1")
	}
	var out struct {
		Threads []storage.Conversation `json:"threads"`
	}
	if err := c.get(ctx, "/threads?"+q.Encode(), &out); err != nil {
		return "", err
	}
	if len(out.Threads) == 0 {
		return "No threads.", nil
	}
	shown := make([]string, 0, len(out.Threads))
	for _, thread := range out.Threads {
		shown = append(shown, showThreadRow(thread))
	}
	return strings.Join(shown, "\n"), nil
}

// readThread reads one conversation as text.
func readThread(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in struct {
		ID string `json:"id"`
	}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", errors.New("a thread ID is needed")
	}
	var view note.Conversation
	if err := c.get(ctx, "/threads/"+url.PathEscape(in.ID), &view); err != nil {
		return "", err
	}
	return showConversation(view), nil
}

// showThreadRow is one line of the list and the start of its question.
func showThreadRow(thread storage.Conversation) string {
	agent := thread.Agent
	if agent == "" {
		agent = "claude"
	}
	return fmt.Sprintf("thread %s by %s, %s, %s, %s, %d messages, agent %s\n  %s",
		thread.ID, thread.UserName, thread.State, place(thread.Dir), when(thread.CreatedAt), thread.Messages, agent, oneLine(thread.Body))
}

// showConversation is a thread's question, its runs and its replies.
func showConversation(view note.Conversation) string {
	var b strings.Builder
	thread := view.Thread
	fmt.Fprintf(&b, "thread %s by %s, %s, %s, agent %s\n%s\n", thread.ID, thread.UserName, thread.State, place(thread.Dir), view.Agent, thread.Body)
	for _, run := range view.Runs {
		fmt.Fprintf(&b, "run %s:\n", run.Run)
		for _, event := range run.Events {
			fmt.Fprintf(&b, "  %s: %s\n", event.Kind, indent(event.Body))
		}
	}
	for _, message := range view.Messages {
		if message.ID == thread.ID {
			continue
		}
		fmt.Fprintf(&b, "reply by %s, %s: %s\n", message.UserName, when(message.CreatedAt), indent(message.Body))
	}
	return b.String()
}

// place names where a thread ran, the root for an empty place.
func place(dir string) string {
	if dir == "" {
		return "_"
	}
	return dir
}

// oneLine is text on one line, cut at a length.
func oneLine(body string) string {
	body = strings.Join(strings.Fields(body), " ")
	if len(body) > 120 {
		return body[:120] + "..."
	}
	return body
}

// indent puts the lines of a body under the line that names it.
func indent(body string) string {
	return strings.ReplaceAll(strings.TrimSpace(body), "\n", "\n    ")
}
