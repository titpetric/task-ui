package view

import (
	"encoding/json"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// SidePin is a pin of the user as the sidebar places it: the pin, the
// folder it names, and for a job pin the job, with the directory it runs
// in and its args.
type SidePin struct {
	Pin    model.Pin
	Folder *folder.Folder
	// Job is the pinned job, nil for a folder pin.
	Job *folder.Job
}

// Args returns the args the pinned job runs with, none for args it cannot read.
func (s SidePin) Args() []string {
	var args []string
	_ = json.Unmarshal([]byte(s.Pin.Args), &args)
	return args
}

// ArgsJSON returns the args as the run button carries them, "" for none.
func (s SidePin) ArgsJSON() string {
	if len(s.Args()) == 0 {
		return ""
	}
	return s.Pin.Args
}

// Label returns a pinned job's name, after its directory, before its args.
func (s SidePin) Label() string {
	label := s.Pin.Job
	if s.Pin.Dir != "" {
		label = s.Pin.Dir + "/ " + label
	}
	if args := s.Args(); len(args) > 0 {
		label += " " + strings.Join(args, " ")
	}
	return label
}

// plain reports whether the pin is a job its folder's row lists as is:
// one of the folder's jobs, no directory, no args.
func (s SidePin) plain(f *folder.Folder) bool {
	if s.Pin.Dir != "" || len(s.Args()) > 0 {
		return false
	}
	for _, job := range f.Jobs {
		if job.ID == s.Pin.Job {
			return true
		}
	}
	return false
}
