package sshprobe

import (
	"fmt"
	"regexp"
	"strings"
)

// escapes matches the escape sequences a terminal prints: colours, cursor
// moves, titles.
var escapes = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[()][A-Za-z0-9]|\x1b.`)

// Report writes the results of a command as one markdown document: the
// command, then a section a host with its login, exit code, any error, and
// its output as plain text.
func Report(command string, results []Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# `%s` on %d hosts\n", strings.ReplaceAll(command, "`", "'"), len(results))
	for _, r := range results {
		login := r.Login
		if login == "" {
			login = r.Host
		}
		fmt.Fprintf(&b, "\n## %s\n\nexit %d, %s\n", login, r.ExitCode, r.Duration)
		if r.Error != "" {
			fmt.Fprintf(&b, "\nerror: %s\n", r.Error)
		}
		if output := strings.TrimRight(escapes.ReplaceAllString(r.Output, ""), "\n"); output != "" {
			fmt.Fprintf(&b, "\n```\n%s\n```\n", strings.ReplaceAll(output, "```", "'''"))
		}
	}
	return b.String()
}
