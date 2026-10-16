package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// npmDetector reads the scripts of a package.json. The package manager
// follows the lock file in the folder.
type npmDetector struct{}

func (npmDetector) Kind() string { return "npm" }

func (d npmDetector) Detect(dir string) *Source {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	source := &Source{Kind: "npm", File: "package.json", Command: []string{"npm", "run"}}
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
		source.Targets = append(source.Targets, &Target{
			Name:        name,
			Description: pkg.Scripts[name],
			Steps:       []string{pkg.Scripts[name]},
		})
	}
	return source
}
