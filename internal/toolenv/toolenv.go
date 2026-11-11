// Package toolenv detects the tools of the environment task-ui starts
// in: the runtimes, the tooling and the instance's own tools, each with
// its version, its path and what it is used for. The detection runs at
// start, every tool under its own timeout, so a missing or slow tool
// leaves no row and starts nothing.
package toolenv

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// The categories a tool belongs to.
const (
	// CategoryRuntime is a language runtime, such as node or go.
	CategoryRuntime = "runtime"
	// CategoryTooling is a general tool of the image, such as git.
	CategoryTooling = "tooling"
	// CategoryCustom is a tool of the instance's own, such as atkins.
	CategoryCustom = "custom"
)

// DefaultTimeout bounds one tool's version command.
const DefaultTimeout = 2 * time.Second

// Tool is a tool found in the PATH.
type Tool struct {
	// Name is how the pages and the prompt name the tool.
	Name string
	// Category is CategoryRuntime, CategoryTooling or CategoryCustom.
	Category string
	// Version is the tool's own version line, empty when it did not say.
	Version string
	// Path is where the PATH found the program.
	Path string
	// Description is one line on what the tool is used for.
	Description string
}

// Spec is a tool to look for.
type Spec struct {
	// ID is the spec's name in a URL, as "docker-compose".
	ID string
	// Name is the tool's name; Program the executable to find, Name
	// when empty, as "docker" finds "docker compose".
	Name        string
	Program     string
	Category    string
	Description string
	// Args ask for the version, such as "--version" or "version".
	Args []string
}

// Specs are the tools the instance looks for, in the order the pages
// list them.
var Specs = []Spec{
	{ID: "node", Name: "node", Category: CategoryRuntime, Description: "JavaScript runtime", Args: []string{"--version"}},
	{ID: "phpscript", Name: "phpscript", Category: CategoryRuntime, Description: "PHP interpreter in Go: runs, formats and tests PHP", Args: []string{"--version"}},
	{ID: "go", Name: "go", Category: CategoryRuntime, Description: "Go toolchain: builds and tests Go", Args: []string{"version"}},

	{ID: "docker", Name: "docker", Category: CategoryTooling, Description: "containers and images", Args: []string{"--version"}},
	{ID: "docker-compose", Name: "docker compose", Program: "docker", Category: CategoryTooling, Description: "multi-container applications", Args: []string{"compose", "version"}},
	{ID: "git", Name: "git", Category: CategoryTooling, Description: "version control", Args: []string{"--version"}},
	{ID: "make", Name: "make", Category: CategoryTooling, Description: "builds from makefiles", Args: []string{"--version"}},

	{ID: "worktree", Name: "worktree", Category: CategoryCustom, Description: "Go workspace tool: maps the modules of a tree", Args: []string{"--version"}},
	{ID: "atkins", Name: "atkins", Category: CategoryCustom, Description: "YAML task runner", Args: []string{"--version"}},
	{ID: "task", Name: "task", Category: CategoryCustom, Description: "Taskfile task runner", Args: []string{"--version"}},
	{ID: "semver", Name: "semver", Category: CategoryCustom, Description: "sorts a repository's tags as semantic versions", Args: []string{"--version"}},
	{ID: "secretsmanager", Name: "secretsmanager", Category: CategoryCustom, Description: "encrypted secrets in a repository", Args: []string{"--version"}},
	{ID: "terraform", Name: "terraform", Category: CategoryCustom, Description: "infrastructure as code", Args: []string{"version"}},
	{ID: "jq", Name: "jq", Category: CategoryCustom, Description: "JSON processor", Args: []string{"--version"}},
}

// SpecByID returns the spec of a tool, by the ID a URL names.
func SpecByID(id string) (Spec, bool) {
	for _, spec := range Specs {
		if spec.ID == id {
			return spec, true
		}
	}
	return Spec{}, false
}

// SpecByName returns the spec of a tool, by the name the pages show.
func SpecByName(name string) (Spec, bool) {
	for _, spec := range Specs {
		if spec.Name == name {
			return spec, true
		}
	}
	return Spec{}, false
}

// Detect probes a PATH for tools.
type Detect struct {
	// Path is the PATH to search, a list of directories as PATH holds
	// it; os.Getenv("PATH") when empty.
	Path string
	// Timeout bounds each version command; DefaultTimeout when zero.
	Timeout time.Duration
	// Specs are the tools to look for; Specs when empty.
	Specs []Spec
}

// Tools probes the PATH for every spec, at once, and returns the tools
// it found in the specs' order.
func (d Detect) Tools(ctx context.Context) []Tool {
	specs := d.Specs
	if len(specs) == 0 {
		specs = Specs
	}
	timeout := d.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	path := d.Path
	if path == "" {
		path = os.Getenv("PATH")
	}
	found := make([]*Tool, len(specs))
	// A handful of probes at a time: fifteen version commands started
	// together is a burst of processes at every start, and one slow tool
	// should not hold the rest past its timeout.
	workers := min(4, len(specs))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				spec := specs[i]
				program := spec.Program
				if program == "" {
					program = spec.Name
				}
				file := lookPath(path, program)
				if file == "" {
					continue
				}
				found[i] = &Tool{
					Name: spec.Name, Category: spec.Category, Path: file,
					Version: version(ctx, file, spec, timeout), Description: spec.Description,
				}
			}
		}()
	}
	for i := range specs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	tools := make([]Tool, 0, len(specs))
	for _, tool := range found {
		if tool != nil {
			tools = append(tools, *tool)
		}
	}
	return tools
}

// Env returns the tools as the value of TASK_UI_TOOLING, which a run's
// Claude context reads: one a line, the name, a colon and what the tool
// is used for; empty without tools.
func Env(tools []Tool) string {
	lines := make([]string, 0, len(tools))
	for _, tool := range tools {
		lines = append(lines, tool.Name+": "+tool.Description)
	}
	return strings.Join(lines, "\n")
}

// lookPath finds program in the directories of path.
func lookPath(path, program string) string {
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			continue
		}
		file := filepath.Join(dir, program)
		if info, err := os.Stat(file); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return file
		}
	}
	return ""
}

// version runs the spec's version command under the timeout and returns
// its first line, less the tool's name and the word version; empty when
// the command failed or said nothing.
func version(ctx context.Context, file string, spec Spec, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, file, spec.Args...)
	// A child of a killed command may hold its output open: stop
	// waiting for it after the timeout too.
	cmd.WaitDelay = timeout
	// Some tools say their version on stderr.
	out, _ := cmd.CombinedOutput()
	return versionLine(string(out), spec.Name)
}

// versionLine is the first line of a version output, less a leading
// name and the word version: "go version go1.22 linux/amd64" reads
// "go1.22 linux/amd64".
func versionLine(out, name string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		for _, prefix := range []string{strings.ToLower(name) + " version ", "version "} {
			if strings.HasPrefix(lower, prefix) {
				return strings.TrimSpace(line[len(prefix):])
			}
		}
		return line
	}
	return ""
}
