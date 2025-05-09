package server

import (
	"html/template"

	"github.com/titpetric/task-ui/server/model"
)

type templateData struct {
	Tasks   []*model.TaskInfo
	Current *model.TaskInfo
}

func (f *templateData) FuncMap() template.FuncMap {
	return template.FuncMap{
		"IsCurrent": f.isCurrent,
	}
}

func (f *templateData) isCurrent(item *model.TaskInfo) bool {
	if f.Current == nil || item == nil {
		return false
	}
	return f.Current.Task == item.Task
}
