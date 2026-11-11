package api_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// The form posts the shortest debug span spelled as a duration, which is how
// the stored row and OIDA_DEBUG_SPAN_MIN_DURATION spell it too. A field left
// empty is the zero duration, which keeps every span the interpreter marked as
// detail.
func TestTelemetryRequestDebugSpanMinDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		body   string
		want   time.Duration
		failed bool
	}{
		{body: `{"debug_span_min_duration":"1ms"}`, want: time.Millisecond},
		{body: `{"debug_span_min_duration":"500us"}`, want: 500 * time.Microsecond},
		{body: `{"debug_span_min_duration":" 2s "}`, want: 2 * time.Second},
		{body: `{"debug_span_min_duration":""}`, want: 0},
		{body: `{"enabled":true}`, want: time.Millisecond},
		{body: `{"debug_span_min_duration":"1000"}`, failed: true},
		{body: `{"debug_span_min_duration":"a while"}`, failed: true},
	}
	for _, test := range tests {
		t.Run(test.body, func(t *testing.T) {
			t.Parallel()
			var req api.TelemetryRequest
			assert.NoError(t, json.Unmarshal([]byte(test.body), &req))

			opts := oida.NewOptions("task-ui")
			opts.DebugSpanMinDuration = time.Millisecond
			err := req.Apply(&opts)
			if test.failed {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "debug_span_min_duration")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, opts.DebugSpanMinDuration)
		})
	}
}

// The answer spells the duration back, so the form it fills is the form it
// posts.
func TestTelemetryViewDebugSpanMinDuration(t *testing.T) {
	t.Parallel()
	opts := oida.NewOptions("task-ui")
	opts.DebugSpanMinDuration = 1500 * time.Microsecond
	assert.Equal(t, "1.5ms", api.NewTelemetryView(opts).DebugSpanMinDuration)
}
