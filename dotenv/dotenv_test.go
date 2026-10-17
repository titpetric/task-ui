package dotenv

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	vars, err := Parse(strings.NewReader("# comment\n\nPORT=8080\nexport NAME=\"my app\"\nQUOTED='x y'\n1BAD=1\nnoequals\n"))
	if err != nil {
		t.Fatal(err)
	}
	if vars["PORT"] != "8080" || vars["NAME"] != "my app" || vars["QUOTED"] != "x y" || len(vars) != 3 {
		t.Fatalf("parsed: %v", vars)
	}
}

func TestWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "x.env")
	vars := map[string]string{"A": "plain", "B": "has space", "C": `say "hi"`, "D": "", "E": "a#b"}
	if err := Write(path, vars); err != nil {
		t.Fatal(err)
	}
	back, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range vars {
		if back[key] != want {
			t.Errorf("%s: got %q, want %q", key, back[key], want)
		}
	}
	if _, err := Read(filepath.Join(t.TempDir(), "missing.env")); err == nil {
		t.Fatal("missing file read")
	}
}

func TestFormat(t *testing.T) {
	got := Format(map[string]string{"B": "2", "A": "1"})
	if got != "A=1\nB=2\n" {
		t.Fatalf("format: %q", got)
	}
}

func TestQuote(t *testing.T) {
	for in, want := range map[string]string{
		"plain":    "plain",
		"":         `""`,
		"a b":      `"a b"`,
		`say "hi"`: `'say "hi"'`,
		`it's "x"`: `it's "x"`,
		"a#b":      `"a#b"`,
	} {
		if got := Quote(in); got != want {
			t.Errorf("Quote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestKeys(t *testing.T) {
	if got := Keys(map[string]string{"b": "", "a": ""}); !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("keys: %v", got)
	}
}
