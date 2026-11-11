package api

// CommentRequest replies to a note and, with a State, moves its
// instruction to working, done or failed.
type CommentRequest struct {
	Body  string `json:"body"`
	State string `json:"state"`
}
