package retention

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// units are the multiples ParseSize reads, as FormatSize writes them.
var units = []struct {
	suffix string
	bytes  int64
}{
	{"TB", 1 << 40}, {"GB", 1 << 30}, {"MB", 1 << 20}, {"KB", 1 << 10},
	{"TIB", 1 << 40}, {"GIB", 1 << 30}, {"MIB", 1 << 20}, {"KIB", 1 << 10},
	{"T", 1 << 40}, {"G", 1 << 30}, {"M", 1 << 20}, {"K", 1 << 10}, {"B", 1},
}

// ErrSize refuses a size ParseSize can't read.
var ErrSize = errors.New("a size is a number and a unit, such as 500MB or 2GB")

// ParseSize reads a size such as 500MB, 2 GB, 1.5GiB or 1048576, in
// bytes; a kilobyte is 1024 bytes, as FormatSize counts. Empty, 0 and
// off are no limit, 0.
func ParseSize(text string) (int64, error) {
	text = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(text), " ", ""))
	if text == "" || text == "0" || text == "OFF" {
		return 0, nil
	}
	multiple := int64(1)
	for _, unit := range units {
		if number, ok := strings.CutSuffix(text, unit.suffix); ok {
			text, multiple = number, unit.bytes
			break
		}
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil || value < 0 {
		return 0, ErrSize
	}
	return int64(value * float64(multiple)), nil
}

// FormatSize writes a byte count as the pages do: 512 B, 1.5 MB.
func FormatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, suffix := float64(n)/unit, "KB"
	for _, next := range []string{"MB", "GB", "TB"} {
		if value < unit {
			break
		}
		value, suffix = value/unit, next
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

// Estimate says about how many items a size holds, at the average size
// of the ones there are: 0 without any.
func Estimate(limit int64, count int, bytes int64) int {
	if count == 0 || bytes <= 0 || limit <= 0 {
		return 0
	}
	return int(float64(limit) / (float64(bytes) / float64(count)))
}
