package tooling

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Prober runs `atkins --stats --json --all` in a directory and keeps
// what it printed for a TTL, or until a database it read from changes.
//
// A probe that fails, times out, or prints something that is not a
// report answers nothing: the page it feeds is a panel beside the runs,
// and a checkout with no atkins gets a page with no panel, one that
// still loads.
type Prober struct {
	// Command is the atkins binary, atkins unless set.
	Command string

	// Timeout bounds one probe, 5s unless set; TTL keeps an answer, 60s
	// unless set.
	//
	// The window is a month of history and a page anyone may reload reads
	// the report, so a report outlives the listing of a folder's jobs in
	// the cache.
	Timeout time.Duration
	TTL     time.Duration

	// Output runs atkins in dir with args and returns what it printed,
	// exec unless set.
	Output func(ctx context.Context, dir string, args ...string) ([]byte, error)

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
	if ok && time.Since(cached.at) < p.ttl() && !changed(cached) {
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

// changed reports whether a database the probe read from changed
// since: a run that ends records to it, and the panel shows
// that run on the next page without waiting out the TTL.
func changed(cached probe) bool {
	for _, scope := range cached.report.Scopes() {
		info, err := os.Stat(scope.Database)
		if err == nil && info.ModTime().After(cached.at) {
			return true
		}
	}
	return false
}

// run asks atkins in dir for the report with every command listed, and
// without them from an atkins that refuses the flag.
func (p *Prober) run(dir string) *Report {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout())
	defer cancel()

	for _, args := range [][]string{{"--stats", "--json", "--all"}, {"--stats", "--json"}} {
		out, err := p.output()(ctx, dir, args...)
		if err != nil {
			continue
		}
		var report Report
		if err := json.Unmarshal(out, &report); err != nil {
			return nil
		}
		return &report
	}
	return nil
}

func (p *Prober) output() func(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if p.Output != nil {
		return p.Output
	}
	return p.exec
}

// exec runs atkins in dir. Its stdin is /dev/null: atkins reads a pipe
// on stdin as a pipeline.
func (p *Prober) exec(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, p.Bin(), args...)
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.WaitDelay = time.Second

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
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
