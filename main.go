// Command task-ui serves the folders under a root and runs what they
// declare from the browser.
package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/server"
)

//go:embed templates/*.tpl public_html/static
var files embed.FS

// assets returns the embedded templates and static files, or the folder
// named in TASK_UI_ASSETS, which serves them from disk while developing.
func assets() fs.FS {
	if dir := os.Getenv("TASK_UI_ASSETS"); dir != "" {
		return os.DirFS(dir)
	}
	return files
}

func start(ctx context.Context) error {
	svc, err := server.New(assets(), os.Args[1:])
	if err != nil {
		return err
	}
	return svc.Start(ctx)
}

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := start(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "task-ui:", err)
		os.Exit(1)
	}
}
