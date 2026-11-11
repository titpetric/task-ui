package git

import (
	"path/filepath"
	"strings"
)

// The cases a failed pull is, the closed list a triage answers from.
const (
	CaseDirty    = "dirty"
	CaseDiverged = "diverged"
	CaseGone     = "upstream-gone"
	CaseDetached = "detached"
	CaseLock     = "index-lock"
	CaseAuth     = "auth"
	CaseBroken   = "broken"
	CaseUnknown  = "unknown"
)

// ParkPrefix starts the refs a recovery keeps what it would otherwise lose under.
const ParkPrefix = "refs/task-ui/parked/"

// Cases returns the case names a recovery reports and a triage answers with.
func Cases() []string {
	return []string{CaseDirty, CaseDiverged, CaseGone, CaseDetached, CaseLock, CaseAuth, CaseBroken, CaseUnknown}
}

// Recovery is what Recover found in a checkout whose pull failed and did about it.
type Recovery struct {
	// Case is which failure it was, one of Cases.
	Case string `json:"case"`
	// Done says in a phrase what the mend ran.
	Done string `json:"done"`
	// Parked are the refs holding the work that was at risk, in the checkout Moved names once that is set.
	Parked []string `json:"parked,omitempty"`
	// Branch is the branch the checkout sits on after the mend.
	Branch string `json:"branch,omitempty"`
	// Moved is where the last resort kept the broken checkout.
	Moved string `json:"moved,omitempty"`
	// Mended reports whether a pull fast forwards now, so a caller retries it.
	Mended bool `json:"mended"`
	// Triaged says a model named the case the rules could not.
	Triaged bool `json:"triaged,omitempty"`
}

// Line returns the one line a run's terminal and a hook's answer show of a recovery.
func (r Recovery) Line() string {
	line := "git recovery: " + r.Case
	if r.Triaged {
		line += " (triaged)"
	}
	if r.Done != "" {
		line += ": " + r.Done
	}
	if len(r.Parked) > 0 {
		line += "; parked at " + strings.Join(r.Parked, " ")
	}
	if r.Moved != "" {
		line += "; the old checkout is " + filepath.Base(r.Moved)
	}
	return line
}
