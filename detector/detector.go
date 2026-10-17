// Package detector reads the conventions a folder follows: a pipeline file,
// a Makefile, package scripts, a scripts directory. Each convention is a
// Detector that returns a Source, the targets it declares and the command
// that runs them. Adding a convention is adding a Detector to Detectors;
// nothing else in task-ui asks which kind a target is.
package detector

import (
	"os"
	"path/filepath"
	"strings"
)

// Detector reads one convention out of a folder. It returns nil when the
// folder does not follow the convention, and a Source with Error set when
// it does but the file does not parse, so the folder still lists what is
// wrong.
type Detector interface {
	// Kind names the convention, as Source.Kind spells it.
	Kind() string
	// Detect reads the convention out of dir.
	Detect(dir string) *Source
}

// Detectors are the conventions task-ui knows, in the order their sources
// are listed: pipelines first, then build files, then scripts.
var Detectors = []Detector{
	pipelineDetector{kind: "atkins", names: atkinsNames},
	pipelineDetector{kind: "task", names: taskfileNames},
	makeDetector{},
	npmDetector{},
	composerDetector{},
	scriptDetector{},
}

// PipelineKinds are the conventions that describe a folder's jobs on
// purpose, as opposed to build files and scripts that happen to be there.
// A folder with one lists only its targets.
var PipelineKinds = []string{"atkins", "task"}

// Pipeline reports whether a kind is a pipeline convention.
func Pipeline(kind string) bool {
	return contains(PipelineKinds, kind)
}

// HasPipeline reports whether a folder has a pipeline file.
func HasPipeline(dir string) bool {
	for _, detector := range Detectors {
		if Pipeline(detector.Kind()) && detector.Detect(dir) != nil {
			return true
		}
	}
	return false
}

// ByKind returns the detector of a convention.
func ByKind(kind string) (Detector, bool) {
	for _, detector := range Detectors {
		if detector.Kind() == kind {
			return detector, true
		}
	}
	return nil, false
}

// Any reports whether a folder follows at least one convention.
func Any(dir string) bool {
	for _, detector := range Detectors {
		if detector.Detect(dir) != nil {
			return true
		}
	}
	return false
}

// All reads every convention out of a folder. With kinds given, only those
// detectors run, which is how a configuration narrows a folder.
func All(dir string, kinds ...string) []*Source {
	var sources []*Source
	for _, detector := range Detectors {
		if len(kinds) > 0 && !contains(kinds, detector.Kind()) {
			continue
		}
		if source := detector.Detect(dir); source != nil {
			sources = append(sources, source)
		}
	}
	return sources
}

// FromFile reads one file as a pipeline source: the file named in a
// configuration, or the one the single file mode serves. A YAML file with
// a name no convention claims is read as an atkins pipeline, which
// understands the Taskfile vocabulary as well. A file that is neither
// returns nil.
func FromFile(path string) *Source {
	dir, name := filepath.Dir(path), filepath.Base(path)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil
	}
	for _, detector := range Detectors {
		if p, ok := detector.(pipelineDetector); ok && contains(p.names, name) {
			return p.detectFile(dir, name)
		}
	}
	if ext := strings.ToLower(filepath.Ext(name)); ext == ".yml" || ext == ".yaml" {
		return pipelineDetector{kind: "atkins", names: []string{name}}.detectFile(dir, name)
	}
	return nil
}

func contains(list []string, item string) bool {
	for _, have := range list {
		if have == item {
			return true
		}
	}
	return false
}

// describeFromComment returns the text of a leading comment line, the way
// scripts and Makefiles document a target: `# text` or `## text`.
func describeFromComment(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return "", false
	}
	text := strings.TrimSpace(strings.TrimLeft(line, "#"))
	if text == "" || strings.HasPrefix(text, "!") {
		return "", false
	}
	return text, true
}
