package form

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalid is what Bind returns for a submission a field refuses; the Submission names each.
var ErrInvalid = errors.New("the form needs a fix")

// Bind returns the submission of the visible fields, with ErrInvalid for a refused value and the submission either way.
func (f *Form) Bind(values map[string]string) (*Submission, error) {
	s := &Submission{Values: map[string]string{}, List: map[string][]string{}, Errors: map[string]string{}}
	var bad []string
	for _, field := range f.Visible(values) {
		value := strings.TrimSpace(values[field.Key])
		if value == "" {
			value = strings.TrimSpace(field.Default)
		}
		if message := field.bind(s, value); message != "" {
			s.Errors[field.Key] = message
			bad = append(bad, field.Key+": "+message)
		}
	}
	if len(bad) > 0 {
		return s, fmt.Errorf("%w: %s", ErrInvalid, strings.Join(bad, "; "))
	}
	return s, nil
}

// bind writes one field's value to the submission, returning the message of a value the field refuses.
func (f *Field) bind(s *Submission, value string) string {
	if f.Component == componentMulti {
		return f.bindMulti(s, value)
	}
	if value == "" {
		if f.Required {
			return "a value is needed"
		}
		if f.Component != componentBool {
			s.Values[f.Key] = ""
			return ""
		}
	}
	switch f.Component {
	case componentNumber:
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Sprintf("%q is no number", value)
		}
		value = strconv.FormatFloat(number, 'f', -1, 64)
	case componentBool:
		truth, ok := truthOf(value)
		if !ok {
			return fmt.Sprintf("%q is no yes or no", value)
		}
		value = strconv.FormatBool(truth)
	case componentChoice:
		if len(f.Options) > 0 && !f.hasOption(value) {
			return fmt.Sprintf("%q is no option of it", value)
		}
	}
	s.Values[f.Key] = value
	return ""
}

// bindMulti splits a multi on the line breaks and the commas, dropping what is empty.
func (f *Field) bindMulti(s *Submission, value string) string {
	var parts []string
	for _, part := range strings.FieldsFunc(value, func(r rune) bool { return r == '\n' || r == ',' }) {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 && f.Required {
		return "a value is needed"
	}
	if len(f.Options) > 0 {
		for _, part := range parts {
			if !f.hasOption(part) {
				return fmt.Sprintf("%q is no option of it", part)
			}
		}
	}
	s.List[f.Key] = parts
	s.Values[f.Key] = strings.Join(parts, ", ")
	return ""
}

// truthOf reads a checkbox value, false and false for one that is neither yes nor no.
func truthOf(value string) (bool, bool) {
	switch strings.ToLower(value) {
	case "1", "t", "y", "on", "yes", "true":
		return true, true
	case "", "0", "f", "n", "off", "no", "false":
		return false, true
	}
	return false, false
}
