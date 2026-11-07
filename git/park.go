package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// parkWork parks the work a mend would write over, the dirty tree and the commits HEAD holds over base.
func (g *Git) parkWork(ctx context.Context, dir, stamp, base string, rec *Recovery, w io.Writer) error {
	out, err := g.capture(ctx, dir, localTimeout, statusArgs...)
	if err != nil {
		return err
	}
	status := ParseStatus(out)
	// An untracked file the mend writes over is staged first, ignored or not, so the parked commit holds it.
	for _, path := range g.collisions(ctx, dir, base, status) {
		if err := g.write(ctx, dir, w, localTimeout, "add", "--force", "--", path); err != nil {
			return err
		}
	}
	sha := ""
	if status.Dirty > 0 {
		if sha, err = g.capture(ctx, dir, localTimeout, "stash", "create", "task-ui parked "+stamp); err != nil {
			return err
		}
	}
	if sha == "" {
		if ahead, _ := g.count(ctx, dir, base+"..HEAD"); ahead > 0 {
			if sha, err = g.capture(ctx, dir, localTimeout, "rev-parse", "HEAD"); err != nil {
				return err
			}
		}
	}
	return g.park(ctx, dir, ParkPrefix+stamp, sha, rec, w)
}

// parkBranch parks the tip of a branch a mend moves, when it holds commits base does not.
func (g *Git) parkBranch(ctx context.Context, dir, stamp, branch, base string, rec *Recovery, w io.Writer) error {
	ref := "refs/heads/" + branch
	sha, err := g.capture(ctx, dir, localTimeout, "rev-parse", "--verify", ref)
	if err != nil {
		return nil
	}
	if ahead, err := g.count(ctx, dir, base+".."+ref); err != nil || ahead == 0 {
		return nil
	}
	return g.park(ctx, dir, ParkPrefix+stamp+"-"+strings.ReplaceAll(branch, "/", "-"), sha, rec, w)
}

// park writes ref over sha, which a reader reaches with git log, and names it in rec.
func (g *Git) park(ctx context.Context, dir, ref, sha string, rec *Recovery, w io.Writer) error {
	if sha == "" {
		return nil
	}
	if err := g.write(ctx, dir, w, localTimeout, "update-ref", ref, sha); err != nil {
		return err
	}
	rec.Parked = append(rec.Parked, ref)
	fmt.Fprintf(w, "parked %s at %s\n", sha[:min(len(sha), 8)], ref)
	return nil
}

// collisions returns the untracked files base holds too, the ones a mend writes over.
func (g *Git) collisions(ctx context.Context, dir, base string, status *Status) []string {
	var untracked []string
	for _, change := range status.Changes {
		if change.Index == "?" {
			untracked = append(untracked, change.Path)
		}
	}
	var found []string
	for chunk := range slices.Chunk(untracked, 64) {
		out, err := g.capture(ctx, dir, localTimeout, append([]string{"ls-tree", "-r", "--name-only", base, "--"}, chunk...)...)
		if err != nil || out == "" {
			continue
		}
		for _, path := range strings.Split(out, "\n") {
			// An untracked directory answers with the files under it, the ones only the remote has included.
			if _, err := os.Lstat(filepath.Join(dir, path)); err == nil {
				found = append(found, path)
			}
		}
	}
	return found
}
