package runner_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

func TestBacklogState(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(runner.BacklogState{Parallel: 4, Auto: 80, Length: 80})
	assert.NoError(t, err)
	var fields map[string]int
	assert.NoError(t, json.Unmarshal(data, &fields))
	assert.Equal(t, 80, fields["backlog_length_auto"])
	assert.Equal(t, 0, fields["backlog_length"])
	assert.Equal(t, 80, fields["backlog_length_effective"])
	assert.Equal(t, 4, fields["parallel"])
}
