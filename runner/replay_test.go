package runner

import (
	"strconv"
	"strings"
	"testing"
)

// Twelve lines scroll off a four row screen; two do not.
func TestReplay(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"version":2,"width":20,"height":4}` + "\n")
	for i := 0; i < 12; i++ {
		b.WriteString(`[0.1,"o","line ` + strconv.Itoa(i) + `\r\n"]` + "\n")
	}
	replay, err := ReplayCast(strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	if replay.Complete() {
		t.Fatal("a screen that scrolled shows the whole output")
	}
	if replay.Lines() != 12 {
		t.Fatalf("lines: %d, want 12", replay.Lines())
	}
	short := `{"version":2,"width":20,"height":4}` + "\n" + `[0.1,"o","one\r\ntwo\r\n"]` + "\n"
	replay, _ = ReplayCast(strings.NewReader(short))
	if !replay.Complete() {
		t.Fatalf("two lines on a four row screen: %q vs %q", replay.Screen.Trim(1).Text(), replay.Output.Text())
	}
}
