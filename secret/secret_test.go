package secret

import (
	"testing"
)

func TestValidKey(t *testing.T) {
	if !ValidKey("TOKEN") || !ValidKey("_x1") {
		t.Fatal("valid key refused")
	}
	if ValidKey("1BAD") || ValidKey("a-b") || ValidKey("") {
		t.Fatal("invalid key accepted")
	}
}

func TestValidScope(t *testing.T) {
	if !ValidScope(ScopeAll) || !ValidScope("infra") {
		t.Fatal("valid scope refused")
	}
	if ValidScope("../etc") || ValidScope("") {
		t.Fatal("invalid scope accepted")
	}
}
