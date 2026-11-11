package phpfront_test

import (
	"context"

	"github.com/titpetric/task-ui/phpfront"
)

// What the fixture answers of the live runs: the runner's sessions, the
// permission file and the dispatch, grouped the way host_run.go declares the
// shapes. What a call recorded is read off the fixture by the test that made
// it, which is why published and dispatched are on the struct.

func (h *host) Running(context.Context) []phpfront.Running {
	return []phpfront.Running{{Key: "infra/build", ID: "01hz", State: "running"}}
}

func (h *host) Flows(context.Context) []phpfront.Flow {
	return []phpfront.Flow{{WorkflowID: "01hw", RunID: "01hr"}}
}

func (h *host) Rules(context.Context) phpfront.Rules {
	return phpfront.Rules{
		Rules:       []phpfront.Rule{{Key: "view", Title: "Allow View"}},
		Sections:    []phpfront.Rule{{Key: "note", Title: "Notes"}},
		Permissions: []phpfront.Rule{{Key: "run", Title: "Run Jobs"}},
		Paths:       []string{"note/all", "note/view", "run"},
		Checks:      []string{"note/view", "run"},
		Extras:      []phpfront.Rule{{Key: "note/pin", Title: "Pin"}},
	}
}

func (h *host) Publish(_ context.Context, topic string, _ map[string]any) error {
	h.published = append(h.published, topic)
	return nil
}

func (h *host) Session(_ context.Context, id string) phpfront.Session {
	if id != "01hz" {
		return phpfront.Session{}
	}
	return phpfront.Session{
		Live:     true,
		Attached: 2,
		Message:  "waits for a free slot, number 1 in line",
		State:    "running",
		Title:    "build the image",
	}
}

func (h *host) Frame(_ context.Context, id string) string {
	if id != "01hz" {
		return ""
	}
	return "<svg></svg>"
}

func (h *host) Screen(_ context.Context, id, castPath string, cols int) phpfront.Screen {
	if castPath != "run.cast" {
		return phpfront.Screen{}
	}
	return phpfront.Screen{SVG: "<svg></svg>", Lines: 42, Complete: true}
}

func (h *host) Expired(_ context.Context, castPath, state string) bool {
	return castPath == "gone.cast" && state == "ok"
}

func (h *host) Claude(context.Context) bool { return true }

func (h *host) Dispatch(_ context.Context, comment, action, folder string) (string, error) {
	h.dispatched = append(h.dispatched, action+":"+comment+":"+folder)
	return "01NOTE", nil
}

func (h *host) RunOf(_ context.Context, note string) string {
	if note == "" {
		return ""
	}
	return "01RUN-" + note
}
