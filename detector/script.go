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
// shebang.
type scriptDetector struct{}

func (scriptDetector) Kind() string { return "script" }

func (d scriptDetector) Detect(dir string) *Source {
	entries, err := os.ReadDir(filepath.Join(dir, "scripts"))
	if err != nil {
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
