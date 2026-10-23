package retention_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// The defaults keep recordings 31 days and the rest for ever; stored
// values round trip, and a wrong one falls back to the default.
func TestSettings(t *testing.T) {
	t.Parallel()
	defaults := retention.Load(nil)
	assert.Equal(t, retention.Rule{Age: retention.Month}, defaults.Recordings)
	assert.Equal(t, retention.Rule{Age: retention.Never}, defaults.Artifacts)
	assert.Equal(t, retention.Rule{Age: retention.Never}, defaults.SecretReads)
	assert.Equal(t, retention.Rule{Age: retention.Never}, defaults.KeyReads)

	s := retention.Settings{Recordings: retention.Rule{Age: retention.Week, Limit: 1 << 30}, Artifacts: retention.Rule{Age: retention.Day}, SecretReads: retention.Rule{Age: retention.Never, Limit: 100}, KeyReads: retention.Rule{Age: retention.Week}}
	assert.True(t, s.Valid())
	assert.Equal(t, s.KeyReads, s.Rule(retention.KeyReads))
	assert.Equal(t, s, retention.Load(s.Values()))
	assert.Equal(t, s.Artifacts, s.Rule(retention.Artifacts))

	values := s.Values()
	values["retention.artifacts.age"] = "3d"
	assert.Equal(t, defaults.Artifacts, retention.Load(values).Artifacts)
	assert.False(t, retention.Settings{Recordings: retention.Rule{Age: "x"}}.Valid())
	assert.Equal(t, []string{"recordings", "artifacts", "secret_reads", "key_reads"}, retention.Kinds)
}
