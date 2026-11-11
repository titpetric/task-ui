package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// composerDetector reads the scripts of a composer.json.
type composerDetector struct{}

func (composerDetector) Kind() string { return "composer" }

func (d composerDetector) Detect(dir string) *Source {
	data, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		return nil
	}
	source := &Source{Kind: "composer", File: "composer.json", Command: []string{"composer", "run-script"}}
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
		source.Targets = append(source.Targets, &Target{Name: name, Description: desc, Steps: steps})
	}
	return source
}
