package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// KeyReadKeep is how many reads of one key the log keeps.
const KeyReadKeep = 1000

// KeyReadStorage keeps who read which private key, encrypted to their
// recipient: the key, the user, the token and the way, MCP or the API,
// the newest KeyReadKeep of each key. Never the key or the recipient.
type KeyReadStorage struct {
	db DB
}

// NewKeyReadStorage returns the storage over a pool.
func NewKeyReadStorage(db DB) *KeyReadStorage {
	return &KeyReadStorage{db: db}
}

// Record adds a read, then prunes the key's log to the newest
// KeyReadKeep.
func (s *KeyReadStorage) Record(ctx context.Context, read *model.KeyRead) error {
	if read.KeyName == "" {
		return errors.New("a key read needs a key")
	}
	if read.Via == "" {
		read.Via = "api"
	}
	read.ID, read.CreatedAt = model.NewID(), stamp()
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Insert(ctx, model.KeyReadTable, read); err != nil {
			return err
		}
		return c.Exec(ctx, "DELETE FROM key_read WHERE key_name = ? AND rowid NOT IN (SELECT rowid FROM key_read WHERE key_name = ? ORDER BY rowid DESC LIMIT ?)", read.KeyName, read.KeyName, KeyReadKeep)
	})
}

// keyReadCount is a key's number of reads, as Counts reads it.
type keyReadCount struct {
	KeyName string `db:"key_name"`
	Count   int    `db:"n"`
}

// Counts returns how many reads the log holds of each key, by name.
func (s *KeyReadStorage) Counts(ctx context.Context) (map[string]int, error) {
	rows, err := client(s.db).Select[keyReadCount](ctx, "SELECT key_name, COUNT(*) AS n FROM key_read GROUP BY key_name")
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.KeyName] = row.Count
	}
	return counts, nil
}

// Last returns the newest read of each key, by name.
func (s *KeyReadStorage) Last(ctx context.Context) (map[string]model.KeyRead, error) {
	rows, err := client(s.db).Select[model.KeyRead](ctx, "SELECT * FROM key_read WHERE rowid IN (SELECT MAX(rowid) FROM key_read GROUP BY key_name)")
	if err != nil {
		return nil, err
	}
	last := map[string]model.KeyRead{}
	for _, row := range rows {
		last[row.KeyName] = row
	}
	return last, nil
}

// All returns every read, oldest first, for the data retention rules.
func (s *KeyReadStorage) All(ctx context.Context) ([]model.KeyRead, error) {
	return client(s.db).Select[model.KeyRead](ctx, "SELECT * FROM key_read ORDER BY rowid ASC")
}

// DeleteThrough removes the reads up to and including the one named, in the order of All.
func (s *KeyReadStorage) DeleteThrough(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM key_read WHERE rowid <= (SELECT rowid FROM key_read WHERE id = ?)", id)
}
