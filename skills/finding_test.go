package skills_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// A finding reads the fields atkins names, and writes the empty ones
// but detail away.
func TestFinding(t *testing.T) {
	t.Parallel()
	var finding skills.Finding
	assert.NoError(t, json.Unmarshal([]byte(`{"pipeline":"p","path":"a.yml","job":"j","issue":"i","detail":"d"}`), &finding))
	assert.Equal(t, skills.Finding{Pipeline: "p", Path: "a.yml", Job: "j", Issue: "i", Detail: "d"}, finding)

	data, err := json.Marshal(skills.Finding{Detail: "d"})
	assert.NoError(t, err)
	assert.Equal(t, `{"detail":"d"}`, string(data))
}
