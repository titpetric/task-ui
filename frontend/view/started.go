package view

import (
	"strconv"
)

// Started is a time as a script reads it: milliseconds since the epoch,
// and an empty string where there is none.
//
// It is what the running rows of a table count from. A script cannot
// read the time back out of what Duration wrote: "2m05s" parses as 2 and
// "1h02m" as 1.
func Started(v any) string {
	t, ok := when(v)
	if !ok {
		return ""
	}
	return strconv.FormatInt(t.UnixMilli(), 10)
}
