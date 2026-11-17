package api

import (
	"testing"
)

func TestSSE(t *testing.T) {
	var _ = (*Handlers).sse
	if heartbeat <= 0 {
		t.Fatalf("heartbeat %v", heartbeat)
	}
}
