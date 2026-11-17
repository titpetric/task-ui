package api

import (
	"encoding/json"
	"testing"
)

func TestThreadRequest(t *testing.T) {
	req := ThreadRequest{To: "bob", ParentID: "p1", Body: "hello", Spoken: "hello", Audio: "voice/a.wav"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var back ThreadRequest
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back != req {
		t.Fatalf("round trip: got %+v", back)
	}
}
