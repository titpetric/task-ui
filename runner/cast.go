package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"

	"github.com/hinshun/vt10x"
)

// OutputRows is how tall a terminal a recording's whole output is
// replayed into: the most of a run's output it shows.
const OutputRows = 5000

// Replay is a recording played back: the screen it ended on, and its
// whole output, the same output in a terminal OutputRows tall without
// its empty rows at the bottom.
type Replay struct {
	Screen Frame
	Output Frame
}

// Lines counts the lines of the whole output, less the empty one the
// cursor waits on after the last.
func (r Replay) Lines() int {
	lines := r.Output.Rows
	if lines > 0 && blank(r.Output.Lines[lines-1]) {
		lines--
	}
	return lines
}

// Complete reports whether the screen already shows the whole output:
// nothing scrolled off it.
func (r Replay) Complete() bool {
	return r.Screen.Trim(1).Text() == r.Output.Text()
}

// FrameFromCast replays an asciicast v2 recording into a terminal of the
// size its header names and returns the frame it ended on: the last
// screen of a run whose session is gone.
func FrameFromCast(r io.Reader) (Frame, error) {
	replay, err := ReplayCast(r)
	return replay.Screen, err
}

// OutputFromCast replays a recording and returns its whole output: every
// line, colours and redraws as they ended up.
func OutputFromCast(r io.Reader) (Frame, error) {
	replay, err := ReplayCast(r)
	return replay.Output, err
}

// ReplayCast plays a recording once into two terminals: one the size its
// header names, for the last screen, and one as wide and OutputRows tall,
// for the whole output.
func ReplayCast(r io.Reader) (Replay, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	if !scanner.Scan() {
		return Replay{}, errors.New("empty recording")
	}
	var header struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil || header.Width <= 0 || header.Height <= 0 {
		return Replay{}, errors.New("not an asciicast recording")
	}
	screen := vt10x.New(vt10x.WithSize(header.Width, header.Height))
	output := vt10x.New(vt10x.WithSize(header.Width, OutputRows))
	for scanner.Scan() {
		var event []any
		if json.Unmarshal(scanner.Bytes(), &event) != nil || len(event) < 3 {
			continue // a torn last line
		}
		if kind, _ := event[1].(string); kind != "o" {
			continue
		}
		if data, ok := event[2].(string); ok {
			_, _ = screen.Write([]byte(data))
			_, _ = output.Write([]byte(data))
		}
	}
	if err := scanner.Err(); err != nil {
		return Replay{}, err
	}
	return Replay{Screen: snapshot(screen), Output: snapshot(output).Trim(1)}, nil
}
