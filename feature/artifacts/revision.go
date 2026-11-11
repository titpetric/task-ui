package artifactstore

// Revision is one commit of the history as a reader of it needs it: what
// to name it by, when it landed, the message the writer left, and the
// paths it touched.
type Revision struct {
	// Commit is the full name; Short is the seven the history prints.
	Commit string
	Short  string
	// When is the author date, as git formats it for a reader: relative,
	// because a revision is read against the ones around it and "18
	// minutes ago" says more here than a timestamp.
	When string
	// Subject is the first line of the message. A run's ingest names the
	// folder, the job and the run; a single write names the path.
	Subject string
	// Files are the paths the commit changed, relative to the tree.
	Files []string
}
