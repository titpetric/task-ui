package folder

import (
	"path/filepath"
	"testing"
)

// The lines of a file the folder's pipeline reads, for its Show button:
// a last line without a newline counts; a file it cannot read is 0.
func TestLines(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "atkins.yml"), "jobs:\n  a: echo a\n")
	write(t, filepath.Join(dir, "vars.yml"), "x: 1\ny: 2")
	write(t, filepath.Join(dir, "empty.yml"), "")
	f := &Folder{ID: "f", Dir: dir}

	for name, want := range map[string]int{"atkins.yml": 2, "vars.yml": 2, "empty.yml": 0, "missing.yml": 0, "../escape.yml": 0} {
		if got := f.Lines(name); got != want {
			t.Errorf("%s: %d lines, want %d", name, got, want)
		}
	}
}
