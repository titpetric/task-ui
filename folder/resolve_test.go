package folder

import (
	"testing"
)

// fixed is a catalogue of folders by ID.
type fixed []*Folder

func (c fixed) List() []*Folder { return c }

func (c fixed) Find(id string) *Folder {
	for _, f := range c {
		if f.ID == id {
			return f
		}
	}
	return nil
}

func TestResolve(t *testing.T) {
	c := fixed{{ID: "infra"}, {ID: "services"}, {ID: "services/api"}, {ID: "services/api/edge"}}

	for path, want := range map[string][2]string{
		"infra":                      {"infra", ""},
		"/infra/":                    {"infra", ""},
		"infra/plan":                 {"infra", "plan"},
		"infra/env/.env":             {"infra", "env/.env"},
		"services/api":               {"services/api", ""},
		"services/api/deploy":        {"services/api", "deploy"},
		"services/api/edge":          {"services/api/edge", ""},
		"services/api/edge/env/.env": {"services/api/edge", "env/.env"},
		"services/web":               {"services", "web"},
	} {
		f, rest := Resolve(c, path)
		if f == nil || f.ID != want[0] || rest != want[1] {
			t.Errorf("Resolve(%q) = %v %q, want %v", path, f, rest, want)
		}
	}
	if f, _ := Resolve(c, "nowhere/plan"); f != nil {
		t.Fatalf("resolved %v", f)
	}
	if f, _ := Resolve(c, "infrastructure"); f != nil {
		t.Fatalf("a prefix is not a folder: %v", f)
	}
}
