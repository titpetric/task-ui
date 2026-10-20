package sshprobe

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// maxOutput bounds what a run keeps of a host's output.
const maxOutput = 256 << 10

// limited keeps the first max bytes written to it and drops the rest.
type limited struct {
	buf bytes.Buffer
	max int
}

func (l *limited) Write(p []byte) (int, error) {
	if room := l.max - l.buf.Len(); room > 0 {
		if len(p) > room {
			l.buf.Write(p[:room])
		} else {
			l.buf.Write(p)
		}
	}
	return len(p), nil
}

// String is what the limit kept.
func (l *limited) String() string {
	return l.buf.String()
}

// Run runs a command on each host over its login, all at once, and
// returns the results in the order of the hosts. A host without a login
// in logins gets an error and no run. Each run ends at timeout.
func (p *Prober) Run(ctx context.Context, hosts []string, logins map[string]string, command string, timeout time.Duration) []Result {
	results := make([]Result, len(hosts))
	var wg sync.WaitGroup
	for i, host := range hosts {
		results[i] = Result{Host: host, Login: logins[host]}
		if results[i].Login == "" {
			results[i].Error = "task-ui has no ssh login to this host"
			results[i].ExitCode = -1
			continue
		}
		wg.Add(1)
		go func(r *Result) {
			defer wg.Done()
			run, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			out := &limited{max: maxOutput}
			cmd := p.terminal(run, r.Login, command)
			cmd.Stdout, cmd.Stderr = out, out
			// A remote command's children may hold the output open after
			// the time limit ends ssh; stop reading a second later.
			cmd.WaitDelay = time.Second
			start := time.Now()
			err := cmd.Run()
			r.Duration = time.Since(start).Round(time.Millisecond)
			// A terminal ends its lines with a carriage return too.
			r.Output = strings.ReplaceAll(out.String(), "\r\n", "\n")
			var exit *exec.ExitError
			switch {
			case run.Err() == context.DeadlineExceeded:
				r.ExitCode, r.Error = -1, "the command ran past "+timeout.String()
			case errors.As(err, &exit):
				r.ExitCode = exit.ExitCode()
				if r.ExitCode == 255 {
					r.Error = "ssh could not run the command"
				}
			case err != nil:
				r.ExitCode, r.Error = -1, err.Error()
			}
		}(&results[i])
	}
	wg.Wait()
	return results
}
