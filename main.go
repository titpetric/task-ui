// Command task-ui serves a web console for the jobs of a directory tree:
// pipelines, Makefiles, package scripts and shell scripts, run in a
// terminal everyone can watch.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Build information, filled in by the linker.
var (
	version = "dev"
	commit  = ""
)

// hash prints a bcrypt hash for the password file. The password is read from
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

// passwd sets a user's password in the password file: added when new,
// replaced when there. The password is read from the arguments, or from
// stdin when there are none.
func passwd(path string, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 || args[0] == "" {
		return errors.New("usage: task-ui passwd <user> [password], or the password on stdin")
	}
	if path == "" {
		return errors.New("no password file: set TASK_UI_HTPASSWD, or put .htpasswd in the root")
	}
	var hashed bytes.Buffer
	if err := hash(args[1:], stdin, &hashed); err != nil {
		return err
	}
	if err := access.SetHtpasswd(path, args[0], strings.TrimSpace(hashed.String())); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "%s: password set in %s\n", args[0], path)
	return err
}

// openUser opens the database and finds a user by name, for the commands
// that change what the user signs in with. Users come from the password
// file, and reach the database when the server reads it.
func openUser(ctx context.Context, dsn, name string) (storage.DB, *model.User, error) {
	if name == "" {
		return nil, nil, errors.New("no user named")
	}
	if err := os.MkdirAll(filepath.Dir(dsn), 0o755); err != nil {
		return nil, nil, err
	}
	db, err := storage.OpenAndMigrate(ctx, dsn)
	if err != nil {
		return nil, nil, err
	}
	user, err := storage.NewUserStorage(db).GetByName(ctx, name)
	if err != nil {
		db.Close()
		if errors.Is(err, storage.ErrNotFound) {
			return nil, nil, fmt.Errorf("no user %s: add it with task-ui passwd, and let the server read the file", name)
		}
		return nil, nil, err
	}
	return db, user, nil
}

// email lists the email addresses a user signs in with, after adding or
// removing the ones given: task-ui email <user> [add|rm <address>...].
func email(ctx context.Context, dsn string, args []string, stdout io.Writer) error {
	if len(args) != 1 && (len(args) < 3 || args[1] != "add" && args[1] != "rm") {
		return errors.New("usage: task-ui email <user> [add|rm <address>...]")
	}
	db, user, err := openUser(ctx, dsn, args[0])
	if err != nil {
		return err
	}
	defer db.Close()
	emails := storage.NewUserEmailStorage(db)
	if len(args) > 1 {
		for _, address := range args[2:] {
			if args[1] == "rm" {
				err = emails.Delete(ctx, user.ID, address)
			} else {
				err = emails.Add(ctx, user.ID, address)
			}
			if err != nil {
				return fmt.Errorf("%s: %w", address, err)
			}
		}
	}
	list, err := emails.List(ctx, user.ID)
	if err != nil {
		return err
	}
	for _, e := range list {
		if _, err := fmt.Fprintln(stdout, e.Email); err != nil {
			return err
		}
	}
	return nil
}

// backup replaces a user's backup passwords with new ones, ten unless a
// count is given, and prints them. Each signs in once.
func backup(ctx context.Context, dsn string, args []string, stdout io.Writer) error {
	if len(args) == 0 || len(args) > 2 {
		return errors.New("usage: task-ui backup <user> [count]")
	}
	count := 10
	if len(args) == 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("count: %w", err)
		}
		count = n
	}
	passwords, hashes, err := access.BackupPasswords(count)
	if err != nil {
		return err
	}
	db, user, err := openUser(ctx, dsn, args[0])
	if err != nil {
		return err
	}
	defer db.Close()
	if err := storage.NewUserAuthPasswordStorage(db).SetBackups(ctx, user.ID, hashes); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s: backup passwords, each signs in once, the previous ones no longer do\n%s\n", user.Name, strings.Join(passwords, "\n"))
	return err
}

// token gives a user a token for a client, such as an MCP client, and
// prints it, the only time it is shown: task-ui token <user> <name>. A
// token of the same name stops working. task-ui token <user> lists the
// names, and task-ui token <user> rm <name> removes one.
func token(ctx context.Context, dsn string, args []string, stdout io.Writer) error {
	if len(args) < 1 || len(args) > 3 || len(args) == 3 && args[1] != "rm" || len(args) == 2 && args[1] == "rm" {
		return errors.New("usage: task-ui token <user> [<name> | rm <name>]")
	}
	db, user, err := openUser(ctx, dsn, args[0])
	if err != nil {
		return err
	}
	defer db.Close()
	tokens := storage.NewUserAuthTokenStorage(db)
	switch len(args) {
	case 3:
		return tokens.Delete(ctx, user.ID, args[2])
	case 2:
		secret, hash, err := access.NewToken()
		if err != nil {
			return err
		}
		if err := tokens.Set(ctx, user.ID, args[1], hash); err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, secret)
		return err
	}
	list, err := tokens.List(ctx, user.ID)
	if err != nil {
		return err
	}
	for _, t := range list {
		used := "never used"
		if t.UsedAt != nil {
			used = "used " + t.UsedAt.Local().Format("2006-01-02 15:04")
		}
		if _, err := fmt.Fprintf(stdout, "%s\t%s\n", t.Name, used); err != nil {
			return err
		}
	}
	return nil
}

// run is main without the exit: the hash, passwd, email, backup and
// token commands, the version, or the server until a signal.
func run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "hash" {
		return hash(args[1:], stdin, stdout)
	}
	cfg, err := app.NewConfig(args)
	if err != nil {
		return err
	}
	if len(args) > 0 && args[0] == "passwd" {
		return passwd(cfg.Htpasswd, args[1:], stdin, stdout)
	}
	if len(args) > 0 && args[0] == "email" {
		return email(ctx, cfg.DB, args[1:], stdout)
	}
	if len(args) > 0 && args[0] == "backup" {
		return backup(ctx, cfg.DB, args[1:], stdout)
	}
	if len(args) > 0 && args[0] == "token" {
		return token(ctx, cfg.DB, args[1:], stdout)
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
