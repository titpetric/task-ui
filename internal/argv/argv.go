package argv

import (
	"errors"
	"fmt"
	"strings"
)

// The limits of a list of arguments.
const (
	// MaxArgs is how many arguments a run takes.
	MaxArgs = 32
	// MaxBytes is how long one argument is.
	MaxBytes = 512
)

// ErrUnterminated reports a quote that does not close.
var ErrUnterminated = errors.New("a quote does not close")

// Split reads a line into arguments by the quoting rules of the package doc.
// The result passes Check.
func Split(line string) ([]string, error) {
	var (
		args    []string
		current strings.Builder
		started bool
		quote   rune
		escaped bool
	)
	for _, r := range line {
		switch {
		case escaped:
			if quote == '"' && r != '"' && r != '\\' {
				current.WriteRune('\\')
			}
			current.WriteRune(r)
			escaped = false
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\\':
			escaped, started = true, true
		case quote == '"':
			if r == '"' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, started = r, true
		case r == ' ' || r == '\t':
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
		default:
			current.WriteRune(r)
			started = true
		}
	}
	if quote != 0 || escaped {
		return nil, ErrUnterminated
	}
	if started {
		args = append(args, current.String())
	}
	if err := Check(args); err != nil {
		return nil, err
	}
	return args, nil
}

// Check holds a list of arguments to the limits: at most MaxArgs, each
// at most MaxBytes, none with a NUL or a line break.
func Check(args []string) error {
	if len(args) > MaxArgs {
		return fmt.Errorf("at most %d arguments", MaxArgs)
	}
	for _, arg := range args {
		if len(arg) > MaxBytes {
			return fmt.Errorf("an argument is at most %d bytes", MaxBytes)
		}
		if strings.ContainsAny(arg, "\x00\n\r") {
			return errors.New("an argument holds no NUL or line break")
		}
	}
	return nil
}
