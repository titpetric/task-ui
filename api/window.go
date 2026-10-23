package api

import (
	"encoding/json"
	"time"
)

// Window is what a window on a run may do to it, and what it did: the
// attach socket passes it every control message the window sends.
type Window struct {
	Terminal RunTerminal
	// Interactive says the job takes input: a window that may run the
	// job types into it.
	Interactive bool
	// ReadOnly windows watch: their input, stop and kill go nowhere.
	ReadOnly bool
	// Taker names the user who may take input on a job that takes
	// none, with the input permission; "" when the user may not.
	Taker string
	// taking says the window holds input: take, and no release since.
	taking bool
}

// Apply applies a control message of the window, a JSON text frame:
// input, resize, take, release, stop or kill. It returns the frame to
// answer with, nil for none. Input to a job that takes none goes to
// the process only while the window holds it.
func (w *Window) Apply(frame []byte) []byte {
	var msg control
	if json.Unmarshal(frame, &msg) != nil {
		return nil
	}
	switch msg.Type {
	case "resize":
		w.Terminal.Resize(msg.Cols, msg.Rows)
	case "input":
		if w.taking || (w.Interactive && !w.ReadOnly) {
			_ = w.Terminal.Input([]byte(msg.Message))
		}
	case "take":
		if w.Taker == "" {
			return controlFrame(control{Type: "input", State: "off", Message: "you need the input permission to type into this run"})
		}
		w.Terminal.TakeInput(w.Taker)
		w.taking = true
		return controlFrame(control{Type: "input", State: "on"})
	case "release":
		w.taking = false
		return controlFrame(control{Type: "input", State: "off"})
	case "stop":
		if !w.ReadOnly {
			w.Terminal.Stop(5 * time.Second)
		}
	case "kill":
		if !w.ReadOnly {
			w.Terminal.Kill(5 * time.Second)
		}
	}
	return nil
}

// controlFrame encodes a control message as a text frame.
func controlFrame(msg control) []byte {
	data, _ := json.Marshal(msg)
	return data
}
