package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// scriptDetector lists the executables and shell scripts of a scripts
// directory. A script describes itself with the first comment after the
// shebang. A scripts directory with a pipeline file or a Makefile of its
// own runs its scripts itself, as a folder of its own, so the folder
// above does not list them again.
type scriptDetector struct{}

// ownsScripts reports whether a scripts directory follows a convention
// other than a scripts folder: a pipeline file, a Makefile, a
// package.json.
func ownsScripts(dir string) bool {
	for _, detector := range Detectors {
		if detector.Kind() != "script" && detector.Detect(dir) != nil {
			return true
		}
	}
	return false
}

func (scriptDetector) Kind() string { return "script" }

func (d scriptDetector) Detect(dir string) *Source {
	entries, err := os.ReadDir(filepath.Join(dir, "scripts"))
	if err != nil {
		return nil
	}
	if ownsScripts(filepath.Join(dir, "scripts")) {
		return nil
	}
	source := &Source{Kind: "script", File: "scripts/", Command: nil}
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
		job := &Target{
			Name:        strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Description: scriptDescription(filepath.Join(dir, rel)),
			Args:        []string{"./" + rel},
			Steps:       []string{"./" + rel},
		}
		if !executable {
			job.Args = []string{"sh", rel}
			job.Steps = []string{"sh " + rel}
		}
		source.Targets = append(source.Targets, job)
	}
	if len(source.Targets) == 0 {
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
