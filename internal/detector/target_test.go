package detector

import (
	"slices"
	"testing"
)

func TestTarget_Argv(t *testing.T) {
	named := &Target{Name: "build"}
	if !slices.Equal(named.Argv(), []string{"build"}) {
		t.Fatalf("name as argv: %v", named.Argv())
	}
	script := &Target{Name: "cleanup", Args: []string{"sh", "scripts/cleanup.sh"}}
	if !slices.Equal(script.Argv(), []string{"sh", "scripts/cleanup.sh"}) {
		t.Fatalf("args as argv: %v", script.Argv())
	}
}
