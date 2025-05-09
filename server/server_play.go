package server

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/apex/log"
	"github.com/go-chi/chi"
	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

func (svc *Server) Play(replay bool) websocket.Handler {
	logError := func(ws *websocket.Conn, message string, err error) {
		log.WithError(err).Info(message)
		ws.Write([]byte(fmt.Sprintf("%s: %s\r\n", message, err)))
	}
	return websocket.Handler(func(ws *websocket.Conn) {
		r := ws.Request()

		historyID := chi.URLParam(r, "id")

		task, timestamp, err := func() (string, string, error) {
			s := strings.Split(historyID, "-")
			if len(s) != 2 {
				return "", "", fmt.Errorf("expected 2 parts, got %d", len(s))
			}
			return s[0], s[1], nil
		}()
		if err != nil {
			logError(ws, "invalid id", err)
		}

		if ts, err := strconv.ParseInt(timestamp, 10, 64); ts == 0 {
			logError(ws, "invalid history id timestamp", err)
			return
		}

		spec, err := model.LoadTaskfile(svc.config.Taskfile)
		if err != nil {
			logError(ws, "error loading "+svc.config.Taskfile, err)
			return
		}

		if _, err := repository.FindTask(spec, task); err != nil {
			logError(ws, "can't find task by id", err)
			return
		}

		if err := svc.viewHistory(ws, historyID, replay); err != nil {
			logError(ws, "can't run command", err)
			return
		}

		ws.Write([]byte("done.\r\n"))
	})
}

func (svc *Server) viewHistory(ws *websocket.Conn, id string, replay bool) error {
	var interactive bool

	command := fmt.Sprintf("ttyplay -n history/%s.ttyrec", id)
	if replay {
		command = fmt.Sprintf("ttyplay history/%s.ttyrec", id)
	}
	commandArgs := []string{"-l", "-c", command}

	// terminal environment
	env := os.Environ()
	env = append(env, []string{
		"HOME=/root",
		"TERM=xterm",
		"COLUMNS=120",
		"ROWS=20",
	}...)

	proc, err := func() (*launchedProcess, error) {
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
