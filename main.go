// Command task-ui serves a web console for the jobs of a directory tree:
// pipelines, Makefiles, package scripts and shell scripts, run in a
// terminal everyone can watch.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
)

// Build information, filled in by the linker.
var (
	version = "dev"
	commit  = ""
)

// hash prints a bcrypt hash for task-ui.yml. The password is read from
// the arguments, or from stdin when there are none.
func hash(args []string, stdin io.Reader, stdout io.Writer) error {
	password := strings.Join(args, " ")
	if password == "" {
		line, err := bufio.NewReader(stdin).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		password = strings.TrimRight(line, "\r\n")
	}
	if password == "" {
		return errors.New("usage: task-ui hash <password>, or the password on stdin")
	}
	hashed, err := access.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, hashed)
	return err
}

// run is main without the exit: the hash command, the version, or the
// server until a signal.
func run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "hash" {
		return hash(args[1:], stdin, stdout)
	}
	cfg, err := app.NewConfig(args)
	if err != nil {
		return err
	}
	if cfg.PrintBuildInfo {
		_, err := fmt.Fprintf(stdout, "task-ui %s %s\n", version, commit)
		return err
	}
	a, err := app.New(ctx, cfg, version)
	if err != nil {
		return err
	}
	defer a.Close()
	return a.Run(ctx)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "task-ui:", err)
		os.Exit(1)
	}
}
