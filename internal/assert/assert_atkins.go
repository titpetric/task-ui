package assert

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// AnError is a sentinel error for tests that need one and do not care
// which. It replaces testify's assert.AnError.
var AnError = errors.New("assert.AnError general error for testing")

// Positive checks that value is a number greater than zero.
func Positive(t testing.TB, value any, msg ...any) {
	if !isPositive(value) {
		Errorf(t, "%v is not positive: %v", value, msg)
	}
}

// Zero checks that value is the zero value for its type.
func Zero(t testing.TB, value any, msg ...any) {
	if !isZero(value) {
		Errorf(t, "%v is not the zero value: %v", value, msg)
	}
}

// NotZero checks that value is not the zero value for its type.
func NotZero(t testing.TB, value any, msg ...any) {
	if isZero(value) {
		Errorf(t, "expecting a non-zero value: %v", msg)
	}
}

// Less checks that actual is less than expected.
func Less(t testing.TB, actual, expected any, msg ...any) {
	if cmp, ok := ordered(actual, expected); !ok || cmp >= 0 {
		Errorf(t, "%v is not less than %v: %v", actual, expected, msg)
	}
}

// LessOrEqual checks that actual is less than or equal to expected.
func LessOrEqual(t testing.TB, actual, expected any, msg ...any) {
	if cmp, ok := ordered(actual, expected); !ok || cmp > 0 {
		Errorf(t, "%v is not less than or equal to %v: %v", actual, expected, msg)
	}
}

// GreaterOrEqual checks that actual is greater than or equal to expected.
func GreaterOrEqual(t testing.TB, actual, expected any, msg ...any) {
	if cmp, ok := ordered(actual, expected); !ok || cmp < 0 {
		Errorf(t, "%v is not greater than or equal to %v: %v", actual, expected, msg)
	}
}

// FileExists checks that path exists and is a regular file.
func FileExists(t testing.TB, path string, msg ...any) {
	info, err := os.Lstat(path)
	switch {
	case err != nil:
		Errorf(t, "%q does not exist: %v", path, msg)
	case info.IsDir():
		Errorf(t, "%q is a directory, not a file: %v", path, msg)
	}
}

// NoFileExists checks that path is not a regular file. A directory at
// path passes, matching testify.
func NoFileExists(t testing.TB, path string, msg ...any) {
	info, err := os.Lstat(path)
	if err == nil && !info.IsDir() {
		Errorf(t, "%q unexpectedly exists: %v", path, msg)
	}
}

// DirExists checks that path exists and is a directory.
func DirExists(t testing.TB, path string, msg ...any) {
	info, err := os.Lstat(path)
	switch {
	case err != nil:
		Errorf(t, "%q does not exist: %v", path, msg)
	case !info.IsDir():
		Errorf(t, "%q is a file, not a directory: %v", path, msg)
	}
}

// NoDirExists checks that path is not a directory. A regular file at
// path passes, matching testify.
func NoDirExists(t testing.TB, path string, msg ...any) {
	info, err := os.Lstat(path)
	if err == nil && info.IsDir() {
		Errorf(t, "%q unexpectedly exists: %v", path, msg)
	}
}

// JSONEq checks that two JSON documents carry the same value, ignoring
// key order and whitespace.
func JSONEq(t testing.TB, expected, actual string, msg ...any) {
	var want, got any
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		Errorf(t, "expected value is not valid JSON: %v: %v", err, msg)
		return
	}
	if err := json.Unmarshal([]byte(actual), &got); err != nil {
		Errorf(t, "actual value is not valid JSON: %v: %v", err, msg)
		return
	}
	if !reflect.DeepEqual(want, got) {
		Errorf(t, "%v unexpectedly not equal %v: %v", actual, expected, msg)
	}
}

// ErrorAs checks that err, or an error it wraps, is assignable to
// target, which must be a non-nil pointer.
func ErrorAs(t testing.TB, err error, target any, msg ...any) {
	if !errors.As(err, target) {
		Errorf(t, "%v does not match %T: %v", err, target, msg)
	}
}

// ElementsMatch checks that two arrays or slices hold the same elements
// with the same multiplicity, in any order.
func ElementsMatch(t testing.TB, listA, listB any, msg ...any) {
	extraA, extraB, ok := diffElements(listA, listB)
	switch {
	case !ok:
		Errorf(t, "%T and %T are not both arrays or slices: %v", listA, listB, msg)
	case len(extraA) > 0 || len(extraB) > 0:
		Errorf(t, "elements differ, only in %v: %v, only in %v: %v: %v", listA, extraA, listB, extraB, msg)
	}
}

// Eventually polls condition every tick until it returns true or
// waitFor elapses.
func Eventually(t testing.TB, condition func() bool, waitFor, tick time.Duration, msg ...any) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	deadline := time.After(waitFor)
	for {
		if condition() {
			return
		}
		select {
		case <-deadline:
			// One last read: the condition may have turned true
			// between the final tick and the deadline.
			if !condition() {
				Errorf(t, "condition was not met within %v: %v", waitFor, msg)
			}
			return
		case <-ticker.C:
		}
	}
}

func isPositive(value any) bool {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() > 0
	case reflect.Float32, reflect.Float64:
		return v.Float() > 0
	default:
		return false
	}
}

func isZero(value any) bool {
	if value == nil {
		return true
	}
	return reflect.ValueOf(value).IsZero()
}

// diffElements reports the elements of listA missing from listB and the
// other way around. The final return is false when either value is not
// an array or a slice.
func diffElements(listA, listB any) (extraA, extraB []any, ok bool) {
	a := reflect.ValueOf(listA)
	b := reflect.ValueOf(listB)
	if !isList(a) || !isList(b) {
		return nil, nil, false
	}

	matched := make([]bool, b.Len())
	for i := 0; i < a.Len(); i++ {
		element := a.Index(i).Interface()
		found := false
		for j := 0; j < b.Len(); j++ {
			if matched[j] || !ObjectsAreEqualValues(element, b.Index(j).Interface()) {
				continue
			}
			matched[j] = true
			found = true
			break
		}
		if !found {
			extraA = append(extraA, element)
		}
	}
	for j := 0; j < b.Len(); j++ {
		if !matched[j] {
			extraB = append(extraB, b.Index(j).Interface())
		}
	}
	return extraA, extraB, true
}

func isList(v reflect.Value) bool {
	return v.IsValid() && (v.Kind() == reflect.Array || v.Kind() == reflect.Slice)
}
