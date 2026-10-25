package retention_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// The defaults keep every kind 31 days with no size limit; stored
// values round trip, a wrong one falls back to the default, and a kind
// the settings lack reads the default.
func TestSettings(t *testing.T) {
	t.Parallel()
	defaults := retention.Load(nil)
	assert.Len(t, defaults, len(retention.Kinds))
	for _, kind := range retention.Kinds {
		assert.Equal(t, retention.Rule{Age: retention.Month}, defaults.Rule(kind), kind)
	}
	assert.Equal(t, []string{"recordings", "runs", "prompts", "workflow_runs", "artifacts", "secret_reads", "tooling"}, retention.Kinds)

	s := retention.Settings{retention.Recordings: retention.Rule{Age: retention.Week, Limit: 1 << 30}, retention.Runs: retention.Rule{Age: retention.Year}, retention.Prompts: retention.Rule{Age: retention.Quarter},
		retention.WorkflowRuns: retention.Rule{Age: retention.Never}, retention.Artifacts: retention.Rule{Age: retention.Day}, retention.SecretReads: retention.Rule{Age: retention.Never}, retention.Tooling: retention.Rule{Age: retention.Month}}
	assert.True(t, s.Valid())
	assert.Equal(t, s, retention.Load(s.Values()))
	assert.Equal(t, retention.Rule{Age: retention.Year}, s.Rule(retention.Runs))
	assert.Equal(t, retention.Rule{Age: retention.Month}, retention.Settings{}.Rule(retention.Runs), "a kind the settings lack keeps the default")

	values := s.Values()
	values["retention.artifacts.age"] = "3d"
	assert.Equal(t, defaults.Rule(retention.Artifacts), retention.Load(values).Rule(retention.Artifacts))
	assert.False(t, retention.Settings{retention.Recordings: retention.Rule{Age: "x"}}.Valid())
}
