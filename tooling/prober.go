package tooling

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"sync"
	"time"
)

// Prober runs `atkins --stats --json` in a directory and keeps the
// answer for a TTL.
//
// A probe that fails, times out, or prints something that is not a
// report answers nothing: the page it feeds is a panel beside the runs,
// and a checkout with no atkins is a checkout with no panel rather than
// a page that will not load.
type Prober struct {
	// Command is the atkins binary, atkins unless set.
	Command string

	// Timeout bounds one probe, 5s unless set; TTL keeps an answer, 60s
	// unless set.
	//
	// The window is a month of history and the report is read by a page
	// anyone may reload, so the answer is worth keeping longer than the
	// listing of a folder's jobs is.
	Timeout time.Duration
	TTL     time.Duration

	mu    sync.Mutex
	cache map[string]probe
}

// probe is what a directory answered, and when.
type probe struct {
	report *Report
	at     time.Time
}

// Probe returns what atkins has recorded for dir. Nil when atkins is
// missing, fails, or reports nothing.
func (p *Prober) Probe(dir string) *Report {
	if p == nil || dir == "" {
		return nil
	}

	p.mu.Lock()
	cached, ok := p.cache[dir]
	p.mu.Unlock()
	if ok && time.Since(cached.at) < p.ttl() {
		return cached.report
	}

	report := p.run(dir)
	p.mu.Lock()
	if p.cache == nil {
		p.cache = map[string]probe{}
	}
	p.cache[dir] = probe{report: report, at: time.Now()}
	p.mu.Unlock()

	return report
}

// run execs atkins in dir. Its stdin is /dev/null: atkins reads a pipe
// on stdin as a pipeline.
func (p *Prober) run(dir string) *Report {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout())
	defer cancel()

	cmd := exec.CommandContext(ctx, p.Bin(), "--stats", "--json")
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.WaitDelay = time.Second

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}

	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		return nil
	}
	return &report
}

// Bin is the atkins the probe runs: Command, or atkins from the PATH.
func (p *Prober) Bin() string {
	if p == nil || p.Command == "" {
		return "atkins"
	}
	return p.Command
}

func (p *Prober) timeout() time.Duration {
	if p.Timeout > 0 {
		return p.Timeout
	}
	return 5 * time.Second
}

func (p *Prober) ttl() time.Duration {
	if p.TTL > 0 {
		return p.TTL
	}
	return time.Minute
}
