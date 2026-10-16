package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/access"
)

func TestHash(t *testing.T) {
	var out bytes.Buffer
	if err := hash([]string{"secret"}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	hashed := strings.TrimSpace(out.String())
	if !access.CheckPassword(&access.User{Password: hashed}, "secret") {
		t.Fatalf("hash does not verify: %q", hashed)
	}

	out.Reset()
	if err := hash(nil, strings.NewReader("from stdin\n"), &out); err != nil {
		t.Fatal(err)
	}
	if !access.CheckPassword(&access.User{Password: strings.TrimSpace(out.String())}, "from stdin") {
		t.Fatal("stdin password does not verify")
	}

	if err := hash(nil, strings.NewReader(""), &out); err == nil {
		t.Fatal("empty input accepted")
	}
}

func TestAssets(t *testing.T) {
	t.Setenv("TASK_UI_ASSETS", "")
	if _, err := assets().Open("templates/_layout.tpl"); err != nil {
		t.Fatalf("embedded layout: %v", err)
	}
	t.Setenv("TASK_UI_ASSETS", ".")
	if _, err := assets().Open("templates/_layout.tpl"); err != nil {
		t.Fatalf("layout from disk: %v", err)
	}
}
