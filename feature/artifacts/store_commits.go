package artifactstore

import (
	"context"
	"strings"
)

// Commits returns the commits of the history, oldest first; none before
// the first Ingest.
func (s *Store) Commits(ctx context.Context) ([]string, error) {
	unlock, err := s.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	return s.commits(ctx)
}

func (s *Store) commits(ctx context.Context) ([]string, error) {
	if !s.exists() || s.head(ctx) == "" {
		return nil, nil
	}
	out, err := s.git(ctx, "rev-list", "--reverse", "HEAD")
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}
