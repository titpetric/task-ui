package access

import (
	"testing"
)

func TestUsers_Find(t *testing.T) {
	users := Users{{Name: "admin", Role: RoleAdmin}, {Name: "ops", Role: RoleRunner}}
	if u := users.Find("ops"); u == nil || u.Role != RoleRunner {
		t.Fatalf("find: %+v", u)
	}
	if users.Find("nobody") != nil {
		t.Fatal("missing user found")
	}
	// A copy comes back, so a caller cannot change the list through it.
	users.Find("ops").Role = RoleAdmin
	if users[1].Role != RoleRunner {
		t.Fatal("Find returned the list's own entry")
	}
}

func TestUsers_Validate(t *testing.T) {
	hash, _ := HashPassword("x")
	good := Users{{Name: "a", Password: hash}, {Name: "b", Password: hash, Role: RoleAdmin}}
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
	if good[0].Role != RoleViewer {
		t.Fatal("missing role did not default to viewer")
	}
	for name, bad := range map[string]Users{
		"no name":    {{Password: hash}},
		"duplicate":  {{Name: "a", Password: hash}, {Name: "a", Password: hash}},
		"bad role":   {{Name: "a", Password: hash, Role: "root"}},
		"plain text": {{Name: "a", Password: "secret"}},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
