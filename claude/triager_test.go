package claude

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/git"
)

func TestCaseOf(t *testing.T) {
	for answer, want := range map[string]string{
		"dirty":                        git.CaseDirty,
		"  Diverged\n":                 git.CaseDiverged,
		"`upstream-gone`":              git.CaseGone,
		"index-lock.":                  git.CaseLock,
		"the reflog says it diverged":  "",
		"":                             "",
		"rm -rf /; echo upstream-gone": "",
	} {
		if got := caseOf(answer); got != want {
			t.Errorf("%q names %q, want %q", answer, got, want)
		}
	}
}

func TestPullTriage(t *testing.T) {
	for _, name := range git.Cases() {
		if !strings.Contains(PullTriage, name) {
			t.Errorf("the prompt leaves out the case %q", name)
		}
	}
}
