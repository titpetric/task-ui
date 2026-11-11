package detector

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// File returns the job of the file rel of a folder, named rel, or nil for a file that does not run.
// An executable runs itself, and a file whose first line is a shebang runs with what the shebang names.
// A header comment describes the job as it does a script's.
func File(dir, rel string) *Target {
	full := filepath.Join(dir, filepath.FromSlash(rel))
	info, err := os.Stat(full)
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	argv := []string{"./" + rel}
	if info.Mode()&0o111 == 0 {
		interpreter := Shebang(full)
		if interpreter == nil {
			return nil
		}
		argv = append(interpreter, rel)
	}
	description, interactive := scriptHeader(full)
	return &Target{
		Name:        path.Clean(rel),
		Description: description,
		Interactive: interactive,
		Args:        argv,
		Steps:       []string{strings.Join(argv, " ")},
	}
}

// Shebang returns what the first line of a file names to run it with,
// the interpreter and its arguments, or nil when the file has no
// shebang.
func Shebang(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()
	line, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && line == "" {
		return nil
	}
	rest, ok := strings.CutPrefix(line, "#!")
	if !ok {
		return nil
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return nil
	}
	return fields
}
