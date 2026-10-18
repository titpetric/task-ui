// Package sshagent serves the private keys of a directory to ssh, as an
// agent on a socket of its own: a checkout's .ssh/ folder, loaded for the
// runs of that checkout. The keys stay in memory; ssh, git, rsync and
// terraform reach them through SSH_AUTH_SOCK.
package sshagent

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Agent is an ssh agent holding the keys of one directory.
type Agent struct {
	// Socket is the path ssh connects to, for SSH_AUTH_SOCK.
	Socket string
	// Keys are the names of the files the keys came from.
	Keys []string

	dir      string
	listener net.Listener
	wg       sync.WaitGroup
}

// Load reads the private keys in keys and serves them on a socket in a new
// private directory under tmp ("" is the system's). Files that are not an
// unencrypted private key, public halves, known_hosts and config among
// them, are skipped. A directory without a key returns nil and no error.
func Load(keys, tmp string) (*Agent, error) {
	ring := agent.NewKeyring()
	var names []string
	entries, err := os.ReadDir(keys)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasSuffix(name, ".pub") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(keys, name))
		if err != nil {
			continue
		}
		key, err := ssh.ParseRawPrivateKey(data)
		if err != nil {
			continue
		}
		if err := ring.Add(agent.AddedKey{PrivateKey: key, Comment: name}); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, nil
	}

	dir, err := os.MkdirTemp(tmp, "task-ui-agent-")
	if err != nil {
		return nil, err
	}
	socket := filepath.Join(dir, "agent.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	a := &Agent{Socket: socket, Keys: names, dir: dir, listener: listener}
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
