package api

// ThreadRequest posts a chat message: to a user when To names one, else a reply in the thread ParentID names.
type ThreadRequest struct {
	To       string `json:"to"`
	ParentID string `json:"parent_id"`
	Body     string `json:"body"`
	Spoken   string `json:"spoken"`
	Audio    string `json:"audio"`
}
