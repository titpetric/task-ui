package storage

import (
	"time"
)

// ChatSummary is a chat thread as one participant sees it.
type ChatSummary struct {
	ID            string     `db:"id" json:"id"`
	Peer          string     `db:"peer" json:"peer"`
	LastID        string     `db:"last_id" json:"last_id"`
	LastUserID    string     `db:"last_user_id" json:"last_user_id"`
	LastUserName  string     `db:"last_user_name" json:"last_user_name"`
	LastBody      string     `db:"last_body" json:"last_body"`
	LastSpoken    string     `db:"last_spoken" json:"last_spoken"`
	LastAudioPath string     `db:"last_audio_path" json:"last_audio_path"`
	LastAt        *time.Time `db:"last_at" json:"last_at"`
	LastReadAt    *time.Time `db:"last_read_at" json:"last_read_at"`
	Unread        int64      `db:"unread" json:"unread"`
}
