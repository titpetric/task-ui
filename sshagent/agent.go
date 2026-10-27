// Package sshagent serves the private keys of a directory to ssh, as an
// agent on a socket of its own: a checkout's .ssh/ folder, loaded for the
// runs of that checkout. The keys stay in memory; ssh, git, rsync and
// terraform reach them through SSH_AUTH_SOCK.
package sshagent

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Agent is an ssh agent holding the keys of one directory. A tmp
// directory of "" stands for Dir.
type Agent struct {
	// Socket is the path ssh connects to, for SSH_AUTH_SOCK.
	Socket string
	// Keys are the paths of the files the keys came from.
	Keys []string

	dir      string
	listener net.Listener
	wg       sync.WaitGroup
}

// Load reads the private keys in a directory and serves them, as
// LoadFiles does with every file in it.
func Load(keys, tmp string) (*Agent, error) {
	entries, err := os.ReadDir(keys)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			paths = append(paths, filepath.Join(keys, entry.Name()))
		}
	}
	return LoadFiles(tmp, paths...)
}

// LoadKeys serves the private keys given, from memory, on a socket in a
// new private directory under tmp ("" is the system's), as LoadFiles
// does; Keys holds their names. It writes no key file.
func LoadKeys(tmp string, keys ...Named) (*Agent, error) {
	ring := agent.NewKeyring()
	var loaded []string
	for _, key := range keys {
		if slices.Contains(loaded, key.Name) {
			continue
		}
		private, err := ssh.ParseRawPrivateKey(key.PEM)
		if err != nil {
			return nil, fmt.Errorf("the key %s: %w", key.Name, err)
		}
		if err := ring.Add(agent.AddedKey{PrivateKey: private, Comment: key.Name}); err != nil {
			return nil, err
		}
		loaded = append(loaded, key.Name)
	}
	if len(loaded) == 0 {
		return nil, nil
	}
	return serveRing(tmp, ring, loaded)
}

// LoadFiles serves the private keys of the files given on a socket in a new private directory under tmp.
// It skips a file given twice and each file of no unencrypted private key, such as public halves and known_hosts.
// No key returns nil and no error.
func LoadFiles(tmp string, paths ...string) (*Agent, error) {
	ring := agent.NewKeyring()
	var loaded []string
	seen := map[string]bool{}
	for _, path := range paths {
		if seen[path] || strings.HasSuffix(path, ".pub") {
			continue
		}
		seen[path] = true
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		key, err := ssh.ParseRawPrivateKey(data)
		if err != nil {
			continue
		}
		if err := ring.Add(agent.AddedKey{PrivateKey: key, Comment: filepath.Base(path)}); err != nil {
			return nil, err
		}
		loaded = append(loaded, path)
	}
	if len(loaded) == 0 {
		return nil, nil
	}
	return serveRing(tmp, ring, loaded)
}

// Prefix starts the name of every agent's directory, so a start of
// task-ui removes the ones a process before it left.
const Prefix = "task-ui-agent-"

// serveRing serves a keyring on a socket in a new private directory.
func serveRing(tmp string, ring agent.Agent, loaded []string) (*Agent, error) {
	if tmp == "" {
		tmp = Dir()
	}
	dir, err := os.MkdirTemp(tmp, Prefix+strconv.Itoa(os.Getpid())+"-")
	if err != nil {
		return nil, err
	}
	socket := filepath.Join(dir, "agent.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	a := &Agent{Socket: socket, Keys: loaded, dir: dir, listener: listener}
	a.wg.Add(1)
	go a.serve(ring)
	return a, nil
}

// serve answers every connection until the listener closes.
func (a *Agent) serve(ring agent.Agent) {
	defer a.wg.Done()
	for {
		conn, err := a.listener.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			_ = agent.ServeAgent(ring, conn)
		}()
	}
}

// Close stops the agent and removes its socket. Safe on nil.
func (a *Agent) Close() error {
	if a == nil {
		return nil
	}
	err := a.listener.Close()
	a.wg.Wait()
	if rmErr := os.RemoveAll(a.dir); rmErr != nil {
		err = errors.Join(err, rmErr)
	}
	return err
}

// Dir is where agents put their sockets by default: the user's runtime
// directory, else /run/task-ui, else the system's temporary directory,
// which a test run or a tmp cleaner may wipe.
func Dir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	if err := os.MkdirAll("/run/task-ui", 0o700); err == nil {
		if f, err := os.CreateTemp("/run/task-ui", ".probe"); err == nil {
			f.Close()
			os.Remove(f.Name())
			return "/run/task-ui"
		}
	}
	return os.TempDir()
}

// Clean removes the agent directories that earlier processes left under tmp ("" is Dir), and returns how many.
// A directory of a process that still runs, another than this one, stays: it holds that process's live agents.
func Clean(tmp string) int {
	if tmp == "" {
		tmp = Dir()
	}
	dirs, _ := filepath.Glob(filepath.Join(tmp, Prefix+"*"))
	removed := 0
	for _, dir := range dirs {
		if owned(filepath.Base(dir)) {
			continue
		}
		if os.RemoveAll(dir) == nil {
			removed++
		}
	}
	return removed
}

// owned reports whether the directory name belongs to another process that is alive.
func owned(name string) bool {
	pid, _, _ := strings.Cut(strings.TrimPrefix(name, Prefix), "-")
	n, err := strconv.Atoi(pid)
	if err != nil || n == os.Getpid() {
		return false
	}
	err = syscall.Kill(n, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
