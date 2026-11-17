package api

// VoiceNoteRequest leaves a spoken note on the board: the transcript and the clip it was heard in.
type VoiceNoteRequest struct {
	ParentID string `json:"parent_id"`
	Spoken   string `json:"spoken"`
	Audio    string `json:"audio"`
}
