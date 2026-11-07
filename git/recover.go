package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// localTimeout bounds a git command that reads the checkout, remoteTimeout one that reaches the remote.
const (
	localTimeout  = 10 * time.Second
	remoteTimeout = 60 * time.Second
)

// statusArgs read the checkout's state with paths a pathspec takes back.
var statusArgs = []string{"-c", "core.quotePath=false", "status", "--porcelain=v2", "--branch"}

// Recover mends the checkout in dir whose pull said failure so the next pull fast forwards, and reports what it did.
// With clone false it refuses the last resort, a fresh clone beside the checkout it renames, which no caller wants for a directory runs are living in.
func (g *Git) Recover(ctx context.Context, dir, failure string, clone bool, w io.Writer) (Recovery, error) {
	if !g.IsRepo(dir) {
		return Recovery{}, ErrNotRepo
	}
	stamp := time.Now().UTC().Format("20060102-150405.000")
	url := g.originURL(ctx, dir)
	if _, err := g.capture(ctx, dir, localTimeout, "rev-parse", "--git-dir"); err != nil {
		return g.reclone(ctx, dir, url, "", stamp, clone, Recovery{Case: CaseBroken, Done:"git does not open the checkout"}, w)
	}
	// The remote first: a pull nothing answered is no reason to touch the tree.
	remote, err := g.capture(ctx, dir, remoteTimeout, "ls-remote", "--symref", "origin", "HEAD")
	if err != nil {
		return Recovery{Case: CaseAuth, Done: "the remote did not answer, the checkout is untouched"}, nil
	}
	lock := indexLock(dir)
	if lock != "" {
		if holder := lockHolder(ctx, lock); holder != "" {
			return Recovery{Case: CaseLock, Done: holder + " holds the index lock, the checkout is untouched"}, nil
		}
		if err := os.Remove(lock); err != nil {
			return Recovery{Case: CaseLock}, err
		}
		fmt.Fprintf(w, "removed the stale %s\n", filepath.Join(".git", filepath.Base(lock)))
	}
	s, ok := g.read(ctx, dir, defaultBranch(remote))
	if !ok {
		return g.reclone(ctx, dir, url, "", stamp, clone, Recovery{Case: CaseBroken, Done:"git does not read the checkout's state"}, w)
	}
	return g.mend(ctx, dir, url, stamp, failure, s, lock != "", clone, w)
}

// mend runs the case's own repair over a checkout that reads, and clones the remote again for the rest.
func (g *Git) mend(ctx context.Context, dir, url, stamp, failure string, s state, lock, clone bool, w io.Writer) (Recovery, error) {
	rec := Recovery{Case: classify(s, lock), Branch: s.branch}
	if rec.Case == CaseUnknown && g.Triage != nil {
		if answer, err := g.Triage(ctx, dir, failure, s.status); err == nil && answer != "" && slices.Contains(Cases(), answer) {
			rec.Case, rec.Triaged = answer, true
		}
	}
	var err error
	switch {
	case rec.Case == CaseAuth:
		rec.Done = "the remote is the failure, the checkout is untouched"
		return rec, nil
	case rec.Case == CaseLock && lock:
		rec.Done = "removed the stale index lock"
	case rec.Case == CaseDirty || rec.Case == CaseDiverged:
		err = g.toUpstream(ctx, dir, stamp, s.upstream, &rec, w)
	case rec.Case == CaseDetached || rec.Case == CaseGone:
		err = g.toDefault(ctx, dir, stamp, s.def, &rec, w)
	default:
		// A case with no mend of its own is the last resort's.
		return g.reclone(ctx, dir, url, s.remote, stamp, clone, rec, w)
	}
	if err != nil {
		return rec, err
	}
	if rec.Mended = g.mended(ctx, dir); rec.Mended {
		return rec, nil
	}
	return g.reclone(ctx, dir, url, s.remote, stamp, clone, rec, w)
}

// state is what a checkout says about itself once its pull failed.
type state struct {
	status   string
	branch   string
	upstream string
	// remote is the branch as the remote still has it, "" when it is gone.
	remote string
	// def is the branch the remote's HEAD names.
	def    string
	ahead  int
	behind int
	// changed counts the tracked changes and the untracked files a mend would write over.
	changed int
}

// classify names which case a checkout's state is, CaseUnknown for none of them.
func classify(s state, lock bool) string {
	switch {
	case s.branch == "":
		return CaseDetached
	case s.upstream == "" || s.remote == "":
		return CaseGone
	case s.ahead > 0 && s.behind > 0:
		return CaseDiverged
	case s.changed > 0 && s.behind > 0:
		return CaseDirty
	case lock:
		return CaseLock
	}
	return CaseUnknown
}

// read returns what the checkout says of its state, and false when git will not say.
func (g *Git) read(ctx context.Context, dir, def string) (state, bool) {
	s := state{def: def}
	out, err := g.capture(ctx, dir, localTimeout, statusArgs...)
	if err != nil {
		return s, false
	}
	s.status = out
	status := ParseStatus(out)
	s.upstream = status.Upstream
	if ref, err := g.capture(ctx, dir, localTimeout, "symbolic-ref", "HEAD"); err == nil {
		s.branch = strings.TrimPrefix(ref, "refs/heads/")
	}
	if s.branch != "" {
		if heads, err := g.capture(ctx, dir, remoteTimeout, "ls-remote", "--heads", "origin", "refs/heads/"+s.branch); err == nil && heads != "" {
			s.remote = s.branch
		}
	}
	if s.upstream != "" {
		// A pull that died before its fetch leaves a checkout that reads as current and is not.
		_ = g.write(ctx, dir, io.Discard, remoteTimeout, "fetch", "origin")
		if s.ahead, err = g.count(ctx, dir, "@{u}..HEAD"); err != nil {
			s.upstream = ""
		}
		s.behind, _ = g.count(ctx, dir, "HEAD..@{u}")
	}
	s.changed = changed(status) + len(g.collisions(ctx, dir, "@{u}", status))
	return s, true
}

