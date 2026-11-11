package api

// NoteRequest leaves a note on the board, or a reply to a note when
// ParentID names one.
type NoteRequest struct {
	ParentID string `json:"parent_id"`
	Body     string `json:"body"`
}
