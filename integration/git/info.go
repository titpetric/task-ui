package git

import (
	"time"
)

// Info describes a folder that is a git checkout.
type Info struct {
	Remote  string    `json:"remote"`
	Branch  string    `json:"branch"`
	Commit  string    `json:"commit"`
	Subject string    `json:"subject"`
	When    time.Time `json:"when"`
}
