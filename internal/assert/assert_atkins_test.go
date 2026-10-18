package assert_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
)

// wrappedError is the type ErrorAs is asked to find.
type wrappedError struct{ msg string }

func (e *wrappedError) Error() string { return e.msg }

// assertAtkins invokes the assertions in assert_atkins.go on the path
// where they hold. TestAssert runs it.
func assertAtkins(t *testing.T) {
	file, dir := fixturePaths(t)

	assert.Positive(t, 1)
	assert.Positive(t, uint(1))
	assert.Positive(t, 0.5)
	assert.Positive(t, time.Second)

	assert.Zero(t, 0)
	assert.Zero(t, "")
	assert.Zero(t, nil)
	assert.NotZero(t, 1)

	assert.Less(t, 1, 2)
	assert.LessOrEqual(t, 2, 2)
	assert.GreaterOrEqual(t, 2, 2)
	assert.LessOrEqual(t, 1*time.Second, 2*time.Second)

	assert.FileExists(t, file)
	assert.NoFileExists(t, filepath.Join(dir, "absent"))
	// A directory is not a file, so NoFileExists holds on one.
	assert.NoFileExists(t, dir)
	assert.DirExists(t, dir)
	assert.NoDirExists(t, filepath.Join(dir, "absent"))
	// And a file is not a directory.
	assert.NoDirExists(t, file)

	assert.JSONEq(t, `{"a":1,"b":2}`, `{"b":2,  "a":1}`)

	var target *wrappedError
	assert.ErrorAs(t, errors.Join(assert.AnError, &wrappedError{"boom"}), &target)
	assert.Equal(t, "boom", target.Error())

	assert.ElementsMatch(t, []int{1, 2, 2}, []int{2, 1, 2})
	assert.ElementsMatch(t, []string(nil), []string{})

	// The condition is already true, so Eventually returns on the first
	// read without waiting for a tick.
	assert.Eventually(t, func() bool { return true }, time.Second, time.Millisecond)

	// And here it turns true on a later read, which is the polling path.
	reads := 0
	assert.Eventually(t, func() bool { reads++; return reads > 2 }, time.Second, time.Millisecond)
}

// assertAtkinsFailure drives the failing path of every assertion in
// assert_atkins.go. TestAssertFailure runs it.
func assertAtkinsFailure(t *testing.T) {
	file, dir := fixturePaths(t)
	absent := filepath.Join(dir, "absent")

	mustFail(t, "Positive", func(tb testing.TB) { assert.Positive(tb, 0) })
	mustFail(t, "PositiveUint", func(tb testing.TB) { assert.Positive(tb, uint(0)) })
	mustFail(t, "PositiveFloat", func(tb testing.TB) { assert.Positive(tb, -0.5) })
	// A kind that is not a number is never positive.
	mustFail(t, "PositiveKind", func(tb testing.TB) { assert.Positive(tb, "1") })
	mustFail(t, "PositiveNil", func(tb testing.TB) { assert.Positive(tb, nil) })

	mustFail(t, "Zero", func(tb testing.TB) { assert.Zero(tb, 1) })
	mustFail(t, "NotZero", func(tb testing.TB) { assert.NotZero(tb, 0) })

	mustFail(t, "Less", func(tb testing.TB) { assert.Less(tb, 2, 1) })
	mustFail(t, "LessOrEqual", func(tb testing.TB) { assert.LessOrEqual(tb, 3, 2) })
	mustFail(t, "GreaterOrEqual", func(tb testing.TB) { assert.GreaterOrEqual(tb, 1, 2) })
	// Kinds that do not compare are a failure rather than a silent pass.
	mustFail(t, "LessOrEqualMixedKinds", func(tb testing.TB) { assert.LessOrEqual(tb, 1, "a") })
	mustFail(t, "GreaterOrEqualMixedKinds", func(tb testing.TB) { assert.GreaterOrEqual(tb, 1, "a") })

	mustFail(t, "FileExists", func(tb testing.TB) { assert.FileExists(tb, absent) })
	mustFail(t, "FileExistsDir", func(tb testing.TB) { assert.FileExists(tb, dir) })
	mustFail(t, "NoFileExists", func(tb testing.TB) { assert.NoFileExists(tb, file) })
	mustFail(t, "DirExists", func(tb testing.TB) { assert.DirExists(tb, absent) })
	mustFail(t, "DirExistsFile", func(tb testing.TB) { assert.DirExists(tb, file) })
	mustFail(t, "NoDirExists", func(tb testing.TB) { assert.NoDirExists(tb, dir) })

	mustFail(t, "JSONEq", func(tb testing.TB) { assert.JSONEq(tb, `{"a":1}`, `{"a":2}`) })
	mustFail(t, "JSONEqExpected", func(tb testing.TB) { assert.JSONEq(tb, `{`, `{}`) })
	mustFail(t, "JSONEqActual", func(tb testing.TB) { assert.JSONEq(tb, `{}`, `{`) })

	var target *wrappedError
	mustFail(t, "ErrorAs", func(tb testing.TB) { assert.ErrorAs(tb, assert.AnError, &target) })

	mustFail(t, "ElementsMatch", func(tb testing.TB) { assert.ElementsMatch(tb, []int{1, 2}, []int{1, 3}) })
	mustFail(t, "ElementsMatchCount", func(tb testing.TB) { assert.ElementsMatch(tb, []int{1, 1}, []int{1}) })
	mustFail(t, "ElementsMatchKind", func(tb testing.TB) { assert.ElementsMatch(tb, 1, []int{1}) })

	mustFail(t, "Eventually", func(tb testing.TB) {
		assert.Eventually(tb, func() bool { return false }, 5*time.Millisecond, time.Millisecond)
	})
}

// fixturePaths returns a regular file and a directory that both exist,
// so the four existence assertions have something to read.
func fixturePaths(t *testing.T) (file, dir string) {
	t.Helper()

	dir = t.TempDir()
	file = filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("writing the fixture file: %v", err)
	}
	return file, dir
}
