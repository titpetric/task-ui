package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Git runs git for the folders: the clone that adds a project, the pull
// before every run, and the facts the pages show. It authenticates with
// the keys in SSHDir, the ones the settings manage.
type Git struct {
	SSHDir string
}

// ErrNotRepo reports a folder without a .git directory.
var ErrNotRepo = errors.New("not a git repository")

// IsRepo reports whether dir is a git checkout.
func (g *Git) IsRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

// Checkout returns the checkout dir belongs to: dir itself when dir
// is one, or the nearest of its parents, looking up at most levels of them.
// A folder's levels are its depth below the root, so the search stays
// inside the root. "" when there is none.
func (g *Git) Checkout(dir string, levels int) string {
	for i := 0; i <= levels; i++ {
		if g.IsRepo(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// env returns the environment git runs with: no prompts, and ssh using
// the instance keys when there are any.
func (g *Git) env() []string {
	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if cmd := g.sshCommand(); cmd != "" {
		env = append(env, "GIT_SSH_COMMAND="+cmd)
	}
	return env
}

// sshCommand builds the ssh command for git: every private key in SSHDir,
// and a known_hosts file next to them that accepts a host on first contact.
func (g *Git) sshCommand() string {
	if g.SSHDir == "" {
		return ""
	}
	entries, err := os.ReadDir(g.SSHDir)
	if err != nil {
		return ""
	}
	var keys []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "id_") || strings.HasSuffix(name, ".pub") || strings.Contains(name, ".bak-") {
			continue
		}
		keys = append(keys, filepath.Join(g.SSHDir, name))
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)
	parts := []string{"ssh", "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile=" + shellQuote(filepath.Join(g.SSHDir, "known_hosts"))}
	for _, key := range keys {
		parts = append(parts, "-i", shellQuote(key))
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (g *Git) run(ctx context.Context, dir string, w io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = g.env()
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func (g *Git) output(dir string, args ...string) (string, error) {
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := g.run(ctx, dir, &out, args...); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// Info returns the remote, branch and last commit of a checkout, or nil
// for a folder that is not one.
func (g *Git) Info(dir string) *Info {
	if !g.IsRepo(dir) {
		return nil
	}
	info := &Info{}
	info.Remote, _ = g.output(dir, "remote", "get-url", "origin")
	info.Branch, _ = g.output(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if last, err := g.output(dir, "log", "-1", "--format=%h%x00%s%x00%cI"); err == nil {
		parts := strings.SplitN(last, "\x00", 3)
		if len(parts) == 3 {
			info.Commit = parts[0]
			info.Subject = parts[1]
			info.When, _ = time.Parse(time.RFC3339, parts[2])
		}
	}
	return info
}

// ValidURL reports whether url is something git clone accepts from a
// form: an https, ssh, git or file URL, or the scp-like user@host:path.
func (g *Git) ValidURL(url string) bool {
	url = strings.TrimSpace(url)
	if url == "" || strings.ContainsAny(url, " \t\n\r") || strings.HasPrefix(url, "-") {
		return false
	}
	for _, prefix := range []string{"https://", "http://", "ssh://", "git://", "file://"} {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	host, path, ok := strings.Cut(url, ":")
	return ok && strings.Contains(host, "@") && !strings.Contains(host, "/") && path != ""
}

// Clone clones url into dir, which must not exist, and returns git's
// output. A failed clone leaves nothing behind.
func (g *Git) Clone(ctx context.Context, url, dir, branch string) (string, error) {
	if !g.ValidURL(url) {
		return "", fmt.Errorf("%q is not a git URL", url)
	}
	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("%s exists already", filepath.Base(dir))
	}
	args := []string{"clone", "--progress"}
	if branch = strings.TrimSpace(branch); branch != "" {
		if strings.HasPrefix(branch, "-") {
			return "", fmt.Errorf("%q is not a branch name", branch)
		}
		args = append(args, "--branch", branch)
	}
	args = append(args, "--", strings.TrimSpace(url), dir)

	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	err := g.run(ctx, filepath.Dir(dir), &out, args...)
	output := strings.TrimSpace(out.String())
	if err != nil {
		os.RemoveAll(dir)
		if output == "" {
			output = err.Error()
		}
		return output, fmt.Errorf("git clone failed: %s", lastLine(output))
	}
	return output, nil
}

// Pull fast forwards a checkout, writing git's output to w.
func (g *Git) Pull(ctx context.Context, dir string, w io.Writer) error {
	if !g.IsRepo(dir) {
		return ErrNotRepo
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	return g.run(ctx, dir, w, "pull", "--ff-only", "--no-rebase")
}

// Remove deletes a cloned folder. It refuses a folder that is not a
// checkout: the instance did not clone that folder, so it does not delete it.
func (g *Git) Remove(dir string) error {
	if !g.IsRepo(dir) {
		return ErrNotRepo
	}
	return os.RemoveAll(dir)
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
