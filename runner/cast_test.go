package runner

import (
	"strconv"
	"strings"
	"testing"
)

// A recording replayed into a terminal ends on the frame the run ended on.
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

// The output of a run is its recording replayed into a tall terminal:
// every line, where the last screen holds the last ones only.
func TestOutputFromCast(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"version":2,"width":20,"height":4}` + "\n")
	for i := 0; i < 12; i++ {
		b.WriteString(`[0.1,"o","line ` + strconv.Itoa(i) + `\r\n"]` + "\n")
	}
	last, err := FrameFromCast(strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(last.Text(), "line 0\n") {
		t.Fatalf("the last screen holds the first line: %q", last.Text())
	}
	all, err := OutputFromCast(strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		if !strings.Contains(all.Text(), "line "+strconv.Itoa(i)) {
			t.Fatalf("line %d missing: %q", i, all.Text())
		}
	}
	if all.Rows != 13 {
		t.Fatalf("rows: %d, want the 12 lines and the cursor's", all.Rows)
	}

}
