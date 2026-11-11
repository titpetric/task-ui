package api_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain runs as the executable bin/where of a checkout when started
// by that name: it prints where it runs and the argv[0] it has.
func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "where" {
		cwd, _ := os.Getwd()
		os.Stdout.WriteString("cwd=" + cwd + "\nargv0=" + os.Args[0] + "\n")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
