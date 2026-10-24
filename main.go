// Command task-ui serves a web console for the jobs of a directory tree:
// pipelines, Makefiles, package scripts and shell scripts, run in a
// terminal everyone can watch.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/channel"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// Build information, filled in by the linker.
var (
	version = "dev"
	commit  = ""
)

// hash prints a bcrypt hash for the password file. The password comes from
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
// replaced when there. The password comes from the arguments, or from
// stdin when there are none.
func passwd(path string, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 || args[0] == "" {
		return errors.New("usage: task-ui passwd <user> [password], or the password on stdin")
	}
	if path == "" {
		return errors.New("no password file: it is .htpasswd in the data directory, TASK_UI_DATA")
	}
	var hashed bytes.Buffer
	if err := hash(args[1:], stdin, &hashed); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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

// backup replaces a user's backup passwords with new ones, ten unless the
// arguments name a count, and prints them. Each signs in once.
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

// token makes a token for a user's client, such as an MCP client, and
// prints it, the only time anyone sees it: task-ui token <user> <name>. A
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

// userCommand runs task-ui user, over the database without the server:
// list prints every user with the admin flag, role and MCP flags; admin,
// role and mcp set one of them. Taking the admin flag from the last
// admin needs --force.
func userCommand(ctx context.Context, dsn string, args []string, stdout io.Writer) error {
	const usage = "usage: task-ui user list | admin <user> on|off [--force] | role <user> runner|viewer | mcp <user> on|off | agent <user> on|off"
	force := false
	var rest []string
	for _, arg := range args {
		if arg == "--force" {
			force = true
			continue
		}
		rest = append(rest, arg)
	}
	if len(rest) == 1 && rest[0] == "list" {
		return listUsers(ctx, dsn, stdout)
	}
	if len(rest) != 3 {
		return errors.New(usage)
	}
	db, user, err := openUser(ctx, dsn, rest[1])
	if err != nil {
		return err
	}
	defer db.Close()
	admin, role, mcp, reply := user.Admin, user.Role, user.Mcp, user.AgentReply
	on, value := rest[2] == "on", rest[2]
	if rest[0] != "role" && value != "on" && value != "off" {
		return errors.New(usage)
	}
	switch rest[0] {
	case "admin":
		admin = on
	case "role":
		role = value
	case "mcp":
		mcp = on
		reply = reply && on
	case "agent":
		reply = on
		mcp = mcp || on
	default:
		return errors.New(usage)
	}
	if err := storage.NewUserStorage(db).SetFlags(ctx, user, admin, role, mcp, reply, force); err != nil {
		if errors.Is(err, storage.ErrLastAdmin) {
			return errors.New("the last admin keeps the admin flag: make another user an admin first, or add --force")
		}
		return err
	}
	return printUser(stdout, user)
}

// listUsers prints every user, a line each.
func listUsers(ctx context.Context, dsn string, stdout io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(dsn), 0o755); err != nil {
		return err
	}
	db, err := storage.OpenAndMigrate(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	users, err := storage.NewUserStorage(db).List(ctx)
	if err != nil {
		return err
	}
	for i := range users {
		if err := printUser(stdout, &users[i]); err != nil {
			return err
		}
	}
	return nil
}

// printUser prints a user's name, role and flags on a line.
func printUser(stdout io.Writer, user *model.User) error {
	flags := []string{}
	if user.Admin {
		flags = append(flags, "admin")
	}
	if user.Mcp {
		flags = append(flags, "mcp")
	}
	if user.AgentReply {
		flags = append(flags, "agent")
	}
	_, err := fmt.Fprintf(stdout, "%s\t%s\t%s\n", user.Name, user.Role, strings.Join(flags, ","))
	return err
}

// channelBridge runs task-ui channel: a stdio MCP server that pushes the
// notes naming the token's user into the Claude Code session that
// started it. Stdout is the session's; what the bridge says goes to
// stderr.
func channelBridge(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	flags := flag.NewFlagSet("channel", flag.ContinueOnError)
	base := flags.String("url", os.Getenv("TASK_UI_URL"), "the instance's base link, such as https://task.example.com")
	tokenFile := flags.String("token-file", "", "a file that holds the token of the user whose notes to push")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *tokenFile == "" {
		return errors.New("usage: task-ui channel --url <instance> --token-file <file>")
	}
	token, err := os.ReadFile(*tokenFile)
	if err != nil {
		return err
	}
	b := &channel.Bridge{URL: *base, Token: strings.TrimSpace(string(token)), In: stdin, Out: stdout, Log: log.New(os.Stderr, "task-ui channel: ", log.LstdFlags)}
	return b.Run(ctx)
}

// getSecret runs task-ui get-secret <name> in a run: the content of a
// secret that reaches the run's folder on stdout, with nothing after it.
func getSecret(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: task-ui get-secret <name>, in a run task-ui started")
	}
	client, err := secret.ClientFromEnv(getenv)
	if err != nil {
		return err
	}
	value, err := client.Get(ctx, args[0])
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, value)
	return err
}

// run is main without the exit: the hash, agent, channel, cleanup,
// migrate-layout, get-secret, passwd, email, backup, token and user commands, the
// version, or the server until a signal.
func run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "get-secret" {
		return getSecret(ctx, args[1:], os.Getenv, stdout)
	}
	if len(args) > 0 && args[0] == "hash" {
		return hash(args[1:], stdin, stdout)
	}
	if len(args) > 0 && args[0] == "channel" {
		return channelBridge(ctx, args[1:], stdin, stdout)
	}
	if len(args) > 0 && args[0] == "cleanup" {
		return cleanup(ctx, args[1:], stdout)
	}
	if len(args) > 0 && args[0] == "migrate-layout" {
		return migrateLayout(ctx, args[1:], stdout)
	}
	if len(args) > 0 && args[0] == "agent" {
		a := &agent.Agent{Env: os.Environ(), Stdin: stdin, Stdout: stdout, Stderr: os.Stderr}
		return a.Run(ctx, args[1:])
	}
	if len(args) > 0 && args[0] == "user" {
		// Its --force is no server flag: the configuration reads none.
		cfg, err := app.NewConfig(nil)
		if err != nil {
			return err
		}
		return userCommand(ctx, cfg.DB, args[1:], stdout)
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
		// A host command ends with the code of the command it ran, which
		// said what went wrong itself.
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "task-ui:", err)
		os.Exit(1)
	}
}
