package server

import (
	"fmt"
	"io"
	"os"
	"time"

	"os/exec"

	"github.com/creack/pty"
)

type launchedProcess struct {
	command *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	tty     *os.File
}

func record(actionName, commandName string, args []string, env []string) (*launchedProcess, error) {
	newName := "ttyrec"
	outfile := fmt.Sprintf("history/%s-%d.ttyrec", actionName, time.Now().Unix())
	newArgs := []string{
		"-T",
		"always",
		"-f",
		outfile,
		"--",
		commandName,
	}
	newArgs = append(newArgs, args...)

	return launch(actionName, newName, newArgs, env)
}

func launch(actionName, commandName string, commandArgs []string, env []string) (*launchedProcess, error) {
	command := exec.Command(commandName, commandArgs...)
	command.Env = env

	tty, err := pty.Start(command)
	if err != nil {
		return nil, err
	}

	return &launchedProcess{
		command: command,
		tty:     tty,
	}, err
}
