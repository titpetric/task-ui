package beszel_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// A pair reads an array of two numbers, and zero from anything else.
func TestPair_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	var p beszel.Pair
	assert.NoError(t, json.Unmarshal([]byte(`[1.5,2]`), &p))
	assert.Equal(t, beszel.Pair{1.5, 2}, p)
	var odd beszel.Pair
	assert.NoError(t, json.Unmarshal([]byte(`"x"`), &odd))
	assert.Equal(t, beszel.Pair{}, odd)
}
