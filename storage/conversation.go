package storage

import (
	"github.com/titpetric/task-ui/app/model"
)

// Conversation is a root note with its newest activity and reply count.
type Conversation struct {
	model.Note
	LastAt   string `db:"last_at"`
	Messages int    `db:"messages"`
}
