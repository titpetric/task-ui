package model

import (
	"encoding/json"
	"sort"
	"strings"
)

// A few columns hold a list or a map as JSON text: the folders a user
// sees, the allow lists of a project, its command overrides. These read
// and write them.

// DecodeList reads a JSON list of strings. Empty text is an empty list.
func DecodeList(text string) []string {
	var list []string
	if text == "" {
		return list
	}
	_ = json.Unmarshal([]byte(text), &list)
	return list
}

// EncodeList writes a list of strings as JSON text, nil as an empty list so the column never holds null.
func EncodeList(list []string) string {
	if list == nil {
		list = []string{}
	}
	data, _ := json.Marshal(list)
	return string(data)
}

// DecodeMap reads a JSON object of string lists, such as the allow lists
// of the jobs of a project. Empty text is an empty map.
func DecodeMap(text string) map[string][]string {
	result := map[string][]string{}
	if text == "" {
		return result
	}
	_ = json.Unmarshal([]byte(text), &result)
	return result
}

// EncodeMap writes a map of string lists as JSON text, keys sorted so the text is stable.
func EncodeMap(m map[string][]string) string {
	if m == nil {
		m = map[string][]string{}
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// DecodeStrings reads a JSON object of strings, such as the command
// overrides of a project.
func DecodeStrings(text string) map[string]string {
	result := map[string]string{}
	if text == "" {
		return result
	}
	_ = json.Unmarshal([]byte(text), &result)
	return result
}

// EncodeStrings writes a map of strings as JSON text.
func EncodeStrings(m map[string]string) string {
	if m == nil {
		m = map[string]string{}
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// SortedKeys returns the keys of a map, sorted.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Lines splits a textarea's value into the entries of a list setting,
// dropping the blank lines and the space around each entry.
//
// A blank line is not an entry. An empty string in a list of paths to
// ignore would match every path, and in a list of networks it would
// match no address at all.
func Lines(value string) []string {
	var out []string
	for _, line := range strings.Split(value, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
