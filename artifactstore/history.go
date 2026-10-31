package artifactstore

import (
	"context"
	"strconv"
	"strings"
)

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

// History is the recent commits that touched a path, newest first, with
// the paths each one changed. An empty path is the whole tree.
//
// This is what makes a rollback possible without a checkout: a reader
// takes a commit from here and reads a file as it stood with Show, which
// is the pair an agent needs to put a file back the way it was.
//
// A store with no repository yet, or a path no commit touched, is no
// history and no error: a thread whose first run has not finished has
// nothing to roll back to.
func (s *Store) History(ctx context.Context, rel string, limit int) ([]Revision, error) {
	if limit <= 0 {
		return nil, nil
	}
	clean := ""
	if rel != "" {
		var err error
		if clean, err = cleanPath(rel); err != nil {
			return nil, ErrNotFound
		}
	}
	unlock, err := s.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if !s.exists() || s.head(ctx) == "" {
		return nil, nil
	}

	// The record separator keeps a subject holding a newline from
	// reading as the start of the next revision; the field separator is
	// a byte no path and no subject carries.
	args := []string{
		"log", "--max-count=" + strconv.Itoa(limit), "--name-only",
		"--format=%x1e%H%x1f%h%x1f%ar%x1f%s",
	}
	if clean != "" {
		args = append(args, "--", clean)
	}
	out, err := s.git(ctx, args...)
	if err != nil {
		return nil, err
	}
	return parseHistory(out), nil
}

// parseHistory reads what the log printed: a record per commit, its four
// fields, then the paths it changed, one to a line.
func parseHistory(out string) []Revision {
	var revisions []Revision
	for _, record := range strings.Split(out, "\x1e") {
		record = strings.TrimLeft(record, "\n")
		if record == "" {
			continue
		}
		head, rest, _ := strings.Cut(record, "\n")
		fields := strings.Split(head, "\x1f")
		if len(fields) < 4 {
			continue
		}
		revision := Revision{Commit: fields[0], Short: fields[1], When: fields[2], Subject: fields[3]}
		for _, line := range strings.Split(rest, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				revision.Files = append(revision.Files, line)
			}
		}
		revisions = append(revisions, revision)
	}
	return revisions
}
