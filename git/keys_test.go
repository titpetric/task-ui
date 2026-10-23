package git

import (
	"testing"
)

// The test's keys are Keys, as the key folder is.
func TestKeys(t *testing.T) {
	t.Parallel()
	var keys Keys = &fakeKeys{t: t}
	if keys.KnownHosts() != "/keys/known_hosts" {
		t.Fatal(keys.KnownHosts())
	}
}
