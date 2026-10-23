package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// PermissionPolicyStorage keeps the permission policies of the roles and
// the users, by name. It caches a subject's policy until a write through
// it changes the subject's entries, so the app keeps one instance: the
// one UserStorage holds.
type PermissionPolicyStorage struct {
	db DB

	mu    sync.Mutex
	cache map[string]permission.Policy
}

// NewPermissionPolicyStorage returns the storage over a pool.
func NewPermissionPolicyStorage(db DB) *PermissionPolicyStorage {
	return &PermissionPolicyStorage{db: db, cache: map[string]permission.Policy{}}
}

func policyKey(kind, subject string) string {
	return kind + "\x00" + subject
}

// Policy returns the entries of a subject, a role or a user by name. The
// caller shares the policy with the cache: it reads it and changes
// nothing.
func (s *PermissionPolicyStorage) Policy(ctx context.Context, kind, subject string) (permission.Policy, error) {
	key := policyKey(kind, subject)
	s.mu.Lock()
	cached, ok := s.cache[key]
	s.mu.Unlock()
	if ok {
		return cached, nil
	}
	rows, err := client(s.db).Select[model.PermissionPolicy](ctx, "SELECT * FROM permission_policy WHERE subject_kind = ? AND subject = ?", kind, subject)
	if err != nil {
		return nil, err
	}
	policy := permission.Policy{}
	for _, row := range rows {
		policy[row.Path] = row.Effect
	}
	s.mu.Lock()
	s.cache[key] = policy
	s.mu.Unlock()
	return policy, nil
}

// Put sets a subject's entry for a path: permission.Allow or
// permission.Deny, or "" to remove it, so the path inherits again. The
// path is one permissions.yml holds; the admin role holds no entries.
func (s *PermissionPolicyStorage) Put(ctx context.Context, kind, subject, path, effect string) error {
	switch {
	case kind != permission.KindRole && kind != permission.KindUser:
		return fmt.Errorf("no subject kind %q", kind)
	case subject == "":
		return errors.New("a policy needs a subject")
	case kind == permission.KindRole && !model.ValidRole(subject):
		return errors.New("unknown role " + subject)
	case kind == permission.KindRole && subject == model.RoleAdmin:
		return errors.New("the admin role can do everything and has no policy to change")
	case !config.Default().Valid(path):
		return fmt.Errorf("%w: %q", config.ErrUnknown, path)
	case effect != "" && effect != permission.Allow && effect != permission.Deny:
		return fmt.Errorf("no effect %q: allow, deny or none", effect)
	}
	defer s.forget(kind, subject)
	db := client(s.db)
	if err := db.Exec(ctx, "DELETE FROM permission_policy WHERE subject_kind = ? AND subject = ? AND path = ?", kind, subject, path); err != nil {
		return err
	}
	if effect == "" {
		return nil
	}
	row := &model.PermissionPolicy{ID: model.NewID(), SubjectKind: kind, Subject: subject, Path: path, Effect: effect, CreatedAt: stamp()}
	return db.Insert(ctx, model.PermissionPolicyTable, row)
}

// DeleteSubject removes every entry of a subject, as when its user goes.
func (s *PermissionPolicyStorage) DeleteSubject(ctx context.Context, kind, subject string) error {
	defer s.forget(kind, subject)
	return client(s.db).Exec(ctx, "DELETE FROM permission_policy WHERE subject_kind = ? AND subject = ?", kind, subject)
}

// forget drops a subject's cached policy.
func (s *PermissionPolicyStorage) forget(kind, subject string) {
	s.mu.Lock()
	delete(s.cache, policyKey(kind, subject))
	s.mu.Unlock()
}

// For returns what a user may do, from the user's policy and the role's.
// A policy that does not read leaves the user with what the other allows, and an admin with everything.
func (s *PermissionPolicyStorage) For(ctx context.Context, user *model.User) permission.Set {
	set := permission.Set{User: user}
	if user == nil || user.CanAdmin() {
		return set
	}
	set.Role, _ = s.Policy(ctx, permission.KindRole, user.Role)
	set.Own, _ = s.Policy(ctx, permission.KindUser, user.Name)
	return set
}
