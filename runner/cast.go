package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

// OutputRows is the height of the terminal a recording's whole output
// replays into: the most of a run's output it shows.
const OutputRows = 5000

// FrameFromCast replays an asciicast v2 recording and returns the screen it ended on, a finished run's last frame.
func FrameFromCast(r io.Reader) (Frame, error) {
	replay, err := ReplayCast(r)
	return replay.Screen, err
}

// OutputFromCast replays a recording and returns its whole output, every line as colours and redraws left it.
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
	screen := newTerminal(header.Width, header.Height)
	output := newTerminal(header.Width, OutputRows)
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
