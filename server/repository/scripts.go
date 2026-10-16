package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/server/model"
)

// npmDetector reads the scripts of a package.json. The package manager
// follows the lock file in the folder.
type npmDetector struct{}

func (npmDetector) Kind() string { return "npm" }

func (d npmDetector) Detect(dir string) *model.Source {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	source := &model.Source{Kind: "npm", File: "package.json", Command: []string{"npm", "run"}}
	for lock, manager := range map[string][]string{
		"pnpm-lock.yaml": {"pnpm", "run"},
		"yarn.lock":      {"yarn", "run"},
		"bun.lockb":      {"bun", "run"},
		"bun.lock":       {"bun", "run"},
	} {
		if _, err := os.Stat(filepath.Join(dir, lock)); err == nil {
			source.Command = manager
			break
		}
	}

	var pkg struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Scripts     map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		source.Error = "package.json: " + err.Error()
		return source
	}
	source.Name = pkg.Name
	source.Description = pkg.Description
	for _, name := range sortedStrings(pkg.Scripts) {
		if strings.HasPrefix(name, "pre") || strings.HasPrefix(name, "post") {
			// Hooks run around the script they are named after.
			base := strings.TrimPrefix(strings.TrimPrefix(name, "pre"), "post")
			if _, ok := pkg.Scripts[base]; ok {
				continue
			}
		}
		source.Jobs = append(source.Jobs, &model.Job{
			Name:        name,
			Description: pkg.Scripts[name],
			Steps:       []string{pkg.Scripts[name]},
		})
	}
	return source
}

// composerDetector reads the scripts of a composer.json.
type composerDetector struct{}

func (composerDetector) Kind() string { return "composer" }

func (d composerDetector) Detect(dir string) *model.Source {
	data, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		return nil
	}
	source := &model.Source{Kind: "composer", File: "composer.json", Command: []string{"composer", "run-script"}}
	var pkg struct {
		Name                string                     `json:"name"`
		Description         string                     `json:"description"`
		Scripts             map[string]json.RawMessage `json:"scripts"`
		ScriptsDescriptions map[string]string          `json:"scripts-descriptions"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		source.Error = "composer.json: " + err.Error()
		return source
	}
	source.Name = pkg.Name
	source.Description = pkg.Description
	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		// Composer runs these itself around install and update.
		if strings.HasPrefix(name, "pre-") || strings.HasPrefix(name, "post-") {
			continue
		}
		var steps []string
		var one string
		if json.Unmarshal(pkg.Scripts[name], &one) == nil {
			steps = []string{one}
		} else {
			_ = json.Unmarshal(pkg.Scripts[name], &steps)
		}
		desc := pkg.ScriptsDescriptions[name]
		if desc == "" && len(steps) > 0 {
			desc = steps[0]
		}
		source.Jobs = append(source.Jobs, &model.Job{Name: name, Description: desc, Steps: steps})
	}
	return source
}

// scriptDetector lists the executables and shell scripts of a scripts
// directory. A script describes itself with the first comment after the
// shebang.
type scriptDetector struct{}

func (scriptDetector) Kind() string { return "script" }

func (d scriptDetector) Detect(dir string) *model.Source {
	entries, err := os.ReadDir(filepath.Join(dir, "scripts"))
	if err != nil {
		return nil
	}
	source := &model.Source{Kind: "script", File: "scripts/", Command: nil}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		executable := info.Mode()&0o111 != 0
		if !executable && !strings.HasSuffix(entry.Name(), ".sh") {
			continue
		}
		rel := "scripts/" + entry.Name()
		job := &model.Job{
			Name:        strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Description: scriptDescription(filepath.Join(dir, rel)),
			Args:        []string{"./" + rel},
			Steps:       []string{"./" + rel},
		}
		if !executable {
			job.Args = []string{"sh", rel}
			job.Steps = []string{"sh " + rel}
		}
		source.Jobs = append(source.Jobs, job)
	}
	if len(source.Jobs) == 0 {
		return nil
	}
	return source
}

// scriptDescription reads the first comment after the shebang, skipping
// editor and shellcheck directives.
func scriptDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan() && i < 10; i++ {
		line := strings.TrimSpace(scanner.Text())
		if i == 0 && strings.HasPrefix(line, "#!") {
			continue
		}
		if line == "" {
			continue
		}
		text, ok := describeFromComment(line)
		if !ok {
			return ""
		}
		if strings.HasPrefix(text, "shellcheck") || strings.HasPrefix(text, "vim:") || strings.HasPrefix(text, "-*-") {
			continue
		}
		return text
	}
	return ""
}

func sortedStrings(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
