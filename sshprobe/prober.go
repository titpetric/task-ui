// Package sshprobe finds who task-ui can sign in to a host as over ssh:
// the first of a list of users whose login works with the instance keys,
// tried in batch mode, so a host that wants a password counts as closed.
package sshprobe

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Prober tries the users on a host and keeps what it found for TTL.
type Prober struct {
	// Users are tried in order; the first that signs in is the answer.
	Users []string
	// KeyDirs hold the private keys tried, the id_* files.
	KeyDirs []string
	// KnownHosts is the file host keys go to on first contact.
	KnownHosts string
	// Timeout bounds one attempt; TTL is how long an answer stays good.
	Timeout time.Duration
	TTL     time.Duration
	// Command is the ssh client, ssh unless set.
	Command string

	mu      sync.Mutex
	found   map[string]probe
	running map[string]bool
}

// probe is what a host answered, and when.
type probe struct {
	user string
	at   time.Time
}

// Logins returns who signs in to each host: user@host, or "" for a host
// none of the users reach. It waits at most wait for the hosts it has no
// fresh answer for; those still being tried come back empty this time and
// fill in for the next call.
func (p *Prober) Logins(ctx context.Context, hosts []string, wait time.Duration) map[string]string {
	var wg sync.WaitGroup
	for _, host := range hosts {
		if host == "" || !p.start(host) {
			continue
		}
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			user := p.try(context.WithoutCancel(ctx), host)
			p.mu.Lock()
			p.found[host] = probe{user: user, at: time.Now()}
			delete(p.running, host)
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
		if found, ok := p.found[host]; ok && found.user != "" {
			logins[host] = found.user + "@" + host
		}
	}
	return logins
}

// start marks a host to probe, unless its answer is fresh or a probe of
// it already runs.
func (p *Prober) start(host string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.found == nil {
		p.found, p.running = map[string]probe{}, map[string]bool{}
	}
	if found, ok := p.found[host]; ok && time.Since(found.at) < p.TTL {
		return false
	}
	if p.running[host] {
		return false
	}
	p.running[host] = true
	return true
}

// try returns the first user that signs in to the host, or "".
func (p *Prober) try(ctx context.Context, host string) string {
	keys := p.keys()
	for _, user := range p.Users {
		attempt, cancel := context.WithTimeout(ctx, p.Timeout)
		args := []string{"-n", "-o", "BatchMode=yes", "-o", "ConnectTimeout=" + seconds(p.Timeout), "-o", "StrictHostKeyChecking=accept-new"}
		if p.KnownHosts != "" {
			args = append(args, "-o", "UserKnownHostsFile="+p.KnownHosts)
		}
		for _, key := range keys {
			args = append(args, "-i", key)
		}
		args = append(args, user+"@"+host, "true")
		command := p.Command
		if command == "" {
			command = "ssh"
		}
		err := exec.CommandContext(attempt, command, args...).Run()
		cancel()
		if err == nil {
			return user
		}
	}
	return ""
}

// keys lists the private keys in the key directories.
func (p *Prober) keys() []string {
	var keys []string
	for _, dir := range p.KeyDirs {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && strings.HasPrefix(name, "id_") && !strings.HasSuffix(name, ".pub") {
				keys = append(keys, filepath.Join(dir, name))
			}
		}
	}
	return keys
}

// seconds is a timeout in whole seconds, at least one, for ssh.
func seconds(d time.Duration) string {
	return strconv.Itoa(max(1, int(d/time.Second)))
}
