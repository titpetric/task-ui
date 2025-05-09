package repository

import (
	"fmt"

	. "github.com/titpetric/task-ui/server/model"
)

// FilterFunc is a signature for task filters.
type FilterFunc func(task *Task) bool

// List tasks lists the names of the defined tasks.
func ListTasks(spec *Taskfile, filters ...FilterFunc) []*TaskInfo {
	result := make([]*TaskInfo, 0, len(spec.Tasks))
	for _, task := range spec.Tasks {
		for _, filter := range filters {
			if filter(task) {
				goto next
			}
		}

		result = append(result, NewTaskInfo(task))

	next:
	}
	return result
}

// FilterOutNoDesc removes all tasks that do not contain a description.
func FilterOutNoDesc(task *Task) bool {
	return task.Description == ""
}

// FilterOutInternal removes all tasks that are marked as internal.
func FilterOutInternal(task *Task) bool {
	return task.Internal
}

// FindTask returns the Task for a given name.
func FindTask(spec *Taskfile, name string) (*Task, error) {
	for _, task := range spec.Tasks {
		if task.Name == name {
			return task, nil
		}
	}
	return nil, fmt.Errorf("no such task: %s: %w", name, ErrNotFound)
}
