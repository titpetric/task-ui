package assert_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// assertExtra invokes the assertions in assert_extra.go on the path where
// they hold. TestAssert runs it.
func assertExtra(t *testing.T) {
	var typed *int

	assert.Nil(t, nil)
	assert.Nil(t, typed)
	assert.NotNil(t, &struct{}{})

	assert.Len(t, "abc", 3)
	assert.Len(t, []int{1, 2}, 2)
	assert.Len(t, map[string]int{"a": 1}, 1)

	assert.Empty(t, nil)
	assert.Empty(t, "")
	assert.Empty(t, []int{})
	assert.Empty(t, 0)
	assert.NotEmpty(t, "x")

	assert.Greater(t, 2, 1)
	assert.Greater(t, uint(2), uint(1))
	assert.Greater(t, 2.5, 1.5)
	assert.Greater(t, "b", "a")

	assert.NotContains(t, "haystack", "needle")
	assert.NotContains(t, []string{"a"}, "b")
	assert.NotContains(t, map[string]int{"a": 1}, "b")

	assert.ErrorIs(t, fmt.Errorf("open: %w", os.ErrNotExist), os.ErrNotExist)

	assert.IsIncreasing(t, []int{1, 2, 3})
	assert.IsIncreasing(t, []string{"a", "b"})
}

// errTarget is the sentinel ErrorIs is asked to match and does not find.
var errTarget = errors.New("target")

// assertExtraFailure drives the failing path of every assertion in
// assert_extra.go. TestAssertFailure runs it.
func assertExtraFailure(t *testing.T) {
	var typed *int

	mustFail(t, "Nil", func(tb testing.TB) { assert.Nil(tb, &struct{}{}) })
	mustFail(t, "NotNil", func(tb testing.TB) { assert.NotNil(tb, typed) })

	mustFail(t, "Len", func(tb testing.TB) { assert.Len(tb, "abc", 2) })
	mustFail(t, "LenNoLength", func(tb testing.TB) { assert.Len(tb, 42, 1) })

	mustFail(t, "Empty", func(tb testing.TB) { assert.Empty(tb, "x") })
	mustFail(t, "NotEmpty", func(tb testing.TB) { assert.NotEmpty(tb, "") })

	mustFail(t, "Greater", func(tb testing.TB) { assert.Greater(tb, 1, 2) })
	mustFail(t, "GreaterUint", func(tb testing.TB) { assert.Greater(tb, uint(1), uint(2)) })
	mustFail(t, "GreaterFloat", func(tb testing.TB) { assert.Greater(tb, 1.5, 2.5) })
	mustFail(t, "GreaterString", func(tb testing.TB) { assert.Greater(tb, "a", "b") })
	// Kinds that do not compare are a failure, not an equal.
	mustFail(t, "GreaterMixedKinds", func(tb testing.TB) { assert.Greater(tb, 1, "a") })

	mustFail(t, "NotContains", func(tb testing.TB) { assert.NotContains(tb, "haystack", "stack") })

	mustFail(t, "ErrorContainsNil", func(tb testing.TB) { assert.ErrorContains(tb, nil, "boom") })
	mustFail(t, "ErrorContains", func(tb testing.TB) { assert.ErrorContains(tb, errors.New("boom"), "bang") })
	mustFail(t, "ErrorIs", func(tb testing.TB) { assert.ErrorIs(tb, errors.New("boom"), errTarget) })

	mustFail(t, "IsIncreasing", func(tb testing.TB) { assert.IsIncreasing(tb, []int{1, 3, 2}) })
	mustFail(t, "IsIncreasingKind", func(tb testing.TB) { assert.IsIncreasing(tb, 42) })
}
