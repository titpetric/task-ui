// Package skills asks atkins what runs in a directory: its pipeline and
// the skills of the .atkins/skills folders above it whose condition holds
// there, each with the directory its jobs run in.
package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Prober runs `atkins -l --json --all` in a directory and keeps the
// sections for TTL, or until the pipeline file or a skills folder it read changes.
// A probe that fails or takes longer than Timeout answers nothing. Stamp
// tells a change apart: it reads the nearest pipeline file and the files of
// every .atkins/skills folder from the directory up.
type Prober struct {
	// Command is the atkins binary, atkins unless set.
	Command string
	// Timeout bounds one probe, 3s unless set; TTL keeps an answer, 30s
	// unless set.
	Timeout time.Duration
	TTL     time.Duration

	mu    sync.Mutex
	cache map[string]probe
}

// probe is what a directory answered, when, and the stamp of the files
// it read then.
type probe struct {
	sections []Section
	at       time.Time
	stamp    string
}

// Probe returns the sections atkins lists for dir, inactive skills included, nil on any failure.
func (p *Prober) Probe(dir string) []Section {
	return p.probe(dir, "")
}

// ProbeFile returns the sections atkins lists for one pipeline file in
// dir, for a configuration that names a file no convention claims.
func (p *Prober) ProbeFile(dir, file string) []Section {
	if file == "" {
		return nil
	}
	return p.probe(dir, file)
}

// probe answers from the cache when the files it read have not changed
// and the cached sections are younger than the TTL, and asks atkins
// otherwise.
func (p *Prober) probe(dir, file string) []Section {
	if p == nil || dir == "" {
		return nil
	}
	key := dir + "\x00" + file
	stamp := Stamp(dir)
	p.mu.Lock()
	cached, ok := p.cache[key]
	p.mu.Unlock()
	if ok && cached.stamp == stamp && time.Since(cached.at) < p.ttl() {
		return cached.sections
	}
	sections := p.run(dir, file)
	p.mu.Lock()
	if p.cache == nil {
		p.cache = map[string]probe{}
	}
	p.cache[key] = probe{sections: sections, at: time.Now(), stamp: stamp}
	p.mu.Unlock()
	return sections
}

// run execs atkins in dir. Its stdin is /dev/null: atkins reads a pipe
// on stdin as the pipeline.
func (p *Prober) run(dir, file string) []Section {
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
	cmd.Stdin = nil
	cmd.WaitDelay = time.Second
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}
	var sections []Section
	if err := json.Unmarshal(out.Bytes(), &sections); err != nil {
		return nil
	}
	return sections
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
