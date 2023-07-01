package config

import (
	"github.com/go-task/task/v3/taskfile"
	"github.com/go-task/task/v3/taskfile/read"
)

func NewTaskfileLoader() *TaskfileLoader {
	return &TaskfileLoader{}
}

type TaskfileLoader struct{}

func (*TaskfileLoader) Load(folder, filename string) (*taskfile.Taskfile, error) {
	taskfile, _, err := read.Taskfile(&read.ReaderNode{
		Dir:        folder,
		Entrypoint: filename,
	})
	return taskfile, err
}
