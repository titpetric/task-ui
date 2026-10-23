package runner

import (
	"strings"
	"unicode"
)

// TitleMax is how many runes of a terminal title a run keeps.
const TitleMax = 120

// oscMax bounds an OSC sequence the scanner collects: a longer one is
// no title, and the scanner drops it.
const oscMax = 4096

// The states of the scanner.
const (
	scanText = iota
	scanEsc
	scanOSC
	scanOSCEsc
)

// titleScanner follows a run's output for the titles its terminal sets:
// OSC 0 and 2, `ESC ] 0;title BEL`, or ended with ST, `ESC \`. A
// sequence may arrive split across writes. The terminal emulator keeps
// a title too, but drops an empty one, which here clears it.
type titleScanner struct {
	state int
	osc   []byte
	// long says the sequence went past oscMax: it ends unread.
	long  bool
	title string
}

// scan reads output and returns the title it leaves, and whether the
// title changed.
func (t *titleScanner) scan(p []byte) (string, bool) {
	before := t.title
	for _, c := range p {
		switch t.state {
		case scanText:
			if c == 0x1b {
				t.state = scanEsc
			}
		case scanEsc:
			switch c {
			case ']':
				t.state, t.osc, t.long = scanOSC, t.osc[:0], false
			case 0x1b:
			default:
				t.state = scanText
			}
		case scanOSC:
			switch c {
			case 0x07:
				t.end()
			case 0x1b:
				t.state = scanOSCEsc
			default:
				if len(t.osc) < oscMax {
					t.osc = append(t.osc, c)
				} else {
					t.long = true
				}
			}
		case scanOSCEsc:
			if c == '\\' {
				t.end()
				continue
			}
			// Any other escape abandons the sequence and starts anew.
			t.state = scanEsc
			if c == ']' {
				t.state, t.osc, t.long = scanOSC, t.osc[:0], false
			}
		}
	}
	return t.title, t.title != before
}

// end finishes an OSC sequence: 0 and 2 set the title.
func (t *titleScanner) end() {
	t.state = scanText
	if t.long {
		return
	}
	code, text, _ := strings.Cut(string(t.osc), ";")
	if code == "0" || code == "2" {
		t.title = CleanTitle(text)
	}
}

// CleanTitle makes a terminal title fit to show: no control characters,
// no space around it, at most TitleMax runes.
func CleanTitle(title string) string {
	title = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == unicode.ReplacementChar {
			return -1
		}
		return r
	}, strings.ToValidUTF8(title, ""))
	title = strings.TrimSpace(title)
	if runes := []rune(title); len(runes) > TitleMax {
		title = strings.TrimSpace(string(runes[:TitleMax]))
	}
	return title
}
