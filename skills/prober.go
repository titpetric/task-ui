// Package skills asks atkins what runs in a directory: its pipeline and
// the skills of the .atkins/skills folders above it whose condition holds
// there, each with the directory its jobs run in.
package skills

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/titpetric/task-ui/layout"
)

// Prober runs `atkins -l --json --all` in a directory and keeps the
// listing for TTL, or until the pipeline file or a skills folder it read changes.
// A probe that fails or takes longer than Timeout answers nothing; one
// atkins' linter refuses answers its findings. Stamp tells a change
// apart: it reads the nearest pipeline file and the files of every
// .atkins/skills folder from the directory up.
type Prober struct {
	// Command is the atkins binary, atkins unless set.
	Command string
	// Timeout bounds one probe, 3s unless set; TTL keeps an answer, 30s
	// unless set.
	Timeout time.Duration
	TTL     time.Duration
	// Home returns the HOME a probe of dir runs atkins with, the home of
	// dir's workspace, for its ~/.atkins/skills; "" keeps the server's.
	// Nil is "" everywhere.
	Home func(dir string) string

	mu    sync.Mutex
	cache map[string]probe
}

// probe is what a directory answered, when, and the stamp of the files
// it read then.
type probe struct {
	listing Listing
	at      time.Time
	stamp   string
}

// Probe returns the sections atkins lists for dir, inactive skills included, nil on any failure.
func (p *Prober) Probe(dir string) []Section {
	return p.probe(dir, "").Sections
}

// ProbeFile returns the sections atkins lists for one pipeline file in
// dir, for a configuration that names a file no convention claims.
func (p *Prober) ProbeFile(dir, file string) []Section {
	if file == "" {
		return nil
	}
	return p.probe(dir, file).Sections
}

// List returns what atkins answers for dir, or for one file in it.
func (p *Prober) List(dir, file string) Listing {
	return p.probe(dir, file)
}

// probe answers from the cache when the files it read have not changed
// and the cached listing is younger than the TTL, and asks atkins
// otherwise.
func (p *Prober) probe(dir, file string) Listing {
	if p == nil || dir == "" {
		return Listing{}
	}
	key := dir + "\x00" + file
	home := p.home(dir)
	stamp := Stamp(dir)
	if home != "" {
		stamp += ";home=" + Stamp(home)
	}
	p.mu.Lock()
	cached, ok := p.cache[key]
	p.mu.Unlock()
	if ok && cached.stamp == stamp && time.Since(cached.at) < p.ttl() {
		return cached.listing
	}
	listing := p.run(dir, file, home)
	p.mu.Lock()
	if p.cache == nil {
		p.cache = map[string]probe{}
	}
	p.cache[key] = probe{listing: listing, at: time.Now(), stamp: stamp}
	p.mu.Unlock()
	return listing
}

// home returns the HOME of a probe of dir, "" for the server's.
func (p *Prober) home(dir string) string {
	if p.Home == nil {
		return ""
	}
	return p.Home(dir)
}

// run execs atkins in dir, with HOME at home when set and ATKINS_DB as a
// run there gets it, the home's history and the instance's mirror. Its
// stdin is /dev/null: atkins reads a pipe on stdin as the pipeline. A
// failing atkins answers only when it prints lint findings, which it
// does with a non-zero exit.
func (p *Prober) run(dir, file, home string) Listing {
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	args := []string{"-l", "--json", "--all"}
	if file != "" {
		args = append([]string{"-f", file}, args...)
	}
	command := p.Bin()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	if home != "" {
		cmd.Env = append(os.Environ(), "HOME="+home, "ATKINS_DB="+layout.AtkinsDB(home, os.Getenv("HOME")))
	}
	cmd.Stdin = nil
	cmd.WaitDelay = time.Second
	var out bytes.Buffer
	cmd.Stdout = &out
	failed := cmd.Run() != nil
	listing, err := ParseListing(out.Bytes())
	if err != nil || failed && !listing.Broken() {
		return Listing{}
	}
	return listing
}

// Bin is the atkins the probe runs, and the one a skill's job runs with:
// Command, or atkins from the PATH.
func (p *Prober) Bin() string {
	if p == nil || p.Command == "" {
		return "atkins"
	}
	return p.Command
}

func (p *Prober) ttl() time.Duration {
	if p.TTL > 0 {
		return p.TTL
	}
	return 30 * time.Second
}

// pipelineNames are the files atkins reads as a directory's pipeline.
var pipelineNames = []string{".atkins.yml", ".atkins.yaml", "atkins.yml", "atkins.yaml"}

// Stamp sums up the modification times of the files a probe of dir reads, so a change to any answers afresh.
func Stamp(dir string) string {
	var parts []string
	pipeline := false
	for {
		if !pipeline {
			for _, name := range pipelineNames {
				if info, err := os.Stat(filepath.Join(dir, name)); err == nil {
					parts = append(parts, name+"@"+dir+"="+strconv.FormatInt(info.ModTime().UnixNano(), 10))
					pipeline = true
					break
				}
			}
		}
		skills := filepath.Join(dir, ".atkins", "skills")
		if entries, err := os.ReadDir(skills); err == nil {
			parts = append(parts, "skills@"+dir)
			for _, entry := range entries {
				if info, err := entry.Info(); err == nil {
					parts = append(parts, entry.Name()+"="+strconv.FormatInt(info.ModTime().UnixNano(), 10))
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return strings.Join(parts, ";")
}
