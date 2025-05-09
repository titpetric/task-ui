package server

import (
	"fmt"
	"io"
	"os"
	"strings"

	"io/ioutil"

	"github.com/apex/log"
	"github.com/go-chi/chi"
	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

func (svc *Server) Launch() websocket.Handler {
	logError := func(ws *websocket.Conn, message string, err error) {
		log.WithError(err).Info(message)
		ws.Write([]byte(fmt.Sprintf("%s: %s\r\n", message, err)))
	}
	return websocket.Handler(func(ws *websocket.Conn) {
		if !svc.semaphore.CanRun() {
			ws.Write([]byte("Something else is running at the moment, try a bit later\r\n"))
			return
		}
		defer svc.semaphore.Done()

		r := ws.Request()
		id := chi.URLParam(r, "id")

		spec, err := model.LoadTaskfile(svc.config.Taskfile)
		if err != nil {
			logError(ws, "error loading "+svc.config.Taskfile, err)
			return
		}

		action, err := repository.FindTask(spec, id)
		if err != nil {
			logError(ws, "can't find task by id", err)
			return
		}

		if err := svc.launchTask(ws, id, action.Interactive); err != nil {
			logError(ws, "can't run command", err)
			return
		}

		ws.Write([]byte("done.\r\n"))
	})
}

func (svc *Server) launchTask(ws *websocket.Conn, id string, interactive bool) error {
	// parse commands into command+args
	commands := strings.Split("task "+id, " ")
	command := commands[0]
	commandArgs := []string{"-l", "-c", command + " " + strings.Join(commands[1:], " ")}

	// terminal environment
	env := os.Environ()
	env = append(env, []string{
		"HOME=/root",
		"TERM=xterm",
		"COLUMNS=120",
		"ROWS=20",
	}...)

	proc, err := func() (*launchedProcess, error) {
		if svc.config.EnableHistory {
			return record(id, "/bin/bash", commandArgs, env)
		}
		return launch(id, "/bin/bash", commandArgs, env)
	}()

	if err != nil {
		return err
	}

	defer proc.tty.Close()

	go func() {
		if interactive {
			io.Copy(proc.tty, ws)
			return
		}
		io.Copy(ioutil.Discard, ws)
	}()
	io.Copy(ws, proc.tty)

	proc.command.Process.Kill()
	proc.command.Wait()
	return nil
}
