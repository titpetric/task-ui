package detector

import (
	"os"
	"path/filepath"
	"strings"
)

// Loose lists the scripts of a folder that is itself a scripts directory
// with a pipeline: the executables and shell scripts beside the pipeline
// file that the pipeline does not run, as a source aside it. A script
// the pipeline names, as a target or in a step, is the pipeline's. Nil
// for any other folder, or when there are none.
func Loose(dir string, pipeline *Source) *Source {
	if filepath.Base(dir) != "scripts" || pipeline == nil || pipeline.Kind == "script" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	source := &Source{Kind: "script", File: "./", Aside: true}
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
		job := scriptTarget(dir, entry.Name(), executable)
		if pipeline.Target(job.Name) != nil || runs(pipeline, entry.Name()) {
			continue
		}
		source.Targets = append(source.Targets, job)
	}
	if len(source.Targets) == 0 {
		return nil
	}
	return source
}

// runs reports whether a step of the pipeline names the file.
func runs(pipeline *Source, file string) bool {
	for _, target := range pipeline.Targets {
		for _, step := range target.Steps {
			if strings.Contains(step, file) {
				return true
			}
		}
	}
	return false
}
