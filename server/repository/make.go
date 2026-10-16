package repository

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/titpetric/task-ui/server/model"
)

// MakefileNames are the file names make reads, in its order.
var MakefileNames = []string{"GNUmakefile", "makefile", "Makefile"}

// makeDetector reads the targets of a Makefile. A target is documented
// with `## text` after its rule, the convention `make help` recipes read,
// or with a comment line right above it. Pattern rules, dot targets and
// variable assignments are not targets.
type makeDetector struct{}

func (makeDetector) Kind() string { return "make" }

var makeRule = regexp.MustCompile(`^([A-Za-z0-9_][A-Za-z0-9_./ -]*?)\s*::?(?:\s|$)(.*)$`)

func (d makeDetector) Detect(dir string) *model.Source {
	for _, name := range MakefileNames {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return d.read(dir, name)
		}
	}
	return nil
}

func (d makeDetector) read(dir, name string) *model.Source {
	source := &model.Source{Kind: "make", File: name, Command: []string{"make"}}
	if name != "Makefile" {
		source.Command = []string{"make", "-f", name}
	}

	file, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		source.Error = err.Error()
		return source
	}
	defer file.Close()

	var (
		comment string
		current *model.Job
		seen    = map[string]bool{}
	)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "\t"):
			// A recipe line belongs to the target above it.
			if current != nil {
				if recipe := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "@")); recipe != "" {
					current.Steps = append(current.Steps, recipe)
				}
			}
			continue
		case strings.TrimSpace(line) == "":
			comment = ""
			current = nil
			continue
		}
		if text, ok := describeFromComment(line); ok {
			comment = text
			continue
		}
		current = nil
		if strings.ContainsAny(line, "=$%") && !strings.Contains(line, "##") {
			// An assignment, a pattern rule or a computed name.
			if i := strings.IndexAny(line, "=$%"); i >= 0 && !strings.Contains(line[:i], ":") {
				comment = ""
				continue
			}
		}
		match := makeRule.FindStringSubmatch(line)
		if match == nil {
			comment = ""
			continue
		}
		names, rest := match[1], match[2]
		if strings.ContainsAny(names, "=$%") {
			comment = ""
			continue
		}
		desc := comment
		if i := strings.Index(rest, "##"); i >= 0 {
			desc = strings.TrimSpace(rest[i+2:])
			rest = rest[:i]
		}
		deps := strings.Fields(strings.TrimSpace(rest))
		comment = ""
		for _, target := range strings.Fields(names) {
			if strings.HasPrefix(target, ".") || seen[target] {
				continue
			}
			seen[target] = true
			job := &model.Job{Name: target, Description: desc}
			for _, dep := range deps {
				if !strings.ContainsAny(dep, "$%|") {
					job.DependsOn = append(job.DependsOn, dep)
				}
			}
			source.Jobs = append(source.Jobs, job)
			current = job
		}
	}
	if err := scanner.Err(); err != nil {
		source.Error = err.Error()
	}
	return source
}
