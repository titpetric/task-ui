package model

import (
	"strings"
)

// The times a check lets the workflow's steps run, the first counted. A
// step naming no tries takes CheckPasses, one naming fewer than
// LeastPasses takes LeastPasses, since a single pass leaves a check
// nothing to send round, and one naming more than MostPasses takes
// MostPasses, so steps that keep writing the marker end the run instead
// of looping on.
const (
	CheckPasses = 3
	MostPasses  = 20
	LeastPasses = 2
)

// Markers are the strings a check looks for, from its comma separated column.
func Markers(csv string) []string {
	var out []string
	for _, marker := range strings.Split(csv, ",") {
		if marker = strings.TrimSpace(marker); marker != "" {
			out = append(out, marker)
		}
	}
	return out
}

// Marked is the first of the markers the text holds, case ignored, and "" for none.
func Marked(markers []string, text string) string {
	text = strings.ToLower(text)
	for _, marker := range markers {
		if strings.Contains(text, strings.ToLower(marker)) {
			return marker
		}
	}
	return ""
}

// Passes is how many times a check lets the steps run, from its tries column.
func Passes(tries int64) int {
	switch {
	case tries <= 0:
		return CheckPasses
	case tries < LeastPasses:
		return LeastPasses
	case tries > MostPasses:
		return MostPasses
	}
	return int(tries)
}
