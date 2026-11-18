package storage

// ConversationFilter narrows the thread feed. Zero values match everything.
type ConversationFilter struct {
	WorkspaceID string
	// Folder matches a place and the directories under it, as Note.Dir holds it.
	Folder   string
	UserName string
	State    string
	Query    string
	// Live keeps open and working threads, and the finished ones of the last hour.
	Live  bool
	Limit int
}
