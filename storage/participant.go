package storage

import (
	"time"
)

// Participant is one user of a chat and the time they last read it.
type Participant struct {
	UserID     string
	UserName   string
	LastReadAt *time.Time
}
