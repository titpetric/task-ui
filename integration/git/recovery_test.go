package git

import (
	"slices"
	"strings"
	"testing"
)

func TestCases(t *testing.T) {
	cases := Cases()
	for _, want := range []string{CaseDirty, CaseDiverged, CaseGone, CaseDetached, CaseLock, CaseAuth, CaseBroken, CaseUnknown} {
		if !slices.Contains(cases, want) {
			t.Errorf("%q is in no case", want)
		}
	}
	if len(cases) != 8 {
		t.Errorf("%d cases, want the eight a triage answers from", len(cases))
	}
}

func TestRecovery_Line(t *testing.T) {
	rec := Recovery{Case: CaseDiverged, Done: "reset onto origin/main", Parked: []string{ParkPrefix + "20260101-120000.000"}, Branch: "main", Mended: true}
	line := rec.Line()
	for _, want := range []string{CaseDiverged, "reset onto origin/main", ParkPrefix} {
		if !strings.Contains(line, want) {
			t.Errorf("the line %q says nothing of %q", line, want)
		}
	}
	triaged := Recovery{Case: CaseAuth, Triaged: true, Moved: "/root/project.broken-20260101-120000.000"}.Line()
	if !strings.Contains(triaged, "(triaged)") || !strings.Contains(triaged, "project.broken-20260101-120000.000") {
		t.Errorf("the line %q says nothing of the triage or the kept checkout", triaged)
	}
}
