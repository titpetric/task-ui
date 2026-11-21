// Package sshprobe finds who task-ui can sign in to a host as over ssh:
// the first of a list of users whose login works with the instance keys,
// tried in batch mode, so a host that only takes a password counts as closed.
package sshprobe

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Prober tries the users on a host and keeps what it found for TTL.
// A host whose probe still runs when Logins stops waiting comes back
// empty, and the probe's answer fills in for the next call.
type Prober struct {
	// Users are the logins Logins tries, in order; Logins keeps the first
	// that signs in.
	Users []string
	// Keys serve the instance's SSH keys to a probe in an ssh agent; nil
	// probes with none. A run of a host command gets them in the run's
	// own agent.
	Keys Keys
	// KnownHosts is the file host keys go to on first contact.
	KnownHosts string
	// Timeout bounds one attempt; TTL is how long an answer stays good.
	Timeout time.Duration
	TTL     time.Duration
	// Command is the ssh client, ssh unless set.
	Command string
	// Terminal runs a command of Run on a terminal (ssh -tt), so programs
	// print their colours, with PAGER=cat so none waits on a pager.
	Terminal bool

	mu      sync.Mutex
	found   map[string]probe
	running map[string]bool
}

// probe is what a host answered, and when.
type probe struct {
	user string
	at   time.Time
}

// Logins returns user@host for each host, "" for one no user reaches, waiting at most wait for fresh answers.
func (p *Prober) Logins(ctx context.Context, hosts []string, wait time.Duration) map[string]string {
	return p.LoginsAs(ctx, hosts, p.Users, wait)
}

// LoginsAs is Logins, trying users first; a page adds a host as user@host, and the user given is tried before the instance's own.
func (p *Prober) LoginsAs(ctx context.Context, hosts, users []string, wait time.Duration) map[string]string {
	key := func(host string) string { return strings.Join(users, ",") + "\x1f" + host }
	order := strings.Join(users, ",")
	var wg sync.WaitGroup
	for _, host := range hosts {
		if host == "" || !p.start(key(host)) {
			continue
		}
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			user := p.try(context.WithoutCancel(ctx), host, users)
			p.mu.Lock()
			p.found[key(host)] = probe{user: user, at: time.Now()}
			delete(p.running, key(host))
			p.mu.Unlock()
		}(host)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(wait):
	case <-ctx.Done():
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	logins := map[string]string{}
	for _, host := range hosts {
		if found, ok := p.found[order+"\x1f"+host]; ok && found.user != "" {
			logins[host] = found.user + "@" + host
		}
	}
	return logins
}

// start marks a host to probe, unless fresh or already running.
func (p *Prober) start(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.found == nil {
		p.found, p.running = map[string]probe{}, map[string]bool{}
	}
	if found, ok := p.found[key]; ok && time.Since(found.at) < p.TTL {
		return false
	}
	if p.running[key] {
		return false
	}
	p.running[key] = true
	return true
}

// try returns the first user that signs in to the host, or "".
func (p *Prober) try(ctx context.Context, host string, users []string) string {
	socket := ""
	if p.Keys != nil {
		agent, err := p.Keys.LoginAgent(ctx)
		if err != nil {
			return ""
		}
		defer agent.Close()
		if agent != nil {
			socket = agent.Socket
		}
	}
	for _, user := range users {
		attempt, cancel := context.WithTimeout(ctx, p.Timeout)
		cmd := p.ssh(attempt, user+"@"+host, "true")
		if socket != "" {
			cmd.Args = append([]string{cmd.Args[0], "-o", "IdentityAgent=" + socket}, cmd.Args[1:]...)
		}
		err := cmd.Run()
		cancel()
		if err == nil {
			return user
		}
	}
	return ""
}

// ssh returns the batch-mode ssh command for a login and its remote command.
func (p *Prober) ssh(ctx context.Context, login, remote string) *exec.Cmd {
	args := []string{"-n", "-o", "BatchMode=yes", "-o", "ConnectTimeout=" + seconds(p.Timeout), "-o", "StrictHostKeyChecking=accept-new"}
	if p.KnownHosts != "" {
		args = append(args, "-o", "UserKnownHostsFile="+p.KnownHosts)
	}
	args = append(args, login, remote)
	command := p.Command
	if command == "" {
		command = "ssh"
	}
	return exec.CommandContext(ctx, command, args...)
}

// terminal returns the ssh command of Run, opening a terminal when Terminal says so.
func (p *Prober) terminal(ctx context.Context, login, remote string) *exec.Cmd {
	if !p.Terminal {
		return p.ssh(ctx, login, remote)
	}
	cmd := p.ssh(ctx, login, "export PAGER=cat SYSTEMD_PAGER=cat GIT_PAGER=cat; "+remote)
	// -tt opens a terminal with stdin closed; -q drops ssh's own close notice.
	cmd.Args = append([]string{cmd.Args[0], "-tt", "-q"}, cmd.Args[1:]...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	return cmd
}

// seconds is a timeout in whole seconds, at least one, for ssh.
func seconds(d time.Duration) string {
	return strconv.Itoa(max(1, int(d/time.Second)))
}
