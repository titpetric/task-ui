package access

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
)

func TestUser_CanRun(t *testing.T) {
	open := &folder.Folder{ID: "web"}
	locked := &folder.Folder{ID: "infra", Allow: []string{"ops"}}
	build := &folder.Job{ID: "build"}
	certs := &folder.Job{ID: "certs", Allow: []string{"admin"}}

	admin := &User{Name: "root", Role: RoleAdmin}
	ops := &User{Name: "ops", Role: RoleRunner}
	dev := &User{Name: "dev", Role: RoleRunner}
	scoped := &User{Name: "web", Role: RoleRunner, Folders: []string{"web"}}
	guest := &User{Name: "guest", Role: RoleViewer}

	cases := []struct {
		user *User
		f    *folder.Folder
		j    *folder.Job
		want bool
	}{
		{admin, locked, certs, true},
		{ops, locked, build, true},
		{ops, locked, certs, false},
		{dev, open, build, true},
		{dev, locked, build, false},
		{scoped, open, build, true},
		{scoped, locked, build, false},
		{guest, open, build, false},
		{nil, open, build, false},
		{admin, nil, build, false},
	}
	for i, c := range cases {
		if got := c.user.CanRun(c.f, c.j); got != c.want {
			t.Errorf("case %d: CanRun = %v, want %v", i, got, c.want)
		}
	}
}

func TestUser_CanAdmin(t *testing.T) {
	if !(&User{Role: RoleAdmin}).CanAdmin() || (&User{Role: RoleRunner}).CanAdmin() || (*User)(nil).CanAdmin() {
		t.Fatal("CanAdmin")
	}
}

func TestUser_CanSee(t *testing.T) {
	scoped := &User{Name: "web", Role: RoleRunner, Folders: []string{"web"}}
	guest := &User{Name: "guest", Role: RoleViewer}
	if scoped.CanSee("infra") || !scoped.CanSee("web") || !guest.CanSee("infra") || (*User)(nil).CanSee("web") {
		t.Fatal("CanSee")
	}
}

func TestUser_Initials(t *testing.T) {
	for name, want := range map[string]string{
		"Admin Back":  "AB",
		"admin":       "AD",
		"tit.petric":  "TP",
		"ops@example": "OE",
		"x":           "X",
		"":            "?",
		"...":         "?",
	} {
		if got := (&User{Name: name}).Initials(); got != want {
			t.Errorf("%q: got %q, want %q", name, got, want)
		}
	}
	if (*User)(nil).Initials() != "?" {
		t.Fatal("nil user initials")
	}
}
