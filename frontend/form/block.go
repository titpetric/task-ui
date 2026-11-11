package form

import (
	"strings"
)

// The fence of a form block: three backticks at the least, and the word the info string opens on.
const (
	minTicks = 3
	infoWord = "form"
)

// Block is one form fence of a markdown body: where it is, the yaml in it and what Parse made of it.
type Block struct {
	// Start is the byte offset the fence opens at, its backticks included.
	Start int
	// End is the byte offset just past the closing backticks.
	End int
	// Body is the yaml between the fences, the opening line's indent taken off.
	Body string
	// Form is the parsed form, nil for one Parse refused.
	Form *Form
	// Err is what Parse refused the block for, so a renderer draws the error instead of the reply.
	Err error
}

// Blocks returns every form fence of a markdown body in order, each with Parse's result.
func Blocks(markdown string) []Block {
	var blocks []Block
	for pos := 0; pos < len(markdown); {
		line, next := lineAt(markdown, pos)
		open, ok := fenceOf(line)
		if !ok {
			pos = next
			continue
		}
		block, after := open.scan(markdown, pos, next)
		if open.form {
			block.Form, block.Err = Parse([]byte(block.Body))
			blocks = append(blocks, block)
		}
		pos = after
	}
	return blocks
}

// fence is a fence's opening line: how it is indented, how many backticks it took, whether it says form.
type fence struct {
	indent string
	ticks  int
	form   bool
}

// fenceOf reads a line as a fence opener, false for a line that opens none.
func fenceOf(line string) (fence, bool) {
	indent, ticks, rest := fenceParts(line)
	if ticks < minTicks || strings.Contains(rest, "`") {
		return fence{}, false
	}
	return fence{indent: indent, ticks: ticks, form: firstWord(rest) == infoWord}, true
}

// closes reports whether a line closes the fence: as many backticks, and nothing after them.
func (f fence) closes(line string) bool {
	_, ticks, rest := fenceParts(line)
	return ticks >= f.ticks && rest == ""
}

// scan returns the block the fence opens at start and the offset past its closing line, the body de-indented.
func (f fence) scan(markdown string, start, pos int) (Block, int) {
	var lines []string
	for pos < len(markdown) {
		line, next := lineAt(markdown, pos)
		if f.closes(line) {
			return Block{Start: start, End: pos + len(line), Body: body(lines)}, next
		}
		lines = append(lines, strings.TrimPrefix(line, f.indent))
		pos = next
	}
	return Block{Start: start, End: len(markdown), Body: body(lines)}, len(markdown)
}

// fenceParts splits a line into its indent, how many backticks follow it and the rest, trimmed.
func fenceParts(line string) (string, int, string) {
	ticked := strings.TrimLeft(line, " \t")
	rest := strings.TrimLeft(ticked, "`")
	return line[:len(line)-len(ticked)], len(ticked) - len(rest), strings.TrimSpace(rest)
}

// lineAt returns the line at an offset without its break, and the offset the next line opens at.
func lineAt(s string, pos int) (string, int) {
	if i := strings.IndexByte(s[pos:], '\n'); i >= 0 {
		return strings.TrimSuffix(s[pos:pos+i], "\r"), pos + i + 1
	}
	return strings.TrimSuffix(s[pos:], "\r"), len(s)
}

// firstWord returns the first word of a string, "" for none.
func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

// body returns a fence's lines as one document, "" for a fence that holds none.
func body(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
