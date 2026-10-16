package folder

import (
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// Load reads every convention out of a directory. With kinds given, only
// those detectors run, which is how the configuration narrows a folder.
func Load(id, dir string, kinds ...string) *Folder {
	f := &Folder{ID: id, Title: id, Dir: dir}
	f.Sources = detector.All(dir, kinds...)
	finish(f)
	return f
}

// LoadFile reads one file as the only source of a folder, for the single
// file mode and for a file the configuration names.
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
		for _, target := range source.Targets {
			job := &Job{
				Target: *target,
				Folder: f.ID,
				ID:     target.Name,
				Kind:   source.Kind,
				Source: source.File,
			}
			if len(f.Sources) > 1 {
				job.ID = source.Kind + ":" + target.Name
			}
			job.Command = append(append([]string{}, source.Command...), target.Argv()...)
			f.Jobs = append(f.Jobs, job)
		}
	}
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
