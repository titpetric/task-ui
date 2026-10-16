package access

import (
	"testing"
)

func TestRole_Valid(t *testing.T) {
	for _, role := range []Role{RoleAdmin, RoleRunner, RoleViewer} {
		if !role.Valid() {
			t.Errorf("%s invalid", role)
		}
	}
	if Role("root").Valid() || Role("").Valid() {
		t.Fatal("unknown role valid")
	}
}
