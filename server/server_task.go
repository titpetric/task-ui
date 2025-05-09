package server

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

func (svc *Server) Task(w http.ResponseWriter, r *http.Request) {
	serverError := func(err error) {
		render.JSON(w, r, model.InternalServerError(err))
	}
	notFoundError := func(err error) {
		render.JSON(w, r, model.NotFoundError(err))
	}

	spec, err := model.LoadTaskfile("Taskfile.yml")
	if err != nil {
		serverError(err)
		return
	}

	id := chi.URLParam(r, "id")
	task, err := repository.FindTask(spec, id)
	if err != nil {
		notFoundError(err)
		return
	}

	taskinfo := model.NewTaskInfo(task)
	taskinfo.History = repository.LoadHistory(id)

	var (
		out  = new(bytes.Buffer)
		data = &templateData{
			Tasks:   repository.ListTasks(spec, repository.FilterOutInternal, repository.FilterOutNoDesc),
			Current: taskinfo,
		}
	)

	if err := svc.template(out, taskTemplate, data); err != nil {
		if errors.Is(err, ErrNotFound) {
			notFoundError(err)
			return
		}
		serverError(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, out)
}
