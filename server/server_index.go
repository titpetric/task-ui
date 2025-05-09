package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/render"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

func (svc *Server) Index(w http.ResponseWriter, r *http.Request) {
	serverError := func(err error) {
		render.JSON(w, r, InternalServerError(err))
	}

	spec, err := model.LoadTaskfile("Taskfile.yml")
	if err != nil {
		serverError(err)
		return
	}

	var (
		out  = new(bytes.Buffer)
		data = &templateData{
			Tasks: repository.ListTasks(spec, repository.FilterOutInternal, repository.FilterOutNoDesc),
		}
	)

	if err := svc.template(out, indexTemplate, data); err != nil {
		serverError(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, out)
}
