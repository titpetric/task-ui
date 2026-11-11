package detector

import (
	"testing"
)

func TestSource_Target(t *testing.T) {
	source := &Source{Kind: "make", Targets: []*Target{{Name: "build"}, {Name: "test"}}}
	if source.Target("test") == nil || source.Target("test").Name != "test" {
		t.Fatal("target by name")
	}
	if source.Target("deploy") != nil {
		t.Fatal("missing target found")
	}
	var none *Source
	if none.Target("build") != nil {
		t.Fatal("nil source returned a target")
	}
}
