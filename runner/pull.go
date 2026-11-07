package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/titpetric/task-ui/git"
)

// checkout returns the checkout the run's folder belongs to: the folder
// itself, or the nearest parent inside the root that is one. "" for none.
func (m *Manager) checkout(session *Session) string {
	if m.opts.Git == nil {
		return ""
	}
	return m.opts.Git.Checkout(session.Folder.Dir, session.Folder.Depth())
}

// pull brings the checkouts of the run up to date, in the run's own
// terminal: the workspace's home project first, so its tooling is
// current, then the folder's when that differs; a folder inside a
// checkout pulls the checkout. Runs of one checkout pull one at a time: two pulls at once
// fight over the index lock, and a workflow starts them one after another.
// A shell pulls nothing: it opens on the checkout as it stands, the place
// to mend one a pull fails in. Nor does a host job, which runs on a host.
func (m *Manager) pull(ctx context.Context, session *Session) error {
	if m.opts.Git == nil || session.Job.IsShell() || session.Job.IsHost() {
		return nil
	}
	home := session.home
	if home != "" && m.opts.Git.IsRepo(home) {
		// The home is where every run of the workspace lives, this one
		// included, so the recovery may not replace it with a fresh clone.
		if err := m.pullDir(ctx, session, home, "git -C "+home+" pull --ff-only", false); err != nil {
			return err
		}
	}
	dir := m.checkout(session)
	if dir == "" || dir == home {
		return nil
	}
	return m.pullDir(ctx, session, dir, "git pull --ff-only", true)
}

// pullDir pulls one checkout under its lock, showing the command line.
func (m *Manager) pullDir(ctx context.Context, session *Session, dir, line string, clone bool) error {
	lock := m.pullLock(dir)
	lock.Lock()
	defer lock.Unlock()

	fmt.Fprintf(session, "\x1b[2m$ %s\x1b[0m\r\n", line)
	if err := pullOnce(ctx, m.opts.Git, dir, clone, crlfWriter{session}); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	fmt.Fprint(session, "\r\n")
	return nil
}

// puller is the git a pull needs: the fast forward, and the mend of a checkout it fails in.
type puller interface {
	Pull(ctx context.Context, dir string, w io.Writer) error
	Recover(ctx context.Context, dir, failure string, clone bool, w io.Writer) (git.Recovery, error)
}

// pullOnce fast forwards a checkout, and mends it and pulls again once when the pull fails; the pull's error is the one it returns.
// clone false keeps the last resort off a directory runs live in, a workspace home, which it would rename out from under them.
func pullOnce(ctx context.Context, g puller, dir string, clone bool, w io.Writer) error {
	said := newRing(4096)
	err := g.Pull(ctx, dir, io.MultiWriter(w, said))
	if err == nil || errors.Is(err, git.ErrNotRepo) {
		return err
	}
	recovery, mendErr := g.Recover(ctx, dir, failureOf(err, said), clone, w)
	if mendErr != nil {
		fmt.Fprintf(w, "git recovery: the checkout was not mended: %s\n", mendErr)
		return err
	}
	fmt.Fprintf(w, "%s\n", recovery.Line())
	if !recovery.Mended {
		return err
	}
	// The cause the reader needs is the pull's, not the retry's.
	if retry := g.Pull(ctx, dir, w); retry != nil {
		return err
	}
	return nil
}

// failureOf returns what the failed pull said, its exit status when it said nothing.
func failureOf(err error, said *ring) string {
	if out := strings.TrimSpace(string(said.Bytes())); out != "" {
		return out
	}
	return err.Error()
}

// PullCheckout fast forwards a checkout under its lock and mends it when the pull fails, writing to w what a push hook answers with; git.ErrNotRepo for no checkout.
// A hook pulls with nobody watching and names a workspace home as readily as a project, so this path refuses the last resort: a run's own pull is where a checkout gets cloned again.
func (m *Manager) PullCheckout(ctx context.Context, dir string, w io.Writer) error {
	if m.opts.Git == nil {
		return git.ErrNotRepo
	}
	lock := m.pullLock(dir)
	lock.Lock()
	defer lock.Unlock()
	return pullOnce(ctx, m.opts.Git, dir, false, w)
}

// pullLock returns the lock of a checkout.
func (m *Manager) pullLock(dir string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pulls == nil {
		m.pulls = map[string]*sync.Mutex{}
	}
	lock, ok := m.pulls[dir]
	if !ok {
		lock = &sync.Mutex{}
		m.pulls[dir] = lock
	}
	return lock
}
