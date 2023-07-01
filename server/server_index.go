package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/render"

	"github.com/titpetric/task-ui/server/config"

	. "github.com/titpetric/task-ui/server/model"
	. "github.com/titpetric/task-ui/server/repository"
)

func (svc *Server) Index(w http.ResponseWriter, r *http.Request) {
	serverError := func(err error) {
		render.JSON(w, r, InternalServerError(err))
	}

	spec, err := config.Load(".", "Taskfile.yml")
	if err != nil {
		serverError(err)
		return
	}

	var (
		out  = new(bytes.Buffer)
		data = &templateData{
			Tasks: ListTasks(spec, FilterOutInternal, FilterOutNoDesc),
		}
	)

	if err := svc.template(out, indexTemplate, data); err != nil {
		serverError(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, out)
}
