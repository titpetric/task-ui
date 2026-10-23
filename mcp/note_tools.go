package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

// noteTools are the tools that write as the token's user: a note on the
// board, and a prompt to Claude in a folder. list_notes and note_comment
// read and answer the notes that name the user; these start them, so an
// agent posts and asks as itself (docs/api.md).
func noteTools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	return []Tool{
		{
			Name: "post_note",
			Description: "Leave a note on the board as you, or a reply to one. A note that names a user with @ is an instruction to that user, " +
				"who picks it up with list_notes and reports with note_comment.",
			InputSchema: object(map[string]Property{
				"body":   str("The note, in markdown."),
				"parent": str("The ID of the note to reply to, from list_notes; none leaves a note of its own."),
			}, "body"),
			call: postNote,
		},
		{
			Name: "ask_prompt",
			Description: "Ask Claude a prompt in a folder, or a directory inside it, as you, as the Prompt box of its page does. " +
				"Claude Code runs there and answers in a reply; wait returns the answer. " +
				"With area workflows, it is the prompt of the Workflows page: Claude builds or changes the workflow the prompt describes, runs it and answers with its link, steps and last output.",
			InputSchema: object(map[string]Property{
				"folder": str("The folder, such as malta, or _ for the root, which area workflows takes unless given."),
				"dir":    str("A directory inside the folder, such as dockerfiles."),
				"prompt": str("What to ask."),
				"area":   {Type: "string", Description: "workflows asks as the Workflows page's prompt does.", Enum: []string{model.AreaWorkflows}},
				"wait":   {Type: "integer", Description: "Seconds to wait for the answer, up to 900; 0 returns once it is asked."},
			}, "prompt"),
			call: askPrompt,
		},
	}
}

type postArgs struct {
	Body   string `json:"body"`
	Parent string `json:"parent"`
}

func postNote(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in postArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if strings.TrimSpace(in.Body) == "" {
		return "", errors.New("a note needs a body")
	}
	data, err := c.call(ctx, http.MethodPost, "/notes", api.NoteRequest{ParentID: in.Parent, Body: in.Body})
	if err != nil {
		return "", err
	}
	var note model.Note
	if err := json.Unmarshal(data, &note); err != nil {
		return "", err
	}
	if note.ParentID != "" {
		return fmt.Sprintf("Replied to note %s as %s, reply %s.", note.ParentID, note.UserName, note.ID), nil
	}
	return fmt.Sprintf("Left note %s as %s.", note.ID, note.UserName), nil
}

type askArgs struct {
	Area   string `json:"area"`
	Folder string `json:"folder"`
	Dir    string `json:"dir"`
	Prompt string `json:"prompt"`
	Wait   int    `json:"wait"`
}

func askPrompt(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in askArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.Area != "" && in.Folder == "" {
		in.Folder = model.Default
	}
	if in.Folder == "" || strings.TrimSpace(in.Prompt) == "" {
		return "", errors.New("a folder and a prompt are needed")
	}
	data, err := c.call(ctx, http.MethodPost, "/prompts", api.PromptRequest{Area: in.Area, Folder: in.Folder, Dir: in.Dir, Prompt: in.Prompt})
	if err != nil {
		return "", err
	}
	var asked model.Note
	if err := json.Unmarshal(data, &asked); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("folder", in.Folder)
	q.Set("dir", in.Dir)
	if in.Area != "" {
		q = url.Values{"area": {in.Area}}
	}
	path := "/prompts?" + q.Encode()
	deadline := time.Now().Add(min(time.Duration(in.Wait)*time.Second, maxWait))
	thread := api.NoteView{Note: asked}
	for !thread.Finished() && time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollEvery):
		}
		var out noteList
		if err := c.get(ctx, path, &out); err != nil {
			return "", err
		}
		for _, view := range out.Notes {
			if view.ID == asked.ID {
				thread = view
			}
		}
	}
	return showPrompt(thread), nil
}

// showPrompt is a prompt, how it stands, and Claude's answer once there is one.
func showPrompt(thread api.NoteView) string {
	var b strings.Builder
	fmt.Fprintf(&b, "prompt %s in %s by %s, %s", thread.ID, thread.Dir, thread.UserName, thread.State)
	if thread.RunID != "" {
		fmt.Fprintf(&b, " (run %s)", thread.RunID)
	}
	b.WriteString("\n")
	if !thread.Finished() {
		b.WriteString("No answer yet.\n")
		return b.String()
	}
	for _, reply := range thread.Replies {
		fmt.Fprintf(&b, "\nreply %s by %s:\n%s\n", reply.ID, reply.UserName, reply.Body)
	}
	return b.String()
}

type notesArgs struct {
	Mention string `json:"mention"`
	State   string `json:"state"`
	Limit   int    `json:"limit"`
}

type noteList struct {
	Notes []api.NoteView `json:"notes"`
}

func listNotes(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	in := notesArgs{State: model.NoteOpen, Limit: 20}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	// The API takes * as any user and any state; empty mention is the
	// caller here.
	if in.Mention == "" {
		if user := access.UserFrom(ctx); user != nil {
			in.Mention = user.Name
		}
	}
	q := url.Values{}
	q.Set("mention", in.Mention)
	q.Set("state", in.State)
	q.Set("limit", fmt.Sprint(in.Limit))
	var out noteList
	if err := c.get(ctx, "/notes?"+q.Encode(), &out); err != nil {
		return "", err
	}
	if len(out.Notes) == 0 {
		return "No notes.", nil
	}
	shown := make([]string, 0, len(out.Notes))
	for _, view := range out.Notes {
		shown = append(shown, showNote(view))
	}
	return strings.Join(shown, "\n"), nil
}

// showNote is a note with its replies, as text.
func showNote(view api.NoteView) string {
	var b strings.Builder
	state := view.State
	if state == "" {
		state = "message"
	}
	fmt.Fprintf(&b, "note %s by %s, %s, %s\n%s\n", view.ID, view.UserName, state, when(view.CreatedAt), view.Body)
	if view.RunID != "" {
		fmt.Fprintf(&b, "(run %s)\n", view.RunID)
	}
	for _, reply := range view.Replies {
		fmt.Fprintf(&b, "  reply by %s, %s: %s\n", reply.UserName, when(reply.CreatedAt), strings.ReplaceAll(reply.Body, "\n", "\n    "))
	}
	return b.String()
}

// when is a time as the tools print it.
func when(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

type commentArgs struct {
	ID    string `json:"id"`
	Body  string `json:"body"`
	State string `json:"state"`
}

func noteComment(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in commentArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", errors.New("a note ID is needed")
	}
	if strings.TrimSpace(in.Body) == "" && in.State == "" {
		return "", errors.New("a reply, a state, or both are needed")
	}
	if _, err := c.call(ctx, http.MethodPost, "/notes/"+url.PathEscape(in.ID)+"/comments", api.CommentRequest{Body: in.Body, State: in.State}); err != nil {
		return "", err
	}
	text := "Replied to note " + in.ID
	if in.State != "" {
		text += "; it is " + in.State
	}
	return text + ".", nil
}
