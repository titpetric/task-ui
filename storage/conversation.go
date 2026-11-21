package storage

import (
	"github.com/titpetric/task-ui/app/model"
)

// Conversation adds newest activity and a reply count to a root note.
type Conversation struct {
	model.Note
	LastAt   string  `db:"last_at" json:"last_at"`
	Messages int     `db:"messages" json:"messages"`
	Tools    int     `json:"tools"`
	Tokens   int     `json:"tokens"`
	Cost     float64 `json:"cost"`
}
