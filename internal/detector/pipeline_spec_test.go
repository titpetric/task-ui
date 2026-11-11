package detector

import (
	"testing"
)

func TestStepSpec_Label(t *testing.T) {
	cases := []struct {
		step stepSpec
		want string
	}{
		{stepSpec{Desc: "Verify", Run: "./verify.sh"}, "Verify"},
		{stepSpec{Name: "verify", Run: "./verify.sh"}, "verify"},
		{stepSpec{Run: "./verify.sh"}, "run: ./verify.sh"},
		{stepSpec{Run: "./verify.sh", bare: true}, "./verify.sh"},
		{stepSpec{Run: "a\nb"}, "run: <script>"},
		{stepSpec{Cmd: "go build"}, "cmd: go build"},
		{stepSpec{Cmds: []string{"a", "b"}}, "cmds: <2 commands>"},
		{stepSpec{Task: "notify"}, "task: notify"},
		{stepSpec{Run: "rm -f lock", Deferred: true}, "defer: rm -f lock"},
		{stepSpec{}, ""},
	}
	for _, c := range cases {
		if got := c.step.Label(); got != c.want {
			t.Errorf("%+v: got %q, want %q", c.step, got, c.want)
		}
	}
	if firstOf("", "b", "c") != "b" || firstOf() != "" {
		t.Fatal("firstOf")
	}
}

func TestDefaultCommand(t *testing.T) {
	if got := defaultCommand("atkins", "x.yml"); got[0] != "atkins" || got[2] != "x.yml" {
		t.Fatalf("atkins: %v", got)
	}
	if got := defaultCommand("task", "Taskfile.yml"); got[0] != "task" || got[1] != "-t" {
		t.Fatalf("task: %v", got)
	}
}
