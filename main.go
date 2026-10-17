// Command task-ui serves the folders under a root and runs what they
// declare from the browser.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	_ "modernc.org/sqlite" // the sqlite driver the storage opens

	"github.com/titpetric/task-ui/access"
)

// hash prints a bcrypt hash for task-ui.yml. The password is read from
// the arguments, or from the first line of stdin.
func hash(args []string, stdin io.Reader, stdout io.Writer) error {
	password := strings.Join(args, " ")
	if password == "" {
		line, err := bufio.NewReader(stdin).ReadString('\n')
		if err != nil && line == "" {
			return errors.New("usage: task-ui hash <password>, or the password on stdin")
		}
		password = strings.TrimRight(line, "\r\n")
	}
	out, err := access.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, out)
	return err
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash" {
		if err := hash(os.Args[2:], os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "task-ui: the server is being rewired, see docs/plan.md")
	os.Exit(1)
}
