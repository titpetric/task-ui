package runner

import (
	"errors"
	"strings"
	"time"

	"github.com/titpetric/task-ui/hours"
)

// The errors the hours gate ends a start in.
var (
	// ErrOutOfHours refuses a run when the job's hours do not allow one
	// now and the request does not force.
	ErrOutOfHours = errors.New("out of hours")
	// ErrReasonNeeded refuses a forced run out of hours that does not
	// say why.
	ErrReasonNeeded = errors.New("a reason is needed to run out of hours")
)

// gate checks the hours of the job: inside them a run goes ahead; outside
// them it needs Force and a reason, which the run records. Scheduled runs
// are not gated: the schedule is the decision.
func (m *Manager) gate(req Request) (bool, error) {
	if req.Scheduled {
		return false, nil
	}
	rules, err := hours.Parse(req.Folder.HoursOf(req.Job))
	if err != nil {
		return false, err
	}
	if rules.Allows(m.now()) {
		return false, nil
	}
	if !req.Force {
		return true, ErrOutOfHours
	}
	if strings.TrimSpace(req.Reason) == "" {
		return true, ErrReasonNeeded
	}
	return true, nil
}

// now is the clock, which tests move.
func (m *Manager) now() time.Time {
	if m.opts.Now != nil {
		return m.opts.Now()
	}
	return time.Now()
}
