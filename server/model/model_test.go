package model

import "testing"

func TestUserCanRun(t *testing.T) {
	open := &Project{ID: "web"}
	locked := &Project{ID: "infra", Allow: []string{"ops"}}
	build := &Runbook{Name: "build"}
	certs := &Runbook{Name: "certs", Allow: []string{"admin"}}

	admin := &User{Name: "root", Role: RoleAdmin}
	ops := &User{Name: "ops", Role: RoleRunner}
	dev := &User{Name: "dev", Role: RoleRunner}
	scoped := &User{Name: "web", Role: RoleRunner, Projects: []string{"web"}}
	guest := &User{Name: "guest", Role: RoleViewer}

	cases := []struct {
		user *User
		p    *Project
		r    *Runbook
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
		if got := c.user.CanRun(c.p, c.r); got != c.want {
			t.Errorf("case %d: CanRun = %v, want %v", i, got, c.want)
		}
	}

	if !admin.CanAdmin() || ops.CanAdmin() || (*User)(nil).CanAdmin() {
		t.Error("CanAdmin")
	}
	if scoped.CanSee("infra") || !scoped.CanSee("web") || !guest.CanSee("infra") {
		t.Error("CanSee")
	}
}

func TestProjectRunbookNil(t *testing.T) {
	var p *Project
	if p.Runbook("x") != nil {
		t.Fatal("nil project returned a runbook")
	}
}
