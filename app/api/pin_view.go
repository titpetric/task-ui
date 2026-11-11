package api

import (
	"encoding/json"
	"time"

	"github.com/titpetric/task-ui/model"
)

// PinView is a pin as the API shows it: its target, the args a list.
type PinView struct {
	ID        string     `json:"id"`
	Folder    string     `json:"folder"`
	Job       string     `json:"job"`
	Dir       string     `json:"dir"`
	Args      []string   `json:"args"`
	CreatedAt *time.Time `json:"created_at"`
}

// NewPinView returns the view of a pin; args it cannot read show as none.
func NewPinView(pin model.Pin) PinView {
	args := []string{}
	_ = json.Unmarshal([]byte(pin.Args), &args)
	if args == nil {
		args = []string{}
	}
	return PinView{ID: pin.ID, Folder: pin.Folder, Job: pin.Job, Dir: pin.Dir, Args: args, CreatedAt: pin.CreatedAt}
}