// changed counts the tracked files a checkout changed, the ones a reset throws away.
func changed(status *Status) int {
	n := 0
	for _, change := range status.Changes {
		if change.Index != "?" {
			n++
		}
	}
	return n
}

// toUpstream parks the work the checkout holds of its own and resets it onto the upstream.
func (g *Git) toUpstream(ctx context.Context, dir, stamp, upstream string, rec *Recovery, w io.Writer) error {
	if err := g.parkWork(ctx, dir, stamp, "@{u}", rec, w); err != nil {
		return err
	}
	if err := g.write(ctx, dir, w, localTimeout, "reset", "--hard", "@{u}"); err != nil {
		return err
	}
	rec.Done = "reset onto " + upstream
	return nil
}

// toDefault parks the work the mend would move and checks the remote's default branch out.
func (g *Git) toDefault(ctx context.Context, dir, stamp, def string, rec *Recovery, w io.Writer) error {
	if def == "" {
		return errors.New("the remote names no default branch to check out")
	}
	if err := g.write(ctx, dir, w, remoteTimeout, "fetch", "origin"); err != nil {
		return err
	}
	start := "refs/remotes/origin/" + def
	if _, err := g.capture(ctx, dir, localTimeout, "rev-parse", "--verify", start); err != nil {
		return fmt.Errorf("the remote's %s is not in the checkout: %w", def, err)
	}
	if err := g.parkWork(ctx, dir, stamp, start, rec, w); err != nil {
		return err
	}
	if err := g.parkBranch(ctx, dir, stamp, def, start, rec, w); err != nil {
		return err
	}
	if err := g.write(ctx, dir, w, localTimeout, "checkout", "--force", "-B", def, start); err != nil {
		return err
	}
	rec.Branch, rec.Done = def, "checked "+def+" out, the remote's default"
	return nil
}

// mended reports whether a pull fast forwards now: an upstream, nothing ahead of it, nothing in its way.
func (g *Git) mended(ctx context.Context, dir string) bool {
	ahead, err := g.count(ctx, dir, "@{u}..HEAD")
	if err != nil || ahead > 0 {
		return false
	}
	out, err := g.capture(ctx, dir, localTimeout, statusArgs...)
	if err != nil {
		return false
	}
	status := ParseStatus(out)
	return changed(status) == 0 && len(g.collisions(ctx, dir, "@{u}", status)) == 0
}

// reclone keeps the broken checkout under a stamped name and clones the remote again in its place.
func (g *Git) reclone(ctx context.Context, dir, url, branch, stamp string, clone bool, rec Recovery, w io.Writer) (Recovery, error) {
	if !clone {
		rec.Done = "only a fresh clone mends this one, which the caller refuses for this checkout"
		return rec, nil
	}
	if url == "" {
		return rec, errors.New("the checkout has no origin to clone again")
	}
	moved := dir + ".broken-" + stamp
	if err := os.Rename(dir, moved); err != nil {
		return rec, err
	}
	fmt.Fprintf(w, "kept the broken checkout as %s\n", filepath.Base(moved))
	out, err := g.Clone(ctx, url, dir, branch)
	fmt.Fprintln(w, out)
	if err != nil {
		if back := os.Rename(moved, dir); back != nil {
			return rec, fmt.Errorf("%w, and it stays at %s: %w", err, filepath.Base(moved), back)
		}
		return rec, err
	}
	rec.Moved, rec.Mended, rec.Done = moved, true, "cloned the remote again"
	rec.Branch, _ = g.capture(ctx, dir, localTimeout, "symbolic-ref", "--short", "HEAD")
	return rec, nil
}

// originURL returns the checkout's origin, out of the config file when git will not open the repository.
func (g *Git) originURL(ctx context.Context, dir string) string {
	if url, err := g.capture(ctx, dir, localTimeout, "remote", "get-url", "origin"); err == nil && url != "" {
		return url
	}
	url, err := g.capture(ctx, dir, localTimeout, "config", "--file", filepath.Join(dir, ".git", "config"), "--get", "remote.origin.url")
	if err != nil {
		return ""
	}
	return url
}

// defaultBranch returns the branch ls-remote --symref names as the remote's HEAD, "" for none.
func defaultBranch(out string) string {
	for _, line := range strings.Split(out, "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "ref: ")
		if !ok {
			continue
		}
		ref, _, _ := strings.Cut(rest, "\t")
		if branch, ok := strings.CutPrefix(ref, "refs/heads/"); ok {
			return branch
		}
	}
	return ""
}

// count returns the commits in a range, and an error for a range git will not resolve.
func (g *Git) count(ctx context.Context, dir, span string) (int, error) {
	out, err := g.capture(ctx, dir, localTimeout, "rev-list", "--count", span)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(out)
}
