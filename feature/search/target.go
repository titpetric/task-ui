package search

// Target is one job a project declares, with the text that describes it.
type Target struct {
	FolderID    string
	FolderTitle string
	Workspace   string
	Project     string
	Dir         string
	ID          string
	Name        string
	Kind        string
	SourceFile  string
	Description string
	Summary     string
	Steps       []string
	Aliases     []string
	Command     []string
}
