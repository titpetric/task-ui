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
