package main

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"

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
// the argument, or from the first line of stdin.
func hash(args []string) error {
	password := strings.Join(args, " ")
	if password == "" {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return fmt.Errorf("usage: task-ui hash <password>, or the password on stdin")
		}
		password = strings.TrimRight(line, "\r\n")
	}
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash" {
		if err := hash(os.Args[2:]); err != nil {
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
