package tooling_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

func TestProber_Bin(t *testing.T) {
	var none *tooling.Prober
	assert.Equal(t, "atkins", none.Bin())
	assert.Equal(t, "atkins", (&tooling.Prober{}).Bin())
	assert.Equal(t, "/usr/local/bin/atkins", (&tooling.Prober{Command: "/usr/local/bin/atkins"}).Bin())
}

// fakeAtkins answers as an atkins that knows --all, or as one that
// refuses it, and keeps the arguments of every call.
type fakeAtkins struct {
	all   bool
	db    string
	calls [][]string
}

func (f *fakeAtkins) output(_ context.Context, _ string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	if slices.Contains(args, "--all") && !f.all {
		return nil, errors.New("unknown flag: --all")
	}
	samples := ""
	if slices.Contains(args, "--all") {
		samples = `, "samples": [{"at": "2026-10-23T09:00:00Z", "command": "go test", "tool": "go", "duration_ms": 10}]`
	}
	return []byte(`{"workspace": {"name": "workspace", "database": "` + f.db + `", "days": 31, "usage": [{"date": "2026-10-23", "tool": "go", "runs": 1, "total_ms": 10}]` + samples + `}}`), nil
}

func TestProber_ProbeAsksForTheSamples(t *testing.T) {
	t.Parallel()

	fake := &fakeAtkins{all: true}
	prober := &tooling.Prober{Output: fake.output, TTL: time.Hour}

	report := prober.Probe("/w")
	assert.NotNil(t, report)
	assert.Len(t, report.Workspace.Samples, 1)
	assert.Equal(t, [][]string{{"--stats", "--json", "--all"}}, fake.calls)
}

// An atkins that refuses --all gets the call again without it, and the
// panel draws the days it reports.
func TestProber_ProbeFallsBackWithoutAll(t *testing.T) {
	t.Parallel()

	fake := &fakeAtkins{}
	prober := &tooling.Prober{Output: fake.output, TTL: time.Hour}

	report := prober.Probe("/w")
	assert.NotNil(t, report)
	assert.Len(t, report.Workspace.Samples, 0)
	assert.Len(t, report.Workspace.Usage, 1)
	assert.Equal(t, [][]string{{"--stats", "--json", "--all"}, {"--stats", "--json"}}, fake.calls)
}

// A kept answer holds until a database it read from changes: a run
// that ends records there, and the next page shows it.
func TestProber_ProbeForgetsOnAWrite(t *testing.T) {
	t.Parallel()

	db := filepath.Join(t.TempDir(), "atkins.db")
	assert.NoError(t, os.WriteFile(db, nil, 0o600))
	past := time.Now().Add(-time.Minute)
	assert.NoError(t, os.Chtimes(db, past, past))

	fake := &fakeAtkins{all: true, db: db}
	prober := &tooling.Prober{Output: fake.output, TTL: time.Hour}

	prober.Probe("/w")
	prober.Probe("/w")
	assert.Len(t, fake.calls, 1)

	future := time.Now().Add(time.Minute)
	assert.NoError(t, os.Chtimes(db, future, future))
	prober.Probe("/w")
	assert.Len(t, fake.calls, 2)
}
