package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/app/model"
)

// ErrNoteForbidden reports a spoken note from someone without note/edit.
var ErrNoteForbidden = errors.New("you need the note/edit permission to leave a note")

// untitledRecording is the body of a note that holds a clip and no words.
const untitledRecording = "A recording."

// Record leaves a spoken note on the board: the words in Body and Spoken, the recording in AudioPath, and no agent started.
func (b *Board) Record(ctx context.Context, user *model.User, parentID, spoken, audio string) (*model.Note, error) {
	if user == nil {
		return nil, fmt.Errorf("%w: a note needs an author", ErrInvalid)
	}
	if !b.permissions(ctx, user).Allowed("note/edit") {
		return nil, ErrNoteForbidden
	}
	spoken, audio = strings.TrimSpace(spoken), strings.TrimSpace(audio)
	if spoken == "" && audio == "" {
		return nil, fmt.Errorf("%w: a recording or its words are needed", ErrInvalid)
	}
	var parent *model.Note
	if parentID != "" {
		var err error
		if parent, err = b.Notes.Get(ctx, parentID); err != nil {
			return nil, err
		}
		if parent.ParentID != "" {
			return nil, fmt.Errorf("%w: a reply goes to a note, not to a reply", ErrInvalid)
		}
		if parent.IsPrompt() {
			return nil, fmt.Errorf("%w: a recording joins a note, not a prompt", ErrInvalid)
		}
	}
	body := spoken
	if body == "" {
		body = untitledRecording
	}
	note := &model.Note{
		Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name,
		Body: body, Spoken: spoken, AudioPath: audio,
	}
	if parent != nil {
		note.Channel = parent.Channel
	}
	if err := b.Notes.Create(ctx, note); err != nil {
		return nil, err
	}
	b.publish(note)
	return note, nil
}
