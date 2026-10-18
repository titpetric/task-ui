package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"

	"github.com/hinshun/vt10x"
)

// FrameFromCast replays an asciicast v2 recording into a terminal of the
// size its header names and returns the frame it ended on: the last
// screen of a run whose session is gone.
func FrameFromCast(r io.Reader) (Frame, error) {
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
	term := vt10x.New(vt10x.WithSize(header.Width, header.Height))
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
