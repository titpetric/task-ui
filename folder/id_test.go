package folder

import (
	"testing"
)

func TestSafeID(t *testing.T) {
	for in, want := range map[string]string{
		"web site": "web-site",
		"infra":    "infra",
		"../x":     "x",
		"":         "default",
		"a/b c":    "a-b-c",
	} {
		if got := SafeID(in); got != want {
			t.Errorf("SafeID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidID(t *testing.T) {
	if !ValidID("infra") || !ValidID("web-site.v2") {
		t.Fatal("valid ids refused")
	}
	if ValidID("") || ValidID("web site") || ValidID("../x") {
		t.Fatal("invalid ids accepted")
	}
}

func TestValidID_Paths(t *testing.T) {
	for id, want := range map[string]bool{"infra": true, "services/api": true, "a/b/c": true, "a//b": false, "/a": false, "a/": false, "..": false, "a/../b": false, "": false} {
		if ValidID(id) != want {
			t.Errorf("ValidID(%q) = %v", id, !want)
		}
	}
	for rel, want := range map[string]string{"services/api": "services/api", "./a/b c": "a/b-c", "": "default", ".": "default"} {
		if got := PathID(rel); got != want {
			t.Errorf("PathID(%q) = %q, want %q", rel, got, want)
		}
	}
}
