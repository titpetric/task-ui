package envfile

import (
	"encoding/json"
	"testing"
)

func TestFile(t *testing.T) {
	data, err := json.Marshal(File{Folder: "infra", Path: ".env", Content: "secret", Keys: []string{"A"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || len(data) == 0 {
		t.Fatal("no json")
	}
	var back File
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Content != "" || back.Path != ".env" {
		t.Fatalf("content must not leave through json: %+v", back)
	}
}
