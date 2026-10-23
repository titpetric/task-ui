package api

// MeView is what the user of a request may do, for an agent that can't
// tell without trying: the user and role, the paths of
// `config/permissions.yml` the user may take, as a list and as a line,
// each area of task-ui with its purpose and what the user may do there,
// and the folders the user sees, the first MaxFolders of them.
type MeView struct {
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	Admin   bool     `json:"admin"`
	Allowed []string `json:"allowed"`
	Summary string   `json:"summary"`
	// Sections are the sections of permissions.yml, then its
	// application permissions.
	Sections    []AreaView `json:"sections"`
	Permissions []AreaView `json:"permissions"`
	// Keys are the SSH keys the user may read over MCP, key/read, by
	// name.
	Keys        []string `json:"keys"`
	FolderCount int      `json:"folder_count"`
	Folders     []string `json:"folders"`
}

// MaxFolders bounds the folders a MeView lists.
const MaxFolders = 100
