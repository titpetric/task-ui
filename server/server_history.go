package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

func (svc *Server) History(w http.ResponseWriter, r *http.Request) {
	response := model.NewHistoryResponse()

	id := chi.URLParam(r, "id")
	match := "history/*.ttyrec"
	if id != "" {
		spec, err := model.LoadTaskfile(svc.config.Taskfile)
		if err != nil {
			render.JSON(w, r, NotFoundError(err))
			return
		}

		if _, err := repository.FindTask(spec, id); err != nil {
			render.JSON(w, r, NotFoundError(err))
			return
		}

		match = "history/" + id + "-*.ttyrec"
	}

	files, err := filepath.Glob(match)
	if err != nil {
		render.JSON(w, r, InternalServerError(err))
		return
	}

	response.Files = files

	repository.FillHistory(response, files)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		render.JSON(w, r, InternalServerError(err))
	}
}
