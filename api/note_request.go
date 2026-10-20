package api

// NoteRequest leaves a note on a channel, notes or chat, or a reply to a
// note when ParentID names one.
type NoteRequest struct {
	Channel  string `json:"channel"`
	ParentID string `json:"parent_id"`
	Body     string `json:"body"`
}
