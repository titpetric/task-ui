// Copyright 2022 Fortio Authors
// Licensed under the Apache License, Version 2.0.

package assert

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// Nil checks that value is nil, including a typed nil.
func Nil(t testing.TB, value any, msg ...any) {
	if !isNil(value) {
		Errorf(t, "expecting nil, got %v: %v", value, msg)
	}
}

// NotNil checks that value is not nil, including a typed nil.
func NotNil(t testing.TB, value any, msg ...any) {
	if isNil(value) {
		Errorf(t, "expecting non-nil: %v", msg)
	}
}

// Len checks the length of a string, array, channel, map, or slice.
func Len(t testing.TB, value any, expected int, msg ...any) {
	v := reflect.ValueOf(value)
	if !hasLength(v) {
		Errorf(t, "%T has no length: %v", value, msg)
		return
	}
	if actual := v.Len(); actual != expected {
		Errorf(t, "length %d unexpectedly not equal %d: %v", actual, expected, msg)
	}
}

// Empty checks that value is nil, zero, or has length zero.
func Empty(t testing.TB, value any, msg ...any) {
	if !isEmpty(value) {
		Errorf(t, "expecting empty, got %v: %v", value, msg)
	}
}

// NotEmpty checks that value is neither nil nor empty.
func NotEmpty(t testing.TB, value any, msg ...any) {
	if isEmpty(value) {
		Errorf(t, "expecting non-empty: %v", msg)
	}
}

// Greater checks that actual is greater than expected.
func Greater(t testing.TB, actual, expected any, msg ...any) {
	if cmp, ok := ordered(actual, expected); !ok || cmp <= 0 {
		Errorf(t, "%v is not greater than %v: %v", actual, expected, msg)
	}
}

// NotContains checks that needle is not in haystack.
func NotContains(t testing.TB, haystack, needle any, msg ...any) {
	if containsElement(haystack, needle) {
		Errorf(t, "%v unexpectedly contains %v: %v", haystack, needle, msg)
	}
}

// ErrorContains checks that err is non-nil and its message contains want.
func ErrorContains(t testing.TB, err error, want string, msg ...any) {
	if err == nil {
		Errorf(t, "expecting an error containing %q, got nil: %v", want, msg)
		return
	}
	if !strings.Contains(err.Error(), want) {
		Errorf(t, "%v does not contain %q: %v", err, want, msg)
	}
}

// ErrorIs checks that err matches target.
func ErrorIs(t testing.TB, err, target error, msg ...any) {
	if !errors.Is(err, target) {
		Errorf(t, "%v does not match %v: %v", err, target, msg)
	}
}

// IsIncreasing checks that every item is greater than the item before it.
func IsIncreasing(t testing.TB, value any, msg ...any) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		Errorf(t, "%T is not an array or slice: %v", value, msg)
		return
	}
	for i := 1; i < v.Len(); i++ {
		if cmp, ok := ordered(v.Index(i).Interface(), v.Index(i-1).Interface()); !ok || cmp <= 0 {
			Errorf(t, "%v is not increasing at index %d: %v", value, i, msg)
			return
		}
	}
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	if hasLength(v) {
		return v.Len() == 0
	}
	return v.IsZero()
}

func hasLength(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return true
	default:
		return false
	}
}

func containsElement(haystack, needle any) bool {
	v := reflect.ValueOf(haystack)
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.String:
		s, ok := needle.(string)
		return ok && strings.Contains(v.String(), s)
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(v.Index(i).Interface(), needle) {
				return true
			}
		}
	case reflect.Map:
		key := reflect.ValueOf(needle)
		return key.IsValid() && key.Type().AssignableTo(v.Type().Key()) && v.MapIndex(key).IsValid()
	}
	return false
}

// ordered compares two values of the same broad kind. The second return
// is false when the two are not comparable, which every caller has to
// treat as a failure: a bare 0 would read as "equal" and let
// LessOrEqual and GreaterOrEqual pass on an int against a string.
func ordered(actual, expected any) (int, bool) {
	a := reflect.ValueOf(actual)
	b := reflect.ValueOf(expected)
	if !a.IsValid() || !b.IsValid() {
		return 0, false
	}

	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if b.Kind() >= reflect.Int && b.Kind() <= reflect.Int64 {
			return compare(a.Int(), b.Int()), true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if b.Kind() >= reflect.Uint && b.Kind() <= reflect.Uintptr {
			return compare(a.Uint(), b.Uint()), true
		}
	case reflect.Float32, reflect.Float64:
		if b.Kind() == reflect.Float32 || b.Kind() == reflect.Float64 {
			return compare(a.Float(), b.Float()), true
		}
	case reflect.String:
		if b.Kind() == reflect.String {
			return strings.Compare(a.String(), b.String()), true
		}
	}
	return 0, false
}

func compare[T ~int64 | ~uint64 | ~float64](a, b T) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
