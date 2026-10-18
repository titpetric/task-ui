package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"

	"github.com/hinshun/vt10x"
)

// OutputRows is how tall a terminal OutputFromCast replays into: the
// most of a run's output it shows.
const OutputRows = 5000

// FrameFromCast replays an asciicast v2 recording into a terminal of the
// size its header names and returns the frame it ended on: the last
// screen of a run whose session is gone.
func FrameFromCast(r io.Reader) (Frame, error) {
	return replay(r, 0)
}

// OutputFromCast replays a recording into a terminal as wide as the
// run's and OutputRows tall, and returns it without its empty rows at the
// bottom: the run's whole output, colours and redraws as they ended up.
func OutputFromCast(r io.Reader) (Frame, error) {
	frame, err := replay(r, OutputRows)
	if err != nil {
		return frame, err
	}
	return frame.Trim(1), nil
}

// replay plays a recording into a terminal, rows tall when rows is not
// zero, else as tall as the recording's header says.
func replay(r io.Reader, rows int) (Frame, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	if !scanner.Scan() {
		return Frame{}, errors.New("empty recording")
	}
	var header struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil || header.Width <= 0 || header.Height <= 0 {
		return Frame{}, errors.New("not an asciicast recording")
	}
	if rows == 0 {
		rows = header.Height
	}
	term := vt10x.New(vt10x.WithSize(header.Width, rows))
	for scanner.Scan() {
		var event []any
		if json.Unmarshal(scanner.Bytes(), &event) != nil || len(event) < 3 {
			continue // a torn last line
		}
		if kind, _ := event[1].(string); kind != "o" {
			continue
		}
		if data, ok := event[2].(string); ok {
			_, _ = term.Write([]byte(data))
		}
	}
	if err := scanner.Err(); err != nil {
		return Frame{}, err
	}
	return snapshot(term), nil
}
