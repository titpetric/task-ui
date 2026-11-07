package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/git"
)

// fakeGit answers a pull with the failures it was given, in order, and a recovery of its own.
type fakeGit struct {
	fails    []error
	says     string
	recovery git.Recovery
	mendErr  error
	pulls    int
	mends    int
	failure  string
	cloned   bool
}

func (f *fakeGit) Pull(_ context.Context, _ string, w io.Writer) error {
	f.pulls++
	fmt.Fprint(w, f.says)
	if f.pulls <= len(f.fails) {
		return f.fails[f.pulls-1]
	}
	return nil
}

func (f *fakeGit) Recover(_ context.Context, _, failure string, clone bool, _ io.Writer) (git.Recovery, error) {
	f.cloned = clone
	f.mends++
	f.failure = failure
	return f.recovery, f.mendErr
}

func TestPullOnce(t *testing.T) {
	broken := errors.New("exit status 1")
	mended := git.Recovery{Case: git.CaseDirty, Done: "reset onto origin/main", Parked: []string{git.ParkPrefix + "20260101-120000.000"}, Mended: true}
	for name, test := range map[string]struct {
		git   *fakeGit
		err   error
		pulls int
		mends int
		says  string
	}{
		"a pull that works is one pull": {
			git: &fakeGit{}, pulls: 1,
		},
		"a folder that is no checkout is no recovery": {
			git: &fakeGit{fails: []error{git.ErrNotRepo}}, err: git.ErrNotRepo, pulls: 1,
		},
		"a mend and one retry": {
			git: &fakeGit{fails: []error{broken}, recovery: mended}, pulls: 2, mends: 1, says: git.ParkPrefix,
		},
		"a retry that fails reports the pull": {
			git: &fakeGit{fails: []error{broken, broken}, recovery: mended}, err: broken, pulls: 2, mends: 1,
		},
		"a case no mend fixes is not retried": {
			git: &fakeGit{fails: []error{broken}, recovery: git.Recovery{Case: git.CaseAuth}}, err: broken, pulls: 1, mends: 1, says: git.CaseAuth,
		},
		"a recovery that errors reports the pull": {
			git: &fakeGit{fails: []error{broken}, mendErr: errors.New("no origin")}, err: broken, pulls: 1, mends: 1, says: "no origin",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var out strings.Builder
			err := pullOnce(context.Background(), test.git, "/root/project", true, &out)
			if !errors.Is(err, test.err) {
				t.Errorf("the pull answered %v, want %v", err, test.err)
			}
			if test.git.pulls != test.pulls || test.git.mends != test.mends {
				t.Errorf("%d pulls and %d mends, want %d and %d", test.git.pulls, test.git.mends, test.pulls, test.mends)
			}
			if test.says != "" && !strings.Contains(out.String(), test.says) {
				t.Errorf("the output %q says nothing of %q", out.String(), test.says)
			}
		})
	}
}

func TestPullOnce_Failure(t *testing.T) {
	fake := &fakeGit{fails: []error{errors.New("exit status 1")}, says: "error: Your local changes would be overwritten\n"}
	var out strings.Builder
	_ = pullOnce(context.Background(), fake, "/root/project", false, &out)
	if !strings.Contains(fake.failure, "local changes would be overwritten") {
		t.Errorf("the mend was told %q, want what the pull said", fake.failure)
	}
	if !strings.Contains(out.String(), "local changes would be overwritten") {
		t.Errorf("the output %q leaves out what the pull said", out.String())
	}
	if fake.cloned {
		t.Error("the mend was allowed the last resort, which this caller refuses")
	}
}

func TestFailureOf(t *testing.T) {
	said := newRing(4096)
	if got := failureOf(errors.New("exit status 128"), said); got != "exit status 128" {
		t.Errorf("a silent pull reads as %q, want its exit status", got)
	}
	fmt.Fprint(said, "fatal: not possible to fast-forward\n")
	if got := failureOf(errors.New("exit status 128"), said); got != "fatal: not possible to fast-forward" {
		t.Errorf("the failure reads as %q, want what git said", got)
	}
}
