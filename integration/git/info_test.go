package git

import (
	"encoding/json"
	"testing"
)

func TestInfo(t *testing.T) {
	data, err := json.Marshal(Info{Remote: "file:///r", Branch: "main", Commit: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	var back Info
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Remote != "file:///r" || back.Branch != "main" || back.Commit != "abc" {
		t.Fatalf("round trip: %+v", back)
	}
}
