package runner

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// BacklogSetting names the instance setting that holds the backlog
// length, empty for the automatic one.
const BacklogSetting = "runs.backlog_length"

// BacklogPerCPU is how many runs wait for a slot per CPU when no admin
// sets a backlog length: 80 on a host with 4 CPUs.
const BacklogPerCPU = 20

// MaxBacklog is the longest backlog a setting takes.
const MaxBacklog = 100000

// ErrBacklogFull refuses a run that would wait for a slot behind a full
// line.
var ErrBacklogFull = errors.New("the backlog is full")

// ErrBacklogInvalid refuses a backlog length that is no positive whole
// number up to MaxBacklog.
var ErrBacklogInvalid = fmt.Errorf("the backlog length is a whole number from 1 to %d, or empty for the automatic one", MaxBacklog)

// SettingValues are the instance settings a backlog reads and writes.
type SettingValues interface {
	Values(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, values map[string]string) error
}

// Backlog is how many runs may wait for a free slot: the length an
// admin sets on the Instance page, else BacklogPerCPU times the CPUs.
// The manager reads it on every start, so a change applies at once.
type Backlog struct {
	Settings SettingValues
	// CPUs is the count the automatic length takes, 0 for the machine's.
	CPUs int
}

// Auto is the automatic length: BacklogPerCPU times the CPUs.
func (b *Backlog) Auto() int {
	cpus := runtime.NumCPU()
	if b != nil && b.CPUs > 0 {
		cpus = b.CPUs
	}
	return BacklogPerCPU * cpus
}

// Configured is the length an admin set, 0 for none.
func (b *Backlog) Configured(ctx context.Context) int {
	if b == nil || b.Settings == nil {
		return 0
	}
	values, err := b.Settings.Values(ctx)
	if err != nil {
		return 0
	}
	n, err := ParseBacklog(values[BacklogSetting])
	if err != nil {
		return 0
	}
	return n
}

// Length is the length in effect: the configured one, else the
// automatic one.
func (b *Backlog) Length(ctx context.Context) int {
	if n := b.Configured(ctx); n > 0 {
		return n
	}
	return b.Auto()
}

// Set stores a length, as the page sends it; empty or 0 is the
// automatic one.
func (b *Backlog) Set(ctx context.Context, value string) error {
	n, err := ParseBacklog(value)
	if err != nil {
		return err
	}
	if b == nil || b.Settings == nil {
		return errors.New("the instance settings aren't available")
	}
	stored := ""
	if n > 0 {
		stored = strconv.Itoa(n)
	}
	return b.Settings.Set(ctx, map[string]string{BacklogSetting: stored})
}

// ParseBacklog reads a backlog length: 0 for empty or 0, the automatic
// one, else a whole number from 1 to MaxBacklog.
func ParseBacklog(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 || n > MaxBacklog {
		return 0, ErrBacklogInvalid
	}
	return n, nil
}

// backlogFull is the error of a start the full line refuses.
func backlogFull(waiting int) error {
	return fmt.Errorf("%w: %d runs wait for a free slot already; start it again later", ErrBacklogFull, waiting)
}
