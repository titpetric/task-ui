package api

import (
	"encoding/json"
	"errors"

	"github.com/titpetric/task-ui/argv"
)

// RunArgs are the arguments a run passes its job after --. They come as
// one line, which argv.Split reads the way a shell splits words with
// nothing expanded, or as a list, taken unchanged: Run again sends a
// run row's args so. Either holds to argv.Check.
type RunArgs []string

// UnmarshalJSON reads a line or a list.
func (a *RunArgs) UnmarshalJSON(data []byte) error {
	var line string
	if err := json.Unmarshal(data, &line); err == nil {
		args, err := argv.Split(line)
		if err != nil {
			return err
		}
		*a = args
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return errors.New("args are a line or a list of strings")
	}
	if err := argv.Check(list); err != nil {
		return err
	}
	*a = list
	return nil
}
