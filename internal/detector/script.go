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
		source.Targets = append(source.Targets, scriptTarget(dir, "scripts/"+entry.Name(), executable))
	}
	if len(source.Targets) == 0 {
		return nil
	}
	return source
}

// scriptTarget is the job of a script, rel to the folder that runs it:
// an executable runs itself, a .sh that is not runs with sh.
func scriptTarget(dir, rel string, executable bool) *Target {
	description, interactive := scriptHeader(filepath.Join(dir, rel))
	name := filepath.Base(rel)
	job := &Target{
		Name:        strings.TrimSuffix(name, filepath.Ext(name)),
		Description: description,
		Interactive: interactive,
		Args:        []string{"./" + rel},
		Steps:       []string{"./" + rel},
	}
	if !executable {
		job.Args = []string{"sh", rel}
		job.Steps = []string{"sh " + rel}
	}
	return job
}

// scriptHeader reads the comments at the top of a script, after the
// shebang: the first is its description, skipping editor and shellcheck
// directives, and a line `# interactive: true` makes it take keyboard
// input, as a pipeline job's interactive does.
func scriptHeader(path string) (string, bool) {
	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close()
	description, interactive := "", false
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
			break
		}
		switch {
		case text == "interactive: true":
			interactive = true
		case strings.HasPrefix(text, "shellcheck") || strings.HasPrefix(text, "vim:") || strings.HasPrefix(text, "-*-"):
		case description == "":
			description = text
		}
	}
	return description, interactive
}

func sortedStrings(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
