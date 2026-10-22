// Command rulelint matches the permission checks of task-ui against the
// embedded `config/permissions.yml` and prints each mismatch: a check of a
// rule the configuration does not hold, and a rule nothing checks. It exits 1
// when there is any.
//
//	go run ./internal/cmd/rulelint [dir]
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/rulelint"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run lints the directory args name, the current one by default, and
// returns the exit code: 1 with findings, 2 when the walk fails.
func run(args []string, stdout, stderr io.Writer) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}
	findings, err := rulelint.Lint(root, config.Default())
	if err != nil {
		fmt.Fprintln(stderr, "rulelint:", err)
		return 2
	}
	for _, finding := range findings {
		fmt.Fprintln(stdout, finding)
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}
