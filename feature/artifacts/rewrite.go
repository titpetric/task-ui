package artifactstore

import (
	"context"
)

// Rewrite is what Expire did to the history: the commits it dropped,
// the new root, and the new name of every commit it kept.
type Rewrite struct {
	// Root is the new first commit, of the tree Expire kept from.
	Root string
	// Renamed maps each commit it kept to its new name, Root among them.
	Renamed map[string]string
	// Dropped are the commits before the one it kept from, oldest first.
	Dropped []string

	store *Store
	ctx   context.Context
}

// Same reports whether Root holds a path's file as a dropped commit
// did: a version of a commit that went that lives on in the tree.
func (r Rewrite) Same(commit, rel string) bool {
	if r.store == nil {
		return false
	}
	before, err := r.store.git(r.ctx, "rev-parse", "--verify", "-q", commit+":"+rel)
	if err != nil || before == "" {
		return false
	}
	after, err := r.store.git(r.ctx, "rev-parse", "--verify", "-q", r.Root+":"+rel)
	return err == nil && after == before
}
