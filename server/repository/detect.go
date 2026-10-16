package repository

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/server/model"
)

// Detector reads one convention out of a folder: a pipeline file, a
// Makefile, a scripts directory. It returns nil when the folder does not
// follow the convention, and a Source with Error set when it does but
// the file does not parse, so the folder still lists what is wrong.
//
// Adding a convention is adding a Detector to Detectors. The rest of
// task-ui runs on Sources and Jobs and never asks which kind they are.
type Detector interface {
	// Kind names the convention, as Source.Kind spells it.
	Kind() string
	// Detect reads the convention out of dir.
	Detect(dir string) *model.Source
}

// Detectors are the conventions task-ui knows, in the order their sources
// are listed: pipelines first, then build files, then scripts.
var Detectors = []Detector{
	pipelineDetector{kind: "atkins", names: AtkinsNames},
	pipelineDetector{kind: "task", names: TaskfileNames},
	makeDetector{},
	npmDetector{},
	composerDetector{},
	scriptDetector{},
}

// DetectorByKind returns the detector of a convention.
func DetectorByKind(kind string) (Detector, bool) {
	for _, detector := range Detectors {
		if detector.Kind() == kind {
			return detector, true
		}
	}
	return nil, false
}

// DetectAny reports whether a folder follows at least one convention.
func DetectAny(dir string) bool {
	for _, detector := range Detectors {
		if detector.Detect(dir) != nil {
			return true
		}
	}
	return false
}

// LoadFolder reads every convention out of a folder. With kinds given,
// only those detectors run, which is how task-ui.yml narrows a folder.
func LoadFolder(id, dir string, kinds ...string) *model.Project {
	project := &model.Project{ID: id, Title: id, Dir: dir}
	for _, detector := range Detectors {
		if len(kinds) > 0 && !contains(kinds, detector.Kind()) {
			continue
		}
		if source := detector.Detect(dir); source != nil {
			project.Sources = append(project.Sources, source)
		}
	}
	finish(project)
	return project
}

// LoadProject reads one file as the only source of a folder, for the
// single file mode and for a file a configuration names.
func LoadProject(id, path string) *model.Project {
	dir := filepath.Dir(path)
	project := &model.Project{ID: id, Title: id, Dir: dir}
	if detector, ok := detectorForFile(filepath.Base(path)); ok {
		if source := detector.detectFile(dir, filepath.Base(path)); source != nil {
			project.Sources = append(project.Sources, source)
		}
	} else {
		project.Error = "no convention reads " + filepath.Base(path)
	}
	finish(project)
	return project
}

// detectorForFile picks the pipeline detector for a file name. A YAML
// file with a name no convention claims is read as an atkins pipeline,
// which understands the Taskfile vocabulary as well.
func detectorForFile(name string) (pipelineDetector, bool) {
	for _, detector := range Detectors {
		if p, ok := detector.(pipelineDetector); ok && contains(p.names, name) {
			return p, true
		}
	}
	if ext := strings.ToLower(filepath.Ext(name)); ext == ".yml" || ext == ".yaml" {
		return pipelineDetector{kind: "atkins", names: []string{name}}, true
	}
	return pipelineDetector{}, false
}

// finish fills the folder from its sources: title and description from
// the first that has them, the job list with IDs that stay apart across
// sources, and the merged include and env file lists.
func finish(project *model.Project) {
	var failures []string
	for _, source := range project.Sources {
		if project.Title == project.ID && source.Name != "" {
			project.Title = source.Name
		}
		if project.Description == "" && source.Description != "" {
			project.Description = source.Description
		}
		project.Includes = appendUnique(project.Includes, source.Includes...)
		project.EnvFiles = appendUnique(project.EnvFiles, source.EnvFiles...)
		if source.Error != "" {
			failures = append(failures, source.Error)
		}
		for _, job := range source.Jobs {
			job.Project = project.ID
			job.Kind = source.Kind
			job.Source = source.File
			job.ID = job.Name
			if len(project.Sources) > 1 {
				job.ID = source.Kind + ":" + job.Name
			}
			if len(job.Args) == 0 {
				job.Args = []string{job.Name}
			}
			job.Command = append(append([]string{}, source.Command...), job.Args...)
			project.Jobs = append(project.Jobs, job)
		}
	}
	if project.Error == "" && len(failures) > 0 {
		project.Error = errors.Join(toErrors(failures)...).Error()
	}
}

func toErrors(messages []string) []error {
	result := make([]error, 0, len(messages))
	for _, m := range messages {
		result = append(result, errors.New(m))
	}
	return result
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
