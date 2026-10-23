package artifactstore

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
)

// Expire drops the history before a commit: that commit's tree becomes
// the first commit, the commits after it follow with their trees,
// messages, authors and times, and what only the dropped commits held
// goes. It calls then with what it did while it still holds the tree,
// before the dropped commits go, so the records follow before anything
// else writes; an error from then leaves the history as it stood. A commit the
// history doesn't hold is ErrNotFound; the first commit changes nothing.
func (s *Store) Expire(ctx context.Context, keep string, then func(Rewrite) error) error {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	defer unlock()
	commits, err := s.commits(ctx)
	if err != nil {
		return err
	}
	at := slices.Index(commits, keep)
	if at < 0 {
		return ErrNotFound
	}
	if at == 0 {
		return nil
	}
	rewrite := Rewrite{Renamed: map[string]string{}, Dropped: commits[:at], store: s, ctx: ctx}
	parent := ""
	for _, commit := range commits[at:] {
		renamed, err := s.copyCommit(ctx, commit, parent)
		if err != nil {
			return err
		}
		rewrite.Renamed[commit] = renamed
		parent = renamed
	}
	rewrite.Root = rewrite.Renamed[keep]
	head := commits[len(commits)-1]
	if err := s.point(ctx, parent, head); err != nil {
		return err
	}
	// The dropped commits are still there for then to compare with;
	// should then fail, main goes back to them.
	if then != nil {
		if err := then(rewrite); err != nil {
			return errors.Join(err, s.point(ctx, head, parent))
		}
	}
	return s.prune(ctx)
}

// copyCommit makes a commit of the same tree, message, author and times
// as one of the history, on a parent, none for a root.
func (s *Store) copyCommit(ctx context.Context, commit, parent string) (string, error) {
	meta, err := s.git(ctx, "log", "-1", "--format=%T%x00%an%x00%ae%x00%aI%x00%cI%x00%B", commit)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(meta, "\x00", 6)
	if len(parts) < 6 {
		return "", ErrNotFound
	}
	args := []string{"commit-tree", parts[0], "-m", parts[5]}
	if parent != "" {
		args = append(args, "-p", parent)
	}
	cmd := s.command(ctx, args...)
	cmd.Env = append(cmd.Env,
		"GIT_AUTHOR_NAME="+parts[1], "GIT_AUTHOR_EMAIL="+parts[2], "GIT_AUTHOR_DATE="+parts[3],
		"GIT_COMMITTER_DATE="+parts[4],
	)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
