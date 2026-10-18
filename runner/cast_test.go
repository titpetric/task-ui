package runner

import (
	"strings"
	"testing"
)

// A recording replayed into a terminal gives the frame it ended on.
func TestFrameFromCast(t *testing.T) {
	cast := `{"version":2,"width":20,"height":4}
[0.1,"o","hello"]
[0.2,"o","\r\nworld\u001b[1m!"]
[0.3,"i","typed, not shown"]
`
	frame, err := FrameFromCast(strings.NewReader(cast))
	if err != nil {
		t.Fatal(err)
	}
	if frame.Cols != 20 || frame.Rows != 4 {
		t.Fatalf("size: %dx%d", frame.Cols, frame.Rows)
	}
	text := frame.Text()
	if !strings.Contains(text, "hello") || !strings.Contains(text, "world!") {
		t.Fatalf("text: %q", text)
	}
	if strings.Contains(text, "typed") {
		t.Fatalf("input shown: %q", text)
	}

	if _, err := FrameFromCast(strings.NewReader("not a cast")); err == nil {
		t.Fatal("a bad header passed")
	}
}
