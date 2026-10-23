package skills

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// Listing is what `atkins -l --json --all` answered for a directory:
// the sections it lists, and, for a pipeline its linter refuses, the
// error and the findings. A broken pipeline lists no sections unless
// atkins prints them after the findings.
type Listing struct {
	Sections []Section
	Error    string
	Lint     []Finding
}

// lintAnswer is the object atkins prints in place of the sections for
// a pipeline with lint findings.
type lintAnswer struct {
	Error string    `json:"error"`
	Lint  []Finding `json:"lint"`
}

// ParseListing reads atkins' answer: one or more JSON values, each the
// list of sections or a lint object, and whatever follows the last one
// it can read, which it ignores. An answer that starts with neither is
// an error.
func ParseListing(out []byte) (Listing, error) {
	var listing Listing
	dec := json.NewDecoder(bytes.NewReader(out))
	for read := 0; ; read++ {
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			if read > 0 {
				return listing, nil
			}
			if errors.Is(err, io.EOF) {
				return listing, errors.New("atkins printed nothing")
			}
			return listing, err
		}
		if err := listing.add(value); err != nil {
			if read > 0 {
				return listing, nil
			}
			return listing, err
		}
	}
}

// add takes one JSON value atkins printed into the listing.
func (l *Listing) add(value json.RawMessage) error {
	if trimmed := bytes.TrimSpace(value); len(trimmed) > 0 && trimmed[0] == '[' {
		var sections []Section
		if err := json.Unmarshal(value, &sections); err != nil {
			return err
		}
		l.Sections = append(l.Sections, sections...)
		return nil
	}
	var answer lintAnswer
	if err := json.Unmarshal(value, &answer); err != nil {
		return err
	}
	if answer.Error == "" && len(answer.Lint) == 0 {
		return errors.New("atkins answered neither sections nor lint")
	}
	if l.Error == "" {
		l.Error = answer.Error
	}
	l.Lint = append(l.Lint, answer.Lint...)
	return nil
}

// Broken reports whether atkins refused a pipeline of the directory.
func (l Listing) Broken() bool {
	return l.Error != "" || len(l.Lint) > 0
}
