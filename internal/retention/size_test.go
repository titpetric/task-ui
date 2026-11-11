package retention_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// A size reads with or without a space, in any case; a kilobyte is 1024
// bytes, as the pages count.
func TestParseSize(t *testing.T) {
	t.Parallel()
	for text, want := range map[string]int64{
		"500MB": 500 << 20, "2GB": 2 << 30, "2 gb": 2 << 30, "1.5GiB": 3 << 29, "100 MB": 100 << 20,
		"1048576": 1 << 20, "10k": 10 << 10, "": 0, "0": 0, "off": 0,
	} {
		got, err := retention.ParseSize(text)
		assert.NoError(t, err, text)
		assert.Equal(t, want, got, text)
	}
	for _, text := range []string{"lots", "-1GB", "1XB"} {
		_, err := retention.ParseSize(text)
		assert.ErrorIs(t, err, retention.ErrSize, text)
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "512 B", retention.FormatSize(512))
	assert.Equal(t, "1.5 MB", retention.FormatSize(3<<19))
	assert.Equal(t, "10.0 GB", retention.FormatSize(10<<30))
}

// The estimate is the limit over the average size there is.
func TestEstimate(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 100, retention.Estimate(100<<20, 10, 10<<20))
	assert.Equal(t, 2048, retention.Estimate(1<<30, 4, 2<<20))
	assert.Equal(t, 0, retention.Estimate(1<<30, 0, 0))
	assert.Equal(t, 0, retention.Estimate(0, 4, 100))
}
