package folder

import (
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// Load reads a directory as a folder, with only the detectors of the kinds given when there are any.
func Load(id, dir string, kinds ...string) *Folder {
	f := &Folder{ID: id, Title: id, Dir: dir}
	f.Sources = detector.All(dir, kinds...)
	if len(kinds) == 0 {
		f.Sources = firstSource(f.Sources)
		if len(f.Sources) == 1 {
			if loose := detector.Loose(dir, f.Sources[0]); loose != nil {
				f.Sources = append(f.Sources, loose)
			}
		}
	}
	finish(f)
	return f
}

// firstSource keeps the one convention that reads a folder: the first
// there is, in the order of detector.Detectors. Atkins wins, then a
// Taskfile, a Makefile, package scripts, and the scripts folder only
// when nothing names the jobs. A project's sources name others on
// purpose.
func firstSource(sources []*detector.Source) []*detector.Source {
	if len(sources) > 1 {
		return sources[:1]
	}
	return sources
}

// LoadFile reads one file as the only source of a folder, for single file mode and a file the configuration names.
func LoadFile(id, path string) *Folder {
	f := &Folder{ID: id, Title: id, Dir: filepath.Dir(path)}
	if source := detector.FromFile(path); source != nil {
		f.Sources = []*detector.Source{source}
	} else {
		f.Error = "no convention reads " + filepath.Base(path)
	}
	finish(f)
	return f
}

// finish fills the folder from its sources: title and description from
// the first that has them, the job list with IDs that stay apart across
// sources, and the merged include and env file lists.
func finish(f *Folder) {
	var failures []string
	// A source aside keeps its targets' names; the others take their
	// kind in front when there is more than one.
	named := 0
	for _, source := range f.Sources {
		if !source.Aside {
			named++
		}
	}
	for _, source := range f.Sources {
		if f.Title == f.ID && source.Name != "" {
			f.Title = source.Name
		}
		if f.Description == "" && source.Description != "" {
			f.Description = source.Description
		}
		f.Includes = appendUnique(f.Includes, source.Includes...)
		f.EnvFiles = appendUnique(f.EnvFiles, source.EnvFiles...)
		if source.Error != "" {
			failures = append(failures, source.Error)
		}
		f.Errors = append(f.Errors, source.Errors...)
		for _, target := range source.Targets {
			job := &Job{
				Target: *target,
				Folder: f.ID,
				ID:     target.Name,
				Kind:   source.Kind,
				Source: source.File,
			}
			if named > 1 && !source.Aside {
				job.ID = source.Kind + ":" + target.Name
			}
			job.Command = append(append([]string{}, source.Command...), target.Argv()...)
			f.Jobs = append(f.Jobs, job)
		}
	}
	DefaultFirst(f.Jobs)
	if f.Error == "" && len(failures) > 0 {
		f.Error = strings.Join(failures, "\n")
	}
}

func appendUnique(list []string, items ...string) []string {
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		found := false
		for _, have := range list {
			if have == item {
				found = true
				break
			}
		}
		if !found {
			list = append(list, item)
		}
	}
	return list
}
